package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/louai60/e-commerce_project/backend/api-gateway/handlers"
)

// SetupGraphQLRoutes sets up GraphQL routes
func SetupGraphQLRoutes(r *gin.Engine, graphqlHandler *handlers.GraphQLHandler) {
	// GraphQL endpoint
	graphql := r.Group("/api/v1/graphql")
	{
		// Allow public access to GraphQL endpoint for queries
		graphql.POST("", graphqlHandler.Handle)
		graphql.GET("", graphqlHandler.Handle) // For GraphiQL interface
	}
}
