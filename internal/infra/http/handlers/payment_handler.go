package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/williamkoller/golang-payment-stripe/internal/app/service"
)

type PaymentHandler struct {
	svc *service.PaymentService
}

func NewPaymentHandler(svc *service.PaymentService) *PaymentHandler { return &PaymentHandler{svc: svc} }

type createReq struct {
	Amount   int64  `json:"amount" example:"5500"`
	Currency string `json:"currency" example:"brl"`
	Email    string `json:"email" example:"cliente@example.com"`
}

// POST /v1/payments -> cria e autoriza (captura manual)
func (h *PaymentHandler) Create(c *gin.Context) {
	var req createReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload", "details": err.Error()})
		return
	}
	out, err := h.svc.CreateAndAuthorize(c.Request.Context(), service.CreateInput{
		Amount: req.Amount, Currency: req.Currency, Email: req.Email,
	})
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, out)
}

// POST /v1/payments/:id/capture -> captura fundos autorizados
func (h *PaymentHandler) Capture(c *gin.Context) {
	id := c.Param("id")
	out, err := h.svc.Capture(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, out)
}

// POST /v1/payments/:id/cancel -> cancela autorização
func (h *PaymentHandler) Cancel(c *gin.Context) {
	id := c.Param("id")
	out, err := h.svc.Cancel(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, out)
}

// GET /v1/payments/:id
func (h *PaymentHandler) Get(c *gin.Context) {
	id := c.Param("id")
	out, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, out)
}
