package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/xuewenG/songlist-proxy/pkg/config"
	"github.com/xuewenG/songlist-proxy/pkg/router"
)

var mode string
var version string
var commitId string

func main() {
	err := config.InitConfig()
	if err != nil {
		log.Fatalf("Init config failed, %v", err)
	}

	log.Printf(
		"Server is starting:\nmode: %s\nversion: %s\ncommitId: %s\nport: %s\nbiliUid: %s\nbiliUrl: %s\nbiliAvatar: %s\n",
		mode,
		version,
		commitId,
		config.Config.App.Port,
		config.Config.Bili.Uid,
		config.Config.Bili.Url,
		config.Config.Bili.Avatar,
	)

	if config.Config.App.Port == "" {
		log.Fatalf("Invalid port: %s\n", config.Config.App.Port)
		return
	}

	if mode == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()
	r.SetTrustedProxies(nil)

	router.Bind(r)

	r.Run(fmt.Sprintf(":%s", config.Config.App.Port))
}
