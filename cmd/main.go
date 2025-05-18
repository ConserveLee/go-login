package main

import (
	"github.com/gin-gonic/gin"
	"go-login/internal/api"
	"go-login/tools"
)

func main() {
	router := gin.Default()
	api.SetupRoutes(router)

	_ = router.Run(":3001")
	close(tools.GetQueue())
}
