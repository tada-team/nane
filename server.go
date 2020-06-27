package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/tada-team/kozma"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true }, // XXX
}

type Session struct {
	conn *websocket.Conn
	name string
}

var connections = make(map[*websocket.Conn]Session)

func init() {
	go func() {
		for range time.Tick(60 * time.Second) {
			broadcast(textMessage(kozma.Name, kozma.Say()))
		}
	}()
}

func broadcast(v interface{}) {
	for conn := range connections {
		if err := conn.WriteJSON(v); err != nil {
			log.Println("write json fail:", err)
		}
	}
}

func rootHandler() http.Handler {
	rtr := mux.NewRouter()

	rtr.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/test.html")
	})

	rtr.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimSpace(r.URL.Query().Get("name"))
		if name == "" {
			sendFail(w, "name required")
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			sendFail(w, "upgrade fail:", err)
			return
		}

		connections[conn] = Session{
			conn: conn,
			name: name,
		}
		log.Println("+connection:", len(connections))

		defer func() {
			delete(connections, conn)
			log.Println("-connection:", len(connections))
			broadcast(systemMessage(fmt.Sprintf("ушёл: %s", name)))
		}()

		broadcast(systemMessage(fmt.Sprintf("пришёл: %s", name)))

		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				log.Println("read fail:", err)
				return
			}

			newMessage := new(Message)
			if err := json.Unmarshal(msg, newMessage); err != nil {
				log.Println("json fail:", err)
				return
			}

			if newMessage.Text != "" {
				broadcast(textMessage(name, newMessage.Text))
				broadcast(textMessage(kozma.Name, kozma.Say()))
			}
		}
	})

	return rtr
}

func sendFail(w http.ResponseWriter, v ...interface{}) {
	w.WriteHeader(500)
	io.WriteString(w, fmt.Sprintln(v...))
}
