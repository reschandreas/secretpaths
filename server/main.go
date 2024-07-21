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

	err := router.Run("localhost:8080")
	if err != nil {
		return
	}
}

func print_policies() {
	ctx := context.Background()

	client := backend.UseToken("my-token")

	policies, err := getPolicies(ctx, client)

	if err != nil {
		log.Fatal(err)
	}
	log.Println("rules:", policies)
	for _, policy := range policies {
		log.Printf("policy: %s has %d rules", policy.Name, policy.AmountOfPolicies())
		if policy.AmountOfPolicies() > 0 {
			for _, rule := range policy.Rules {
				log.Printf("rule: %s has %s capabilities", rule.Path, rule.Capabilities)
			}
		}
	}
	log.Println("policies:", len(policies))
	log.Println("")
	log.Println("")
	log.Println("")

	needle := "secret/bar/zip"

	for _, policy := range policies {
		if policy.Name != "policy_testing" {
			continue
		}
		if policy.HasAccessTo(needle) {
			log.Printf("policy: %s has access to %s", policy.Name, needle)
		} else {
			log.Printf("policy: %s has no access to %s", policy.Name, needle)
		}
	}
}
