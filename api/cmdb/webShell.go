package cmdb

import (
	"bytes"
	"github.com/duke-git/lancet/convertor"
	"github.com/duke-git/lancet/v2/netutil"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"golang.org/x/crypto/ssh"
	"inner/conf/platform_conf"
	"inner/modules/databases"
	"inner/modules/kits"
	"io"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"
)

var (
	upGrader = websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024 * 1024 * 10,
		CheckOrigin: func(r *http.Request) bool {
			return true
		}}
	Rows    = 850
	Cols    = 185
	Encrypt = kits.NewEncrypt([]byte(platform_conf.CryptKey), 16)
)

type wsBufferWriter struct {
	buffer bytes.Buffer
	mu     sync.Mutex
}
type SshConn struct {
	// calling Write
	StdinPipe io.WriteCloser
	// Write() be called to receive data from ssh server
	ComboOutput *wsBufferWriter
	Session     *ssh.Session
}

// implement Write
func (w *wsBufferWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.buffer.Write(p)
}

// flushComboOutput flush ssh.session
func flushComboOutput(w *wsBufferWriter, wsConn *websocket.Conn) error {
	if w.buffer.Len() != 0 {
		err := wsConn.WriteMessage(websocket.TextMessage, w.buffer.Bytes())
		if err != nil {
			return err
		}
		w.buffer.Reset()
	}
	return nil
}

func NewSshConn(cols, rows int, sshClient *ssh.Client) (*SshConn, error) {
	sshSession, err := sshClient.NewSession()
	if err != nil {
		return nil, err
	}
	stdinP, err := sshSession.StdinPipe()
	if err != nil {
		return nil, err
	}
	comboWriter := new(wsBufferWriter)
	//ssh.stdout
	sshSession.Stdout = comboWriter
	sshSession.Stderr = comboWriter
	modes := ssh.TerminalModes{
		ssh.ECHO:          1,     // disable echo
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4
	}
	// Request pseudo terminal
	if err := sshSession.RequestPty("xterm", rows, cols, modes); err != nil {
		return nil, err
	}
	// Start remote shell
	if err := sshSession.Shell(); err != nil {
		return nil, err
	}
	return &SshConn{StdinPipe: stdinP, ComboOutput: comboWriter, Session: sshSession}, nil
}

func (ss *SshConn) Close() {
	if ss.Session != nil {
		_ = ss.Session.Close()
	}

}

// ReceiveWsMsg  receive websocket msg
func (Conn *SshConn) ReceiveWsMsg(wsConn *websocket.Conn, exitCh chan bool, UserId, HostId string) {
	//tells other go routine quit
	defer setQuit(exitCh)
	for {
		select {
		case <-exitCh:
			return
		default:
			//read websocket msg
			_, wsData, err := wsConn.ReadMessage()
			if err != nil {
				Log.Error("reading webSocket message failed")
				return
			}
			//handle xterm.js stdin
			decodeBytes := wsData
			if err != nil {
				Log.Error("websock cmd string base64 decoding failed")
			}
			if _, err := Conn.StdinPipe.Write(decodeBytes); err != nil {
				Log.Error("ws cmd bytes write to ssh.stdin pipe failed")
			}
			if err := Conn.Session.WindowChange(Rows, Cols); err != nil {
				Log.Error("ssh pty change windows size failed")
			}
		}
	}
}
func (Conn *SshConn) SendComboOutput(wsConn *websocket.Conn, exitCh chan bool) {
	//tells other go routine quit
	defer setQuit(exitCh)
	//every 120ms
	tick := time.NewTicker(time.Millisecond * time.Duration(120))
	//for range time.Tick(120 * time.Millisecond){}
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			//write combine
			if err := flushComboOutput(Conn.ComboOutput, wsConn); err != nil {
				Log.Error("ssh sending combo output to webSocket failed")
				return
			}
		case <-exitCh:
			return
		}
	}
}

func (Conn *SshConn) SessionWait(quitChan chan bool) {
	if err := Conn.Session.Wait(); err != nil {
		Log.Error("ssh session wait failed")
		setQuit(quitChan)
	}
}

func setQuit(ch chan bool) {
	ch <- true
}

func NewSshClient(User, Host, Key, Password string, Port int) (*ssh.Client, error) {
	config := &ssh.ClientConfig{
		Timeout:         time.Second * 5,
		User:            User,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), //这个可以， 但是不够安全
	}
	config.Auth = []ssh.AuthMethod{ssh.Password(Password)}
	signer, err := ssh.ParsePrivateKey([]byte(Key))
	if err == nil {
		config.Auth = append(config.Auth, ssh.PublicKeys(signer))
	}
	c, err := ssh.Dial("tcp", Host+":"+strconv.Itoa(Port), config)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func WsHandler(c *gin.Context) {
	var (
		AssetServer     []databases.AssetServer
		AssetNet        []databases.AssetNet
		AssetServerPool []databases.AssetServerPool
		AssetSwitch     []databases.AssetSwitch
		AssetSwitchPool []databases.AssetSwitchPool
		SshKey          []databases.SshKey
		host            string
		sshUser         string
		sshPort         = 22
		passwd          []byte
		pkey            string
	)
	userName := c.GetString("user_name")
	if userName != "" && userName != "guest" {
		// 得到websocket长连接
		AssetId := c.Query("asset_id")
		AssetType := c.Query("asset_type")
		R, _ := convertor.ToInt(c.Query("rows"))
		C, _ := convertor.ToInt(c.Query("cols"))
		// 获取k8s rest client配置
		switch AssetType {
		case "server":
			db.Where("host_id=?", AssetId).Find(&AssetServer)
			if len(AssetServer) > 0 {
				db.Where("id=?", AssetServer[0].PoolId).Find(&AssetServerPool)
				db.Where("host_id=?", AssetId).Find(&AssetNet)
				if len(AssetNet) > 0 {
					for _, n := range AssetNet {
						if netutil.IsInternalIP(net.ParseIP(n.Ip)) {
							host = n.Ip
							break
						}
					}
				}
				if len(AssetServerPool) > 0 && host != "" {
					sshUser = AssetServerPool[0].SshUser
					sshPort = AssetServerPool[0].SshPort
					passwd, _ = Encrypt.DecryptString(AssetServerPool[0].SshPassword, true)
					db.Where("key_name=?", AssetServerPool[0].SshKeyName).Find(&SshKey)
					if len(SshKey) > 0 {
						k, err := Encrypt.DecryptString(SshKey[0].SshKey, true)
						if err == nil {
							pkey = string(k)
						}
					}
				}
			}
		case "switch":
			db.Where("switch_id=?", AssetId).Find(&AssetSwitch)
			if len(AssetSwitch) > 0 {
				host = AssetSwitch[0].SwitchIp
				db.Where("id=?", AssetSwitch[0].SwitchPoolId).Find(&AssetSwitchPool)
				if len(AssetSwitchPool) > 0 && host != "" {
					sshUser = AssetSwitchPool[0].SwitchUser
					sshPort = AssetSwitchPool[0].SwitchPort
					passwd, _ = Encrypt.DecryptString(AssetSwitchPool[0].SwitchPassword, true)
				}
			}
		}
		if sshUser != "" && passwd != nil && host != "" {
			wsConn, err := upGrader.Upgrade(c.Writer, c.Request, nil)
			defer func(wsConn *websocket.Conn) {
				_ = wsConn.Close()
			}(wsConn)
			if err == nil {
				client, err := NewSshClient(sshUser, host, pkey, string(passwd), sshPort)
				if err == nil {
					defer func(client *ssh.Client) {
						_ = client.Close()
					}(client)
					if R > 0 && C > 0 {
						ssConn, err := NewSshConn(int(C), int(R), client)
						if err == nil {
							defer ssConn.Close()
							quitChan := make(chan bool, 3)
							go ssConn.ReceiveWsMsg(wsConn, quitChan, c.GetString("user_id"), AssetId)
							go ssConn.SendComboOutput(wsConn, quitChan)
							go ssConn.SessionWait(quitChan)
							<-quitChan
						}
					}
				}
			}
		}
	}
	return
}
