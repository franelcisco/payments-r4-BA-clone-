package handlers

import (
	"bone_appetit_r4_service/internal/models"
	"bone_appetit_r4_service/internal/services"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type R4Handler struct {
	r4Service services.R4Service
}

func NewR4Handler(r4Service services.R4Service) *R4Handler {
	return &R4Handler{r4Service: r4Service}
}

// GetBCVTasa handles requests to get the BCV exchange rate for USD
func (p *R4Handler) GetBCVTasa(c *gin.Context) {
	tasa, err := p.r4Service.GetBCVTasaUSD(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tasa)
}

// HandleGenerateOTP handles requests to generate an OTP
func (p *R4Handler) HandleGenerateOTP(c *gin.Context) {
	var req models.OTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Printf("Error binding JSON: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	if err := p.r4Service.GenerateOTP(c, &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "OTP generated successfully"})
}

// HandleValidateImmediateDebit handles requests to validate an immediate debit transaction using OTP
func (p *R4Handler) HandleValidateImmediateDebit(c *gin.Context) {
	var req models.ValidateOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Printf("Error binding JSON: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	resp, err := p.r4Service.ValidateImmediateDebit(c, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// HandleChangePaid handles requests to change paid in Bolivares
func (p *R4Handler) HandleChangePaid(c *gin.Context) {
	var req models.ChangePaidRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Printf("Error binding JSON: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	resp, err := p.r4Service.ChangePaid(c, &req)
	if err != nil {
		fmt.Printf("Error processing ChangePaid request: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (p *R4Handler) HandleGetOperationByID(c *gin.Context) {
	operationID := c.Param("id")
	if operationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Operation ID is required"})
		return
	}

	resp, err := p.r4Service.GetOperationByID(c, operationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}
