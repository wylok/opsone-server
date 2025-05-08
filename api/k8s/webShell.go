package k8s

import (
	"context"
	"errors"
	"github.com/duke-git/lancet/v2/convertor"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"inner/modules/common"
	"inner/modules/databases"
	"io"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	"net/http"
	"sync"
)

var (
	ClientSet *kubernetes.Clientset
	WsUp      = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		}}
	Rows string
	Cols string
)

// WsMessage websocket消息
type WsMessage struct {
	MessageType int
	Data        []byte
}

// WsConnection 封装websocket连接
type WsConnection struct {
	wsSocket  *websocket.Conn // 底层websocket
	inChan    chan *WsMessage // 读取队列
	outChan   chan *WsMessage // 发送队列
	mutex     sync.Mutex      // 避免重复关闭管道
	isClosed  bool
	closeChan chan byte // 关闭通知
}

// 读取协程
func (wsConn *WsConnection) wsReadLoop() {
	var (
		msgType int
		data    []byte
		msg     *WsMessage
		err     error
	)
	for {
		// 读一个message
		if msgType, data, err = wsConn.wsSocket.ReadMessage(); err != nil {
			goto ERROR
		}
		msg = &WsMessage{
			msgType,
			data,
		}
		// 放入请求队列
		select {
		case wsConn.inChan <- msg:
		case <-wsConn.closeChan:
			goto CLOSED
		}
	}
ERROR:
	wsConn.WsClose()
CLOSED:
}

// 发送协程
func (wsConn *WsConnection) wsWriteLoop() {
	var (
		msg *WsMessage
		err error
	)
	for {
		select {
		// 取一个应答
		case msg = <-wsConn.outChan:
			// 写给websocket
			if err = wsConn.wsSocket.WriteMessage(msg.MessageType, msg.Data); err != nil {
				goto ERROR
			}
		case <-wsConn.closeChan:
			goto CLOSED
		}
	}
ERROR:
	wsConn.WsClose()
CLOSED:
}

// InitWebsocket 并发安全API
func InitWebsocket(resp http.ResponseWriter, req *http.Request) (wsConn *WsConnection, err error) {
	var (
		wsSocket *websocket.Conn
	)
	// 应答客户端告知升级连接为websocket
	if wsSocket, err = WsUp.Upgrade(resp, req, nil); err != nil {
		return
	}
	wsConn = &WsConnection{
		wsSocket:  wsSocket,
		inChan:    make(chan *WsMessage, 1000),
		outChan:   make(chan *WsMessage, 1000),
		closeChan: make(chan byte),
		isClosed:  false,
	}
	// 读协程
	go wsConn.wsReadLoop()
	// 写协程
	go wsConn.wsWriteLoop()
	return
}

// WsWrite 发送消息
func (wsConn *WsConnection) WsWrite(messageType int, data []byte) (err error) {
	select {
	case wsConn.outChan <- &WsMessage{messageType, data}:
	case <-wsConn.closeChan:
		err = errors.New("websocket closed")
	}
	return
}

// WsRead 读取消息
func (wsConn *WsConnection) WsRead() (msg *WsMessage, err error) {
	select {
	case msg = <-wsConn.inChan:
		return
	case <-wsConn.closeChan:
		err = errors.New("websocket closed")
	}
	return
}

// WsClose 关闭连接
func (wsConn *WsConnection) WsClose() {
	_ = wsConn.wsSocket.Close()
	wsConn.mutex.Lock()
	defer wsConn.mutex.Unlock()
	if !wsConn.isClosed {
		wsConn.isClosed = true
		close(wsConn.closeChan)
	}
}

// ssh流式处理器
type streamHandler struct {
	wsConn      *WsConnection
	resizeEvent chan remotecommand.TerminalSize
	UserId      string
	k8s         string
}

// Next executor回调获取web是否resize
func (handler *streamHandler) Next() (size *remotecommand.TerminalSize) {
	ret := <-handler.resizeEvent
	size = &ret
	return
}

// executor回调读取web端的输入
func (handler *streamHandler) Read(p []byte) (size int, err error) {
	var (
		msg *WsMessage
	)
	// 读web发来的输入
	if msg, err = handler.wsConn.WsRead(); err != nil {
		return
	}
	// 解析客户端请求
	w, e1 := convertor.ToInt(Cols)
	h, e2 := convertor.ToInt(Rows)
	if Cols != "" && Rows != "" && e1 == nil && e2 == nil {
		handler.resizeEvent <- remotecommand.TerminalSize{Width: uint16(w), Height: uint16(h)}
	} else {
		handler.resizeEvent <- remotecommand.TerminalSize{}
	}
	// copy到p数组中
	size = len(string(msg.Data))
	copy(p, msg.Data)
	return
}

// executor回调向web端输出
func (handler *streamHandler) Write(p []byte) (size int, err error) {
	var (
		copyData []byte
	)
	// 产生副本
	copyData = make([]byte, len(p))
	copy(copyData, p)
	size = len(p)
	err = handler.wsConn.WsWrite(websocket.TextMessage, copyData)
	return
}

func WsHandler(c *gin.Context) {
	var (
		wsConn     *WsConnection
		restConf   *rest.Config
		sshReq     *rest.Request
		executor   remotecommand.Executor
		handler    *streamHandler
		err        error
		K8sCluster []databases.K8sCluster
	)
	// 得到websocket长连接
	if wsConn, err = InitWebsocket(c.Writer, c.Request); err != nil {
		Log.Error(err)
		return
	}
	userName := c.GetString("user_name")
	k8sId := c.Query("k8s_id")
	namespace := c.Query("namespace")
	podName := c.Query("pod")
	container := c.Query("container")
	Rows = c.Query("rows")
	Cols = c.Query("cols")
	subResource := c.Query("subResource")
	// 获取k8s rest client配置
	db.Where("k8s_id=?", k8sId).Find(&K8sCluster)
	if len(K8sCluster) > 0 {
		k8s := common.K8sCluster{Name: K8sCluster[0].K8sName, Config: K8sCluster[0].K8sConfig}
		ClientSet = k8s.Client()
		restConf = k8s.RestConfig()
		switch subResource {
		case "logs":
			lines := int64(500)
			sshReq = ClientSet.CoreV1().Pods(namespace).GetLogs(podName, &v1.PodLogOptions{Container: container, Follow: true, TailLines: &lines})
			stream, _ := sshReq.Stream(context.TODO())
			defer func(stream io.ReadCloser) {
				err = stream.Close()
			}(stream)
			buf := make([]byte, 2000)
			for {
				numBytes, err := stream.Read(buf)
				if numBytes == 0 || err == io.EOF {
					break // Exit on end of stream
				}
				err = wsConn.WsWrite(websocket.TextMessage, buf[:numBytes])
			}
		default:
			sshReq = ClientSet.CoreV1().RESTClient().Post().
				Resource("pods").
				Name(podName).
				Namespace(namespace).
				SubResource("exec").
				VersionedParams(&v1.PodExecOptions{
					Container: container,
					Command:   []string{"bash"},
					Stdin:     true,
					Stdout:    true,
					Stderr:    true,
					TTY:       true,
				}, scheme.ParameterCodec)
			// 创建到容器的连接
			if executor, err = remotecommand.NewSPDYExecutor(restConf, "POST", sshReq.URL()); err != nil {
				Log.Error(err)
				wsConn.WsClose()
			}
			userId := c.GetString("user_id")
			assetID := k8s.Name + "/" + namespace + "/" + podName
			// 配置与容器之间的数据流处理回调
			handler = &streamHandler{wsConn: wsConn, UserId: userId,
				k8s: assetID, resizeEvent: make(chan remotecommand.TerminalSize)}
			if userName == "guest" {
				wsConn.WsClose()
			} else {
				if err = executor.Stream(remotecommand.StreamOptions{
					Stdin:             handler,
					Stdout:            handler,
					Stderr:            handler,
					TerminalSizeQueue: handler,
					Tty:               true,
				}); err != nil {
					Log.Error(err)
					wsConn.WsClose()
				}
			}
		}
	}
	return
}
