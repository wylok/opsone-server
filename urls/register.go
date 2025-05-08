package urls

import (
	"github.com/gin-gonic/gin"
	"inner/conf/platform_conf"
	"inner/modules/common"
	"inner/modules/kits"
	"time"
)

func RegisterRouter(r *gin.Engine) {
	var (
		log     = kits.Log{}
		Routes  = r.Routes()
		rc, ctx = common.RedisConnect()
	)
	go func() {
		for {
			log.Info("RegisterRouter start working ......")
			routeKeys := map[string]struct{}{}
			for _, v := range Routes {
				routeKeys[v.Handler] = struct{}{}
				rc.HSet(ctx, platform_conf.RouterKey, v.Handler, v.Path+":"+v.Method)
			}
			data := rc.HGetAll(ctx, platform_conf.RouterKey).Val()
			if len(data) > 0 {
				for k := range data {
					_, ok := routeKeys[k]
					if !ok {
						rc.HDel(ctx, platform_conf.RouterKey, k)
					}
				}
			}
			time.Sleep(5 * time.Minute)
		}
	}()
}
