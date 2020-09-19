package nane

import (
	"time"
)

type Settings struct {
	MaxMessageLength   int           `json:"max_message_length"`
	MaxRoomTitleLength int           `json:"max_room_title_length"`
	MaxUsernameLength  int           `json:"max_username_length"`
	Uptime             time.Duration `json:"uptime"`
}

type Sender struct {
	Username string `json:"username"`
}

type Room struct {
	Name        string   `json:"name"`
	LastMessage *Message `json:"last_message"`
}

type Message struct {
	Id      string    `json:"id,omitempty"`
	Room    string    `json:"room"`
	Created time.Time `json:"created,omitempty"`
	Sender  Sender    `json:"sender"`
	Text    string    `json:"text"`
}

type ApiResponse struct {
	Result interface{} `json:"result,omitempty"`
	Error  string      `json:"error,omitempty"`
}
