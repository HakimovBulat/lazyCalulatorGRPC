package main

import (
	"github.com/HakimovBulat/lazyCalulatorGRPC/router"
	"github.com/HakimovBulat/lazyCalulatorGRPC/utils"
	"go.uber.org/zap"
)

func main() {
	router := router.SetupRouter()
	if err := router.Run(":8080"); err != nil {
		utils.Logger.Error(err.Error(), zap.String("router", "f*ck down"))
	}
}
