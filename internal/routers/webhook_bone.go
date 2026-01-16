package routers

import (
	"bone_appetit_r4_service/internal/handlers"
	"bone_appetit_r4_service/pkg/middleware"

	"github.com/gin-gonic/gin"
)

type WebhookRouter struct {
	webhookHandler *handlers.WebhookHandler
}

func NewWebhookRouter(webhookHandler *handlers.WebhookHandler) *WebhookRouter {
	return &WebhookRouter{webhookHandler: webhookHandler}
}

// SetRouter sets up the webhook-related routes
func (w *WebhookRouter) SetRouter(router *gin.Engine, auth *middleware.WebhookAuthMiddleware) {
	group := router.Group("/", auth.Auth())
	group.POST("/R4consulta", w.webhookHandler.HandlerBoneR4Consulta)
	group.POST("/R4notifica", w.webhookHandler.HandlerBoneR4Notifica)
}
