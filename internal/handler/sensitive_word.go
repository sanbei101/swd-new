package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"swd-new/internal/service"
)

type SensitiveWordHandler struct {
	*Handler
	sensitiveWordService service.SensitiveWordService
}

type checkRequest struct {
	Text string `json:"text" binding:"required"`
}

func NewSensitiveWordHandler(handler *Handler, sensitiveWordService service.SensitiveWordService) *SensitiveWordHandler {
	return &SensitiveWordHandler{
		Handler:              handler,
		sensitiveWordService: sensitiveWordService,
	}
}

func (h *SensitiveWordHandler) Check(ctx *gin.Context) {
	var req checkRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.sensitiveWordService.Check(req.Text)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, result)
}
