package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Health Is Export to check
func Health(ctx *gin.Context) {
	ctx.Writer.WriteHeader(http.StatusOK)
}
