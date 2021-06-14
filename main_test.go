package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"

	"github.com/tada-team/nane/nane"
)

func TestRootHandler(t *testing.T) {
	reset()

	srv := http.NewServeMux()
	srv.Handle("/", rootHandler())

	ts := httptest.NewServer(srv)
	defer ts.Close()

	username := "tester #1"
	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws?name=" + url.QueryEscape(username)

	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("could not open a ws connection on %s %v", wsURL, err)
	}
	defer ws.Close()

	t.Run("ping-pong", func(t *testing.T) {
		if err := ws.WriteJSON(nane.Ping{Ping: true}); err != nil {
			t.Fatalf("could not send message over ws connection %v", err)
		}
		_, msg, err := ws.ReadMessage()
		if err != nil {
			t.Fatal(err)
		}
		pong := new(nane.Pong)
		if err := json.Unmarshal(msg, &pong); err != nil {
			t.Fatal(err)
		}
		if !pong.Pong {
			t.Fatal("invalid pong")
		}
	})

	t.Run("invalid message", func(t *testing.T) {
		message := map[string]string{"xx": "123"}
		if err := ws.WriteJSON(message); err != nil {
			t.Fatalf("could not send message over ws connection %v", err)
		}
	})

	for _, roomName := range []string{
		"room #1",
		"another room",
		"44 %88 & me / 55",
		//"////", // fixme
		//"44 %88 & me / 55/history", // fixme
	} {
		t.Run("message to "+roomName, func(t *testing.T) {
			message := nane.Message{
				Room: roomName,
				Text: "message to " + roomName,
			}

			if err := ws.WriteJSON(message); err != nil {
				t.Fatalf("could not send message over ws connection %v", err)
			}

			time.Sleep(25 * time.Millisecond) // XXX

			t.Run("info", func(t *testing.T) {
				resp := new(struct {
					Result nane.Room `json:"result"`
					Error  string    `json:"error"`
				})

				if err := doGet(ts.URL+"/api/rooms/"+url.PathEscape(roomName), resp); err != nil {
					t.Fatal(err)
				}

				if resp.Error != "" {
					t.Fatal(resp.Error)
				}

				if resp.Result.Name != roomName {
					t.Error("invalid room name:", resp.Result.Name, "want:", roomName)
				}

				if resp.Result.LastMessage == nil || resp.Result.LastMessage.Text != message.Text || resp.Result.LastMessage.Sender.Username != username {
					t.Error("invalid last message:", debugJSON(resp))
				}
			})

			t.Run("history", func(t *testing.T) {
				resp := new(struct {
					Result []nane.Message `json:"result"`
					Error  string         `json:"error"`
				})

				if err := doGet(ts.URL+"/api/rooms/"+url.PathEscape(roomName)+"/history", resp); err != nil {
					t.Fatal(err)
				}

				if resp.Error != "" {
					t.Fatal(resp.Error)
				}

				if len(resp.Result) != 1 {
					t.Error("invalid history: want 1 message, got:", debugJSON(resp))
				}

				if resp.Result[0].Sender.Username != username {
					t.Error("invalid sender username:", resp.Result[0].Sender.Username, "want:", username)
				}
			})
		})
	}
}

func doGet(path string, v interface{}) error {
	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		return errors.Wrap(err, "new request fail")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "client do fail")
	}
	defer resp.Body.Close()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "read body fail")
	}

	if resp.StatusCode != 200 {
		return errors.Wrapf(err, "status code: %d %s", resp.StatusCode, string(respData))
	}

	if err := json.Unmarshal(respData, &v); err != nil {
		return errors.Wrapf(err, "unmarshal fail on: %s", string(respData))
	}

	return nil
}
