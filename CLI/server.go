package CLI

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var num int

func SetupRouter(port int, route string, webhook string) {
	http.HandleFunc("/cli", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		ForwardDataHandler(w, r, port, route)
	})

	http.HandleFunc("/url", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		websocket, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}

		TestURLHandler(websocket, port, route)
	})

	fmt.Println("CLI has successfully connected with your local app")
	fmt.Println("CLI is hosted at port :3000")
	http.ListenAndServe(":3000", nil)
}

func WebhookTransfer(w http.ResponseWriter, r *http.Request, webhook string) {

}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}
