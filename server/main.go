package main

import (
	"context"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-co-op/gocron/v2"
	"github.com/hashicorp/vault-client-go"
	"github.com/maypok86/otter"
	"log"
	"net/http"
	"os"
	"secretpaths/backend"
	"secretpaths/models"
	"strconv"
	"time"
)

func getPolicies(c *gin.Context) {
	cache := c.MustGet("cache").(otter.Cache[string, any])
	if cache.Has("policies") {
		var policies, _ = cache.Get("policies")
		c.IndentedJSON(http.StatusOK, policies)
	} else {
		ctx := context.Background()
		client, err := backend.AutoAuth(ctx)
		if err != nil {
			log.Printf("error: %v", err)
		}
		policies, err := GetPolicies(ctx, client)
		if err != nil {
			log.Println(err)
		}
		cache.Set("policies", policies)
		c.IndentedJSON(http.StatusOK, policies)
	}
}

func healthz(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

func info(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"version":      "0.0.1",
		"vaultAddress": os.Getenv("VAULT_ADDR"),
		"kvEngine":     os.Getenv("VAULT_KV_ENGINE"),
	})
}

func getPaths(c *gin.Context) {
	cache := c.MustGet("cache").(otter.Cache[string, any])
	if cache.Has("paths") {
		var paths, _ = cache.Get("paths")
		c.IndentedJSON(http.StatusOK, paths)
	} else {
		ctx := context.Background()
		client, err := backend.AutoAuth(ctx)
		if err != nil {
			log.Printf("error: %v", err)
		}
		paths, err := GetPaths(ctx, client)
		if err != nil {
			log.Println(err)
		}
		cache.Set("paths", paths)
		c.IndentedJSON(http.StatusOK, paths)
	}
}

func getGraph(c *gin.Context) {
	cache := c.MustGet("cache").(otter.Cache[string, any])
	if cache.Has("graph") {
		var paths, _ = cache.Get("graph")
		c.IndentedJSON(http.StatusOK, paths)
	} else {
		ctx := context.Background()
		client, err := backend.AutoAuth(ctx)
		if err != nil {
			log.Printf("error: %v", err)
		}
		paths, err := getGraphPaths(ctx, client, -1)
		if err != nil {
			log.Println(err)
		}

		cache.Set("graph", paths)
		c.IndentedJSON(http.StatusOK, paths)
	}
}

func compressedGraphLevel(c *gin.Context) {
	ctx := context.Background()
	level := c.Query("l")
	//convert level to int
	l, _ := strconv.Atoi(level)
	client, err := backend.AutoAuth(ctx)

	if err != nil {
		log.Printf("error: %v", err)
	}
	paths, err := getGraphPaths(ctx, client, l)
	if err != nil {
		log.Println(err)
	}

	root := models.CompressedGraphEntry{
		Prefix:   paths.AbsolutePath,
		Children: []models.CompressedGraphEntry{},
	}

	for _, path := range paths.Children {
		root.Children = append(root.Children, models.CompressedGraphEntry{
			Prefix:   path.AbsolutePath,
			Children: appendChildren(ctx, path.AbsolutePath, path, l-1),
		})
	}
	possible := []models.CompressedGraphEntry{root}
	for i := 0; i < l; i++ {
		newOnes := []models.CompressedGraphEntry{}
		for _, path := range possible {
			println(path.Prefix)
			newChildren := []models.CompressedGraphEntry{}
			for _, child := range path.Children {
				absolutePath := child.Prefix
				if path.Prefix != "/" {
					absolutePath = path.Prefix + "/" + child.Prefix
				}
				what := models.CompressedGraphEntry{
					Prefix:   absolutePath,
					Children: child.Children,
				}
				newChildren = append(newChildren, what)
			}
			newOnes = append(newOnes, newChildren...)
		}
		possible = newOnes
		if len(possible) == 0 {
			c.IndentedJSON(http.StatusOK, nil)
			return
		}
	}
	children := []models.CompressedGraphEntry{}
	for _, path := range possible {
		path.Children = nil
		children = append(children, path)
	}
	root.Children = children
	c.IndentedJSON(http.StatusOK, root)
}

func getCompressedGraph(ctx context.Context, client *vault.Client) (models.CompressedGraphEntry, error) {
	paths, err := getGraphPaths(ctx, client, -1)
	if err != nil {
		log.Println(err)
	}

	root := models.CompressedGraphEntry{
		Prefix:   paths.AbsolutePath,
		Children: []models.CompressedGraphEntry{},
	}

	for _, path := range paths.Children {
		root.Children = append(root.Children, models.CompressedGraphEntry{
			Prefix:   path.Name,
			Children: appendChildren(ctx, path.AbsolutePath, path, -1),
		})
	}
	return root, nil
}

func compressedGraph(c *gin.Context) {
	ctx := context.Background()
	client, err := backend.AutoAuth(ctx)
	cache := c.MustGet("cache").(otter.Cache[string, any])

	if err != nil {
		log.Printf("error: %v", err)
	}

	if cache.Has("compressed-graph") {
		var paths, _ = cache.Get("compressed-graph")
		c.IndentedJSON(http.StatusOK, paths)
	} else {
		compressedGraph, _ := getCompressedGraph(ctx, client)
		cache.Set("compressed-graph", compressedGraph)
		c.IndentedJSON(http.StatusOK, compressedGraph)
	}
}

func appendChildren(ctx context.Context, prefix string, nodes models.GraphEntry, stopAtRecursion int) []models.CompressedGraphEntry {
	if stopAtRecursion == 0 {
		return nil
	}
	if stopAtRecursion > 0 {
		stopAtRecursion--
	}
	if len(nodes.Children) == 0 {
		return nil
	}
	var compressed models.CompressedGraphEntry
	compressed.Prefix = prefix
	var children []models.CompressedGraphEntry
	for _, node := range nodes.Children {
		children = append(children, models.CompressedGraphEntry{
			Prefix:   node.Name,
			Children: appendChildren(ctx, node.AbsolutePath, node, stopAtRecursion),
		})
	}
	return children
}

func getAnnotatedSecret(c *gin.Context) {
	path := c.Query("path")
	cache := c.MustGet("cache").(otter.Cache[string, any])
	if !cache.Has("annotatedSecrets") {
		_, err := annotateSecrets(context.Background(), cache)
		if err != nil {
			return
		}
	}
	if cache.Has(path) {
		var analyzedSecret, _ = cache.Get(path)
		c.IndentedJSON(http.StatusOK, analyzedSecret)
	} else {
		c.IndentedJSON(http.StatusNotFound, []string{})
	}
}

func annotateSecrets(ctx context.Context, cache otter.Cache[string, any]) ([]models.AnnotatedSecret, error) {
	client, err := backend.AutoAuth(ctx)
	if err != nil {
		log.Printf("could not authenticate: %v", err)
		return nil, err
	}
	paths, err := GetPaths(ctx, client)
	if err != nil {
		log.Printf("could not get paths: %v", err)
		return nil, err
	}
	policies, _ := GetPolicies(ctx, client)
	var analyzedPaths = []models.AnnotatedSecret{}
	for _, path := range paths {
		var accessiblePolicies []models.Policy
		for _, policy := range policies {
			if policy.HasAccessTo(path.Path) {
				accessiblePolicies = append(accessiblePolicies, policy)
			}
		}
		for _, policy := range accessiblePolicies {
			if !cache.Has(path.Path) {
				cache.Set(path.Path, []string{policy.Name})
			} else {
				entry, _ := cache.Get(path.Path)
				cached := entry.([]string)
				cached = append(cached, policy.Name)
				// make a set out of the cached values
				set := make(map[string]struct{})
				for _, value := range cached {
					set[value] = struct{}{}
				}
				cached = []string{}
				for key := range set {
					cached = append(cached, key)
				}
				cache.Set(path.Path, cached)
			}
		}
		analyzedPaths = append(analyzedPaths, models.AnnotatedSecret{Path: path, Policies: accessiblePolicies})
	}

	cache.Set("annotatedSecrets", analyzedPaths)
	return analyzedPaths, nil
}

func getAnnotatedSecrets(c *gin.Context) {
	cache := c.MustGet("cache").(otter.Cache[string, any])
	if cache.Has("annotatedSecrets") {
		var analyzedSecret, _ = cache.Get("annotatedSecrets")
		c.IndentedJSON(http.StatusOK, analyzedSecret)
	} else {
		response, _ := annotateSecrets(context.Background(), cache)
		c.IndentedJSON(http.StatusOK, response)
	}
}

func CacheProvider() gin.HandlerFunc {
	cache, err := otter.MustBuilder[string, any](10_000).
		CollectStats().
		Cost(func(key string, value any) uint32 {
			return 1
		}).
		WithTTL(5 * time.Minute).
		Build()
	if err != nil {
		panic(err)
	}
	return func(c *gin.Context) {
		c.Set("cache", cache)
		c.Next()
	}
}

func UpdateCaches(c *gin.Context) {
	log.Printf("updating caches")
	cache := c.MustGet("cache").(otter.Cache[string, any])
	client, err := backend.AutoAuth(context.Background())
	if err != nil {
		log.Printf("error: %v", err)
	}
	compressedGraph, _ := getCompressedGraph(context.Background(), client)
	cache.Set("compressed-graph", compressedGraph)
}

func callUpdateEndpoint() (map[string]interface{}, error) {
	// Create a request to the /hello endpoint
	resp, err := http.Get("http://localhost:8081/update")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Decode the JSON response
	var response map[string]interface{}

	return response, nil
}

func main() {
	router := gin.New()
	scheduler, err := gocron.NewScheduler()
	router.Use(
		gin.LoggerWithWriter(gin.DefaultWriter, "/v1/healthz"),
		gin.Recovery(),
	)

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"*"},
		AllowHeaders:     []string{"Origin"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	router.Use(CacheProvider())
	router.GET("/v1/info", info)
	router.GET("/v1/healthz", healthz)
	router.GET("/v1/paths", getPaths)
	router.GET("/v1/level", compressedGraphLevel)
	router.GET("/v1/graph", compressedGraph)
	router.GET("/v1/policies", getPolicies)
	router.GET("/v1/annotated", getAnnotatedSecret)
	router.GET("/v1/annotatedSecrets", getAnnotatedSecrets)
	router.GET("/update", func(c *gin.Context) {
		UpdateCaches(c)
	})

	job, err := scheduler.NewJob(
		gocron.DurationJob(
			3*time.Minute,
		),
		gocron.NewTask(callUpdateEndpoint),
	)

	log.Printf("job: %v", job)
	scheduler.Start()

	err = router.Run(":8081")
	if err != nil {
		return
	}
}
