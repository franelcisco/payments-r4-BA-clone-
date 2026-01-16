package routers

import (
	"bone_appetit_r4_service/internal/handlers"
	"bone_appetit_r4_service/pkg/middleware"

	"github.com/gin-gonic/gin"
)

type r4AppaRoutes struct {
	r4Handler *handlers.R4Handler
}

func NewR4AppaRoutes(r4Handler *handlers.R4Handler) *r4AppaRoutes {
	return &r4AppaRoutes{r4Handler: r4Handler}
}

// SetRouter sets up the R4-related routes
func (p *r4AppaRoutes) SetRouter(router *gin.Engine, auth *middleware.WebhookAuthMiddleware) {
	group := router.Group("/r4/appa", auth.Auth())
	group.GET("/bcv-tasa", p.r4Handler.GetBCVTasa)
	group.POST("/generate-otp", p.r4Handler.HandleGenerateOTP)
	group.POST("/validate-immediate-debit", p.r4Handler.HandleValidateImmediateDebit)
	group.POST("/change-paid", p.r4Handler.HandleChangePaid)
	group.GET("/get-operation/:id", p.r4Handler.HandleGetOperationByID)
}
