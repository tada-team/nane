package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
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
		"room",
		"44 %88 & me / 55",
		//"////", // fixme
		//"44 %88 & me / 55/history", // fixme
	} {
		t.Run("send message to " + roomName, func(t *testing.T) {
			message := nane.Message{
				Room: roomName,
				Text: "olo 0-0-8",
			}
			if err := ws.WriteJSON(message); err != nil {
				t.Fatalf("could not send message over ws connection %v", err)
			}

			t.Run("room info", func(t *testing.T) {
				v := new(struct {
					Result nane.Room `json:"result"`
					Error  string    `json:"error"`
				})

				if err := doGet(ts.URL+"/api/rooms/"+url.PathEscape(message.Room), v); err != nil {
					t.Fatal(err)
				}

				if v.Error != "" {
					t.Fatal(v.Error)
				}

				if v.Result.LastMessage == nil || v.Result.LastMessage.Text != message.Text {
					t.Error("invalid last message:", debugJSON(v))
				}
			})

			t.Run("room history", func(t *testing.T) {
				v := new(struct {
					Result []nane.Message `json:"result"`
					Error  string         `json:"error"`
				})

				if err := doGet(ts.URL+"/api/rooms/"+url.PathEscape(message.Room)+"/history", v); err != nil {
					t.Fatal(err)
				}

				if v.Error != "" {
					t.Fatal(v.Error)
				}

				if len(v.Result) != 1 {
					t.Error("invalid history:", debugJSON(v))
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

	respData, err := ioutil.ReadAll(resp.Body)
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
