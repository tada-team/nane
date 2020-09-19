package main

import (
	"io"
	"log"
	"net/http"
	"time"

	"github.com/alexflint/go-arg"
	"github.com/gorilla/mux"
)

var start = time.Now()

var settings struct {
	Addr               string `default:"localhost:8022"`
	MaxMessagesInRoom  int    `default:"1024"`
	MaxRoomTitleLength int    `default:"50"`
	MaxUsernameLength  int    `default:"50"`
	MaxMessageLength   int    `default:"10500"`
}

func main() {
	arg.MustParse(&settings)
	log.Printf("start server at http://%s", settings.Addr)

	server := &http.Server{
		Addr:         settings.Addr,
		Handler:      rootHandler(),
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  30 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Panicln("listen fail:", err)
	}
}

func rootHandler() http.Handler {
	rtr := mux.NewRouter()
	for _, item := range []struct {
		path string
		fn   func(http.ResponseWriter, *http.Request) error
	}{
		{"/", indexHandler},
		{"/ws", wsHandler},
		{"/api/settings", settingsHandler},
		{"/api/rooms", roomsHandler},
		{"/api/rooms/{name}", roomHandler},
		{"/api/rooms/{name}/history", historyHandler},
	} {
		rtr.HandleFunc(item.path, wrap(item.fn))
	}
	return rtr
}

func wrap(fn func(w http.ResponseWriter, r *http.Request) error) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.Method, r.URL.Path)
		err := fn(w, r)
		if err != nil {
			log.Println("fail:", err)
		}
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) error {
	const index = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Ай-нанэ-нанэ!</title>
</head>
<body>
    Docs: <a href="https://github.com/tada-team/nane">https://github.com/tada-team/nane</a>
</body>
</html>`
	io.WriteString(w, index)
	return nil
}
