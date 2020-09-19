package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/tada-team/nane/nane"
)

func settingsHandler(w http.ResponseWriter, r *http.Request) error {
	return jsonResponse(w, nane.ApiResponse{
		Result: nane.Settings{
			MaxMessageLength:   settings.MaxMessageLength,
			MaxRoomTitleLength: settings.MaxRoomTitleLength,
			MaxUsernameLength:  settings.MaxUsernameLength,
			Uptime:             time.Since(start),
		},
	})
}

func roomsHandler(w http.ResponseWriter, r *http.Request) error {
	return jsonResponse(w, nane.ApiResponse{
		Result: getRooms(),
	})
}

func roomHandler(w http.ResponseWriter, r *http.Request) error {
	name := mux.Vars(r)["name"]
	room := getRoom(name)
	if room == nil {
		w.WriteHeader(http.StatusNotFound)
		return jsonResponse(w, nane.ApiResponse{
			Error: fmt.Sprintf("room %s not found", name),
		})
	}
	return jsonResponse(w, nane.ApiResponse{
		Result: room,
	})
}

func historyHandler(w http.ResponseWriter, r *http.Request) error {
	name := mux.Vars(r)["name"]
	room := getRoom(name)
	if room == nil {
		w.WriteHeader(http.StatusNotFound)
		return jsonResponse(w, nane.ApiResponse{
			Error: fmt.Sprintf("room %s not found", name),
		})
	}
	return jsonResponse(w, nane.ApiResponse{
		Result: room.getMessages(),
	})
}

func jsonResponse(w http.ResponseWriter, v nane.ApiResponse) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	encoder := json.NewEncoder(w)
	if err := encoder.Encode(v); err != nil {
		return errors.Wrap(err, "encode fail")
	}
	return nil
}
