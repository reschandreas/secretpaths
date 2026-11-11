package main

import (
	"context"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/go-co-op/gocron/v2"
	"github.com/hashicorp/vault-client-go"
	"golang.org/x/sync/errgroup"
	"github.com/maypok86/otter"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"secretpaths/backend"
	"secretpaths/models"
	"strconv"
	"sync"
	"time"
)

func getPolicies(c *gin.Context) {
	cache := c.MustGet("cache").(otter.Cache[string, any])
	if cached, ok := cache.Get("policies"); ok {
		c.IndentedJSON(http.StatusOK, cached)
		return
	}
	ctx := context.Background()
	client, err := backend.AutoAuth(ctx)
	if err != nil {
		log.Printf("error: %v", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}
	policies, err := GetPolicies(ctx, client)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	cache.Set("policies", policies)
	c.IndentedJSON(http.StatusOK, policies)
}

func healthz(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

func readyz(c *gin.Context) {
	ctx := context.Background()
	client, err := backend.AutoAuth(ctx)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not ready", "error": err.Error()})
		return
	}
	// Verify Vault is reachable and authenticated.
	if _, err := client.System.ReadHealthStatus(ctx); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not ready", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ready"})
}

func info(c *gin.Context) {
	engine := os.Getenv("VAULT_KV_ENGINE")
	if engine == "" {
		engine = "secret"
	}
	c.JSON(http.StatusOK, gin.H{
		"version":      "0.0.2",
		"vaultAddress": os.Getenv("VAULT_ADDR"),
		"kvEngine":     engine,
	})
}

func getPaths(c *gin.Context) {
	cache := c.MustGet("cache").(otter.Cache[string, any])
	q := c.Query("q")
	limitStr := c.Query("limit")
	offsetStr := c.Query("offset")

	var all []models.Secret
	if cached, ok := cache.Get("paths"); ok {
		all = cached.([]models.Secret)
	} else {
		ctx := context.Background()
		client, err := backend.AutoAuth(ctx)
		if err != nil {
			log.Printf("error: %v", err)
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
			return
		}
		paths, err := GetPaths(ctx, client)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
			return
		}
		all = paths
		cache.Set("paths", all)
	}

	filtered := filterSecrets(all, q)
	limit := 0
	if limitStr != "" {
		fmt.Sscan(limitStr, &limit)
	}
	offset := 0
	if offsetStr != "" {
		fmt.Sscan(offsetStr, &offset)
	}
	c.IndentedJSON(http.StatusOK, paginateSlice(filtered, offset, limit))
}

func compressedGraphLevel(c *gin.Context) {
	ctx := context.Background()
	level := c.Query("l")
	l, _ := strconv.Atoi(level)
	cache := c.MustGet("cache").(otter.Cache[string, any])
	cacheKey := fmt.Sprintf("compressed-graph-level-%d", l)
	if cached, ok := cache.Get(cacheKey); ok {
		c.IndentedJSON(http.StatusOK, cached)
		return
	}
	client, err := backend.AutoAuth(ctx)
	if err != nil {
		log.Printf("error: %v", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}
	paths, err := getGraphPaths(ctx, client, l)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
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
	cache.Set(cacheKey, root)
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
	cache := c.MustGet("cache").(otter.Cache[string, any])
	client, err := backend.AutoAuth(ctx)
	if err != nil {
		log.Printf("error: %v", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
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

	analyzedPaths := make([]models.AnnotatedSecret, len(paths))
	var mu sync.Mutex
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(16)
	for i, path := range paths {
		i, path := i, path
		g.Go(func() error {
			var accessiblePolicies []models.Policy
			for _, policy := range policies {
				if policy.HasAccessTo(path.Path) {
					accessiblePolicies = append(accessiblePolicies, policy)
				}
			}
			// update per-path cache of policy names
			names := make([]string, 0, len(accessiblePolicies))
			for _, p := range accessiblePolicies {
				names = append(names, p.Name)
			}
			mu.Lock()
			if !cache.Has(path.Path) {
				cache.Set(path.Path, names)
			} else {
				entry, _ := cache.Get(path.Path)
				cached := entry.([]string)
				cached = append(cached, names...)
				set := make(map[string]struct{})
				for _, v := range cached {
					set[v] = struct{}{}
				}
				cached = cached[:0]
				for key := range set {
					cached = append(cached, key)
				}
				cache.Set(path.Path, cached)
			}
			analyzedPaths[i] = models.AnnotatedSecret{Path: path, Policies: accessiblePolicies}
			mu.Unlock()
			_ = gctx
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}

	cache.Set("annotatedSecrets", analyzedPaths)
	return analyzedPaths, nil
}

func getAnnotatedSecrets(c *gin.Context) {
	cache := c.MustGet("cache").(otter.Cache[string, any])
	q := c.Query("q")
	limitStr := c.Query("limit")
	offsetStr := c.Query("offset")

	var all []models.AnnotatedSecret
	if cached, ok := cache.Get("annotatedSecrets"); ok {
		all = cached.([]models.AnnotatedSecret)
	} else {
		response, err := annotateSecrets(context.Background(), cache)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
			return
		}
		all = response
		cache.Set("annotatedSecrets", all)
	}

	filtered := filterAnnotatedSecrets(all, q)
	limit := 0
	if limitStr != "" {
		fmt.Sscan(limitStr, &limit)
	}
	offset := 0
	if offsetStr != "" {
		fmt.Sscan(offsetStr, &offset)
	}
	c.IndentedJSON(http.StatusOK, paginateSlice(filtered, offset, limit))
}

func filterSecrets(all []models.Secret, q string) []models.Secret {
	if q == "" {
		return all
	}
	out := make([]models.Secret, 0, len(all))
	for _, s := range all {
		if strings.Contains(s.Path, q) {
			out = append(out, s)
		}
	}
	return out
}

func filterAnnotatedSecrets(all []models.AnnotatedSecret, q string) []models.AnnotatedSecret {
	if q == "" {
		return all
	}
	out := make([]models.AnnotatedSecret, 0, len(all))
	for _, a := range all {
		if strings.Contains(a.Path.Path, q) {
			out = append(out, a)
		}
	}
	return out
}

func paginateSlice[T any](items []T, offset, limit int) []T {
	if limit <= 0 {
		return items
	}
	if offset > len(items) {
		offset = len(items)
	}
	end := offset + limit
	if end > len(items) {
		end = len(items)
	}
	return items[offset:end]
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
	annotatedSecrets, _ := annotateSecrets(context.Background(), cache)
	cache.Set("annotatedSecrets", annotatedSecrets)
	paths, err := GetPaths(c, client)
	if err != nil {
		cache.Set("paths", paths)
	}
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

	corsOrigin := os.Getenv("CORS_ORIGIN")
	if corsOrigin == "" {
		corsOrigin = "http://localhost:5173"
	}
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{corsOrigin},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	router.Use(gzip.Gzip(gzip.DefaultCompression))
	router.Use(CacheProvider())
	router.GET("/v1/info", info)
	router.GET("/v1/healthz", healthz)
	router.GET("/v1/readyz", readyz)
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
