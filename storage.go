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
	roomsMap map[string]*Room
	roomsMux sync.Mutex
	sessions = make(map[*websocket.Conn]Session)
)

func reset() {
	roomsMap = make(map[string]*Room)
}

func init() {
	reset()
}

type contentError string

func (e contentError) Error() string {
	return string(e)
}

func addMessage(sender nane.Sender, msg *nane.Message) error {
	msg.Text = truncateString(strings.TrimSpace(msg.Text), settings.GetMaxMessageLength())
	if msg.Text == "" {
		return contentError("text required")
	}

	msg.Room = truncateString(strings.TrimSpace(msg.Room), settings.GetMaxRoomTitleLength())
	if msg.Room == "" {
		return contentError("room required")
	}

	room := getOrCreateRoom(msg.Room)

	room.mux.Lock()
	defer room.mux.Unlock()

	msg.Room = room.Name
	msg.Created = time.Now()
	msg.Sender = sender

	room.LastMessage = msg
	room.messages.Value = *msg
	room.messages = room.messages.Next()

	return nil
}

func (room *Room) getMessages() []nane.Message {
	var result []nane.Message
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
	result := make([]nane.Room, 0, len(roomsMap))
	for _, room := range roomsMap {
		result = append(result, room.Room)
	}
	return result
}

func getRoom(name string) *Room {
	return roomsMap[normalizeRoomName(name)]
}

func getOrCreateRoom(name string) *Room {
	roomsMux.Lock()
	defer roomsMux.Unlock()

	room := getRoom(name)
	if room == nil {
		room = &Room{
			Room:     nane.Room{Name: name},
			messages: ring.New(settings.GetMaxMessagesInRoom()),
			mux:      new(sync.Mutex),
		}
		roomsMap[normalizeRoomName(name)] = room
	}

	return room
}
