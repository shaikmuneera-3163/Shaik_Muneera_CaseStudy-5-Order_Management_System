package graphql

import (
	"github.com/gin-gonic/gin"
	gql "github.com/graphql-go/graphql"
	graphqlhandler "github.com/graphql-go/handler"
)

func NewHandler(schema gql.Schema) gin.HandlerFunc {
	h := graphqlhandler.New(&graphqlhandler.Config{
		Schema:   &schema,
		Pretty:   true,
		GraphiQL: true,
	})

	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}
