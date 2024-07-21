package main

import (
	"context"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"secretpaths/backend"
	"secretpaths/models"
	"time"
)

func getPoliciesAPI(c *gin.Context) {
	ctx := context.Background()
	client := backend.UseAppRole(ctx)
	policies, err := getPolicies(ctx, client)
	if err != nil {
		log.Fatal(err)
	}
	c.IndentedJSON(http.StatusOK, policies)
}

func getPathsAPI(c *gin.Context) {
	ctx := context.Background()
	client := backend.UseAppRole(ctx)
	paths, err := getPaths(ctx, client)
	if err != nil {
		log.Fatal(err)
	}

	c.IndentedJSON(http.StatusOK, paths)
}

func getGraphApi(c *gin.Context) {
	ctx := context.Background()
	client := backend.UseAppRole(ctx)
	paths, err := getGraphPaths(ctx, client)
	if err != nil {
		log.Fatal(err)
	}

	c.IndentedJSON(http.StatusOK, paths)
}

func getAnalyzedSecretsAPI(c *gin.Context) {
	ctx := context.Background()
	client := backend.UseAppRole(ctx)
	paths, err := getPaths(ctx, client)
	if err != nil {
		log.Fatal(err)
	}
	policies, _ := getPolicies(ctx, client)
	analyzedPaths := []models.AnalyzedSecret{}

	for _, path := range paths {
		accessiblePolicies := []models.Policy{}
		for _, policy := range policies {
			if policy.HasAccessTo(path.Path) {
				accessiblePolicies = append(accessiblePolicies, policy)
			}
		}
		analyzedPaths = append(analyzedPaths, models.AnalyzedSecret{Path: path, Policies: accessiblePolicies})
	}

	c.IndentedJSON(http.StatusOK, analyzedPaths)
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
	router.GET("/paths", getPathsAPI)
	router.GET("/graph", getGraphApi)
	router.GET("/policies", getPoliciesAPI)
	router.GET("/analyzedSecrets", getAnalyzedSecretsAPI)

	err := router.Run("0.0.0.0:8080")
	if err != nil {
		return
	}
}
