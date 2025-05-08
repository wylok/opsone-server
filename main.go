package main

import (
	"github.com/Depado/ginprom"
	"github.com/gin-gonic/gin"
	_ "inner/docs"
	_ "inner/task/auth"
	_ "inner/task/cloud"
	_ "inner/task/cmdb"
	_ "inner/task/job"
	_ "inner/task/k8s"
	_ "inner/task/monitor"
	_ "inner/task/platform"
	"inner/urls"
)

func main() {
	// @title 智能运维管理平台
	// @version v4.5
	// @securityDefinitions.apikey ApiKeyAuth
	// @in header
	// @name token
	gin.SetMode(gin.DebugMode)
	r := gin.Default()
	p := ginprom.New(
		ginprom.Engine(r),
		ginprom.Subsystem("opsone"),
		ginprom.Path("/api/v1/metrics"),
	)
	r.Use(p.Instrument())
	urls.Use(r)
	_ = r.Run(":8888")
}
