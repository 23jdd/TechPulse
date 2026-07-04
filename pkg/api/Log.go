package api

import (
	"lib"

	"github.com/gin-gonic/gin"
)

// LogMiddleware I
func LogMiddleware() gin.HandlerFunc {
	return func(context *gin.Context) {
		lib.Log.Info("LogMiddleware")
	}
}
