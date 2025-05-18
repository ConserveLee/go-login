package api

import (
	"github.com/gin-gonic/gin"
	v1 "go-login/internal/api/handlers/v1"
	v2 "go-login/internal/api/handlers/v2"
)

func SetupRoutes(router *gin.Engine) {
	router.POST("v1", v1.Login)
	router.POST("v2", v2.Login)
}
