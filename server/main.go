package main

import (
	"context"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/maypok86/otter"
	"log"
	"net/http"
	"os"
	"secretpaths/backend"
	"secretpaths/models"
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
		paths, err := getGraphPaths(ctx, client)
		if err != nil {
			log.Println(err)
		}

		cache.Set("graph", paths)
		c.IndentedJSON(http.StatusOK, paths)
	}
}

func getAnalyzedSecret(c *gin.Context) {
	path := c.Query("path")
	cache := c.MustGet("cache").(otter.Cache[string, any])
	if !cache.Has("analyzedSecrets") {
		_, err := analyzeSecrets(context.Background(), cache)
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

func analyzeSecrets(ctx context.Context, cache otter.Cache[string, any]) ([]models.AnalyzedSecret, error) {
	client, err := backend.AutoAuth(ctx)
	if err != nil {
		log.Printf("could not authenticate: %v", err)
	}
	paths, err := GetPaths(ctx, client)
	if err != nil {
		log.Printf("could not get paths: %v", err)
	}
	policies, _ := GetPolicies(ctx, client)
	var analyzedPaths []models.AnalyzedSecret
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
		analyzedPaths = append(analyzedPaths, models.AnalyzedSecret{Path: path, Policies: accessiblePolicies})
	}
	cache.Set("analyzedSecrets", analyzedPaths)
	return analyzedPaths, nil
}

func getAnalyzedSecrets(c *gin.Context) {
	cache := c.MustGet("cache").(otter.Cache[string, any])
	if cache.Has("analyzedSecrets") {
		var analyzedSecret, _ = cache.Get("analyzedSecrets")
		c.IndentedJSON(http.StatusOK, analyzedSecret)
	} else {
		response, _ := analyzeSecrets(context.Background(), cache)
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

func main() {
	router := gin.Default()

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
	router.GET("/healthz", healthz)
	router.GET("/v1/healthz", healthz)
	router.GET("/v1/paths", getPaths)
	router.GET("/v1/graph", getGraph)
	router.GET("/v1/policies", getPolicies)
	router.GET("/v1/analyzed", getAnalyzedSecret)
	router.GET("/v1/analyzedSecrets", getAnalyzedSecrets)

	err := router.Run(":8081")
	if err != nil {
		return
	}
}
