package platform

import (
	"github.com/gorilla/websocket"
	"inner/modules/common"
	"inner/modules/databases"
	"inner/modules/kits"
	"net/http"
)

var (
	Log      kits.Log
	db       = databases.DB
	rc, ctx  = common.RedisConnect()
	UpGrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }, ReadBufferSize: 1024,
		WriteBufferSize: 1024}
)
