package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/tada-team/nane/nane"
)

func TestRootHandler(t *testing.T) {
	srv := http.NewServeMux()
	srv.Handle("/", rootHandler())

	ts := httptest.NewServer(srv)
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws?name=tester"

	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("could not open a ws connection on %s %v", wsURL, err)
	}
	defer ws.Close()

	message := nane.Message{
		Room: "test room",
		Text: "333",
	}
	if err := ws.WriteJSON(message); err != nil {
		t.Fatalf("could not send message over ws connection %v", err)
	}
}
