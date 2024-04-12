package whtest

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Websocket connection logic
var (
	wsConn *websocket.Conn
	connMu sync.Mutex
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func waitForConnection() (*websocket.Conn, error) {
	for {
		conn, err := getWebSocketConnection()
		if err != nil {
			return nil, err
		}
		if conn != nil {
			return conn, nil
		}
		time.Sleep(time.Millisecond * 100)
	}
}

func setWebSocketConnection(conn *websocket.Conn) {
	connMu.Lock()
	defer connMu.Unlock()
	wsConn = conn
}

func getWebSocketConnection() (*websocket.Conn, error) {
	connMu.Lock()
	defer connMu.Unlock()
	return wsConn, nil
}
