package server

import (
	"swd-new/internal/handler"
	"swd-new/internal/middleware"
	"swd-new/pkg/helper/resp"
	"swd-new/pkg/log"

	"github.com/gin-gonic/gin"
)

func NewServerHTTP(
	logger *log.Logger,
	sensitiveWordHandler *handler.SensitiveWordHandler,
) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(
		middleware.CORSMiddleware(),
	)
	r.GET("/", func(ctx *gin.Context) {
		resp.HandleSuccess(ctx, map[string]interface{}{
			"say": "Hi Nunu!",
		})
	})
	r.POST("/check", sensitiveWordHandler.Check)

	return r
}
