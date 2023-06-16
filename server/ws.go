package server

import (
	"log"
	"net/http"

	"github.com/94d/goquiz/entity"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

func (s *server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	usr, err := getUser(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading to WebSocket:", err)
		return
	}
	defer func() {
		conn.Close()
		entity.Onlines.Remove(usr.ID)
	}()

	for {
		entity.Onlines.Add(usr.ID)

		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}
