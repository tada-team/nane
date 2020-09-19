package main

import (
	"container/ring"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/tada-team/nane/nane"
)

type Session struct {
	nane.Sender
	conn *websocket.Conn
}

type Room struct {
	nane.Room
	messages *ring.Ring
	mux      *sync.Mutex
}

var (
	sessions = make(map[*websocket.Conn]Session)
	rooms    = make(map[string]*Room)
	roomsMux = new(sync.Mutex)
)

type contentError string

func (e contentError) Error() string {
	return string(e)
}

func addMessage(sender nane.Sender, msg *nane.Message) error {
	msg.Text = truncateString(strings.TrimSpace(msg.Text), settings.MaxMessageLength)
	if msg.Text == "" {
		return contentError("text required")
	}

	msg.Room = truncateString(strings.TrimSpace(msg.Room), settings.MaxRoomTitleLength)
	if msg.Room == "" {
		return contentError("room required")
	}

	room := getOrCreateRoom(msg.Room)

	room.mux.Lock()
	defer room.mux.Unlock()

	msg.Room = room.Name
	msg.Created = time.Now()
	msg.Sender = sender
	msg.Text = truncateString(msg.Text, settings.MaxMessageLength)

	room.LastMessage = msg
	room.messages.Value = *msg
	room.messages = room.messages.Next()

	return nil
}

func (room *Room) getMessages() []nane.Message {
	result := make([]nane.Message, 0)
	room.messages.Do(func(i interface{}) {
		if i != nil {
			result = append(result, i.(nane.Message))
		}
	})
	sort.Slice(result, func(i, j int) bool {
		return result[i].Created.Before(result[j].Created)
	})
	return result
}

func getRooms() []nane.Room {
	result := make([]nane.Room, 0, len(rooms))
	for _, room := range rooms {
		result = append(result, room.Room)
	}
	return result
}

func getRoom(name string) *Room {
	return rooms[strings.ToLower(name)]
}

func getOrCreateRoom(name string) *Room {
	roomsMux.Lock()
	defer roomsMux.Unlock()

	room := getRoom(name)
	if room == nil {
		room = &Room{
			Room:     nane.Room{Name: name},
			messages: ring.New(settings.MaxMessagesInRoom),
			mux:      new(sync.Mutex),
		}
		rooms[strings.ToLower(name)] = room
	}

	return room
}

func truncateString(s string, maxLength int) string {
	r := []rune(s)
	if len(r) <= maxLength {
		return s
	}
	return string(r[:maxLength])
}
