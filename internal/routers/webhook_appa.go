package routers

import (
	"bone_appetit_r4_service/internal/handlers"
	"bone_appetit_r4_service/pkg/middleware"

	"github.com/gin-gonic/gin"
)

type WebhookAppaRouter struct {
	webhookHandler *handlers.WebhookHandler
}

func NewWebhookAppaRouter(webhookHandler *handlers.WebhookHandler) *WebhookAppaRouter {
	return &WebhookAppaRouter{webhookHandler: webhookHandler}
}

// SetRouter sets up the webhook-related routes
func (w *WebhookAppaRouter) SetRouter(router *gin.Engine, auth *middleware.WebhookAuthMiddleware) {
	group := router.Group("/appa", auth.Auth())
	group.POST("/R4consulta", w.webhookHandler.HandlerAppaR4Consulta)
	group.POST("/R4notifica", w.webhookHandler.HandlerAppaR4Notifica)
}
