package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/tada-team/nane/nane"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true }, // XXX
}

func broadcast(v *nane.Message) {
	for conn := range sessions {
		if err := conn.WriteJSON(v); err != nil {
			log.Println("write json fail:", err)
		}
	}
}

func wsHandler(w http.ResponseWriter, r *http.Request) error {
	username := strings.TrimSpace(r.URL.Query().Get("username"))
	if username == "" {
		username = strings.TrimSpace(r.URL.Query().Get("name")) // backward compatibility
	}

	if username == "" {
		w.WriteHeader(http.StatusUnauthorized)
		io.WriteString(w, "name required")
		return nil
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		w.WriteHeader(http.StatusUpgradeRequired)
		io.WriteString(w, "upgrade failed")
		return nil
	}

	session := Session{conn: conn}
	session.Username = username
	sessions[conn] = session
	log.Println("+connection:", len(sessions))

	defer func() {
		delete(sessions, conn)
		log.Println("-connection:", len(sessions))
	}()

	for {
		messageType, msg, err := conn.ReadMessage()
		if err != nil {
			return err
		}

		switch messageType {
		case websocket.CloseMessage:
			return nil
		case websocket.TextMessage, websocket.BinaryMessage:
			ping := new(nane.Ping)
			if err := json.Unmarshal(msg, &ping); err != nil {
				return err
			}

			if ping.Ping {
				log.Println("got ping from", username)
				if err := conn.WriteJSON(nane.Pong{Pong: true}); err != nil {
					return err
				}
				continue
			}

			message := new(nane.Message)
			if err := json.Unmarshal(msg, &message); err != nil {
				return err
			}

			log.Println("got message from", username, strings.TrimSpace(string(msg)))
			if err := addMessage(session.Sender, message); err != nil {
				_, ok := err.(contentError)
				if ok {
					log.Println("warn:", err)
					continue
				}
				return err
			}

			broadcast(message)
			message.Id = ""
		}
	}
}
