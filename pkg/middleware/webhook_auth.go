package middleware

import (
	"bone_appetit_r4_service/pkg/r4bank"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type WebhookAuthMiddleware struct {
	secret        string
	commerceToken string
}

func NewWebhookAuthMiddleware(secret, commerceToken string) *WebhookAuthMiddleware {
	return &WebhookAuthMiddleware{
		secret:        secret,
		commerceToken: commerceToken,
	}
}

// Auth middleware function to validate webhook requests
func (m *WebhookAuthMiddleware) Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			fmt.Println("Missing Authorization header")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"abono": false})
			return
		}

		// Validate the authorization header
		if !r4bank.ValidateAuthToken(m.commerceToken, m.secret, authHeader) {
			fmt.Println("Invalid Authorization token", authHeader)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"abono": false})
			return
		}

		c.Next()
	}
}
