package urls

import (
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"inner/conf/platform_conf"
	"os"
	"os/exec"
)

func init() {
	go func() {
		<-platform_conf.Qch
		os.Exit(0)
	}()
	if len(os.Args) > 1 {
		if os.Args[1] == "-d" {
			cmd := exec.Command(os.Args[0])
			if err := cmd.Start(); err != nil {
				fmt.Printf("start %s failed, error: %v\n", os.Args[0], err)
				os.Exit(1)
			}
			os.Exit(0)
		} else {
			fmt.Println("run app as a daemon with -d")
		}
	}
}
func Use(r *gin.Engine) {
	// 加入跨域中间件
	CorsConfig := cors.DefaultConfig()
	CorsConfig.AllowAllOrigins = true
	CorsConfig.AllowCredentials = true
	CorsConfig.AllowMethods = []string{"PUT", "PATCH", "POST", "GET", "DELETE", "OPTIONS", "HEADER"}
	CorsConfig.AllowHeaders = []string{"*"}
	CorsConfig.ExposeHeaders = []string{"Authorization"}
	r.Use(cors.New(CorsConfig))
	platform_conf.RootPath, _ = os.Getwd()
	//静态资源
	r.Static("/api/v1/ag/", platform_conf.RootPath+"/opsone/static/agent")
	r.Static("/api/v1/conf/", platform_conf.RootPath+"/opsone/static/config")
	r.Static("/api/v1/webshell/", platform_conf.RootPath+"/opsone/static/webshell")
	// 加入swagger中间件
	url := ginSwagger.URL("/api/v1/swagger/doc.json")
	r.GET("/api/v1/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, url))
	//性能调试
	//pprof.Register(r)
	// 添加路由信息
	CmdbGroup(r)
	PlatformGroup(r)
	MonitorGroup(r)
	MsgGroup(r)
	AuthGroup(r)
	WorkOrderGroup(r)
	JobGroup(r)
	CloudGroup(r)
	K8sGroup(r)
	//注册路由
	RegisterRouter(r)
}
