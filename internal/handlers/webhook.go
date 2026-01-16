package handlers

import (
	"bone_appetit_r4_service/internal/models"
	"bone_appetit_r4_service/internal/services"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type WebhookHandler struct {
	service services.WebhookService
}

func NewWebhookHandler(service services.WebhookService) *WebhookHandler {
	return &WebhookHandler{
		service: service,
	}
}

// HandlerBoneR4Consulta is the handler for the R4Consulta webhook
func (h *WebhookHandler) HandlerBoneR4Consulta(c *gin.Context) {
	var request models.R4ConsultaRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		fmt.Printf("Error binding JSON: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"status": false})
		return
	}

	// Process asynchronously
	go func(request models.R4ConsultaRequest) {
		if err := h.service.RegisterR4MobilePaymentPreview(&request, "bone"); err != nil {
			fmt.Printf("Error registering R4 mobile payment preview: %v\n", err)
		}
	}(request)

	c.JSON(http.StatusOK, gin.H{
		"status": true,
	})
}

// HandlerBoneR4Notifica is the handler for the R4Notifica webhook
func (h *WebhookHandler) HandlerBoneR4Notifica(c *gin.Context) {
	var request models.R4NotificaRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		fmt.Printf("Error binding JSON: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"status": false})
		return
	}

	err := h.service.RegisterR4MobilePayment(&request, "bone")
	if err != nil {
		fmt.Printf("Error registering R4 mobile payment: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"status": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": true,
	})
}

// HandlerAppaR4Consulta is the handler for the R4Consulta webhook
func (h *WebhookHandler) HandlerAppaR4Consulta(c *gin.Context) {
	var request models.R4ConsultaRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		fmt.Printf("Error binding JSON: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"status": false})
		return
	}

	// Process asynchronously
	go func(request models.R4ConsultaRequest) {
		if err := h.service.RegisterR4MobilePaymentPreview(&request, "appa"); err != nil {
			fmt.Printf("Error registering R4 mobile payment preview: %v\n", err)
		}
	}(request)

	c.JSON(http.StatusOK, gin.H{
		"status": true,
	})
}

// HandlerAppaR4Notifica is the handler for the R4Notifica webhook
func (h *WebhookHandler) HandlerAppaR4Notifica(c *gin.Context) {
	var request models.R4NotificaRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		fmt.Printf("Error binding JSON: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"status": false})
		return
	}

	err := h.service.RegisterR4MobilePayment(&request, "appa")
	if err != nil {
		fmt.Printf("Error registering R4 mobile payment: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"status": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": true,
	})
}
