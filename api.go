package main

import "time"

type Message struct {
	Created time.Time `json:"created,omitempty"`
	Name    string    `json:"name,omitempty"`
	Text    string    `json:"text"`
}

func textMessage(name, text string) Message {
	return Message{
		Created: time.Now(),
		Name:    name,
		Text:    text,
	}
}

func systemMessage(text string) Message {
	return Message{
		Created: time.Now(),
		Text:    text,
	}
}
