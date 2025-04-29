package main

import (
	"fmt"
	"server/controller"
	"server/utils"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func main(){
	// Load environment variables
	cfg := utils.LoadConfig()

	log.Infof("Starting GenDev server on port %d", cfg.Server.Port)
	gin.SetMode(gin.ReleaseMode)

	r := controller.SetupRouter()
	log.Panic(r.Run(fmt.Sprintf(":%d", cfg.Server.Port)))
}