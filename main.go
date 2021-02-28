package main

import (
	"io"
	"log"
	"net/http"
	"time"

	"github.com/alexflint/go-arg"
	"github.com/getsentry/sentry-go"
	"github.com/gorilla/mux"
)

var (
	start    = time.Now()
	settings globalSettings
)

func main() {
	arg.MustParse(&settings)
	log.Printf("start server at http://%s", settings.Addr)

	if settings.SentryDsn != "" {
		err := sentry.Init(sentry.ClientOptions{
			Dsn: settings.SentryDsn,
		})
		if err != nil {
			log.Panicln("sentry fail:", err)
		}
	}

	if settings.Kozma {
		log.Println("kozma enabled")
		go startKozma()
	}

	server := &http.Server{
		Addr:         settings.Addr,
		Handler:      rootHandler(),
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  30 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		sentry.CaptureException(err)
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
		{"/api/rooms/{name:.+}/history", historyHandler},
		{"/api/rooms/{name:.+}", roomHandler},
	} {
		rtr.HandleFunc(item.path, wrap(item.fn))
	}
	return rtr
}

func wrap(fn func(w http.ResponseWriter, r *http.Request) error) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.Method, r.URL.Path)
		if err := fn(w, r); err != nil {
			sentry.CaptureException(err)
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
	<style>
	body { font-family: sans-serif; }
	b { display: block; }
	</style>
</head>
<body>
	<h2>Документация</h2>
 	<a href="https://github.com/tada-team/nane">https://github.com/tada-team/nane</a>	
	<h2>Проверка связи</h2>
	<div id="messages"></div>
	<script>
	let box = document.getElementById("messages");

	function connect() {
		let url = location.origin.replace(/^http/, "ws") + "/ws?name=Anonymous";
		let ws = new WebSocket(url);
		
		ws.addEventListener("open", function (event) {
			let body = document.createElement("p");
			body.appendChild(document.createTextNode("[есть контакт]"));
			box.appendChild(body);
		});
	
		ws.addEventListener("close", function (event) {
			let body = document.createElement("p");
			body.appendChild(document.createTextNode("[переподключение]"));
			box.appendChild(body);
			window.setTimeout(connect, 3000);
		});
	
		ws.addEventListener("message", function (event) {
			let msg = JSON.parse(event.data);
			
			let title = document.createElement("b");
			title.append(document.createTextNode(msg.sender.username));
			title.append(document.createTextNode(" @ "));
			title.append(document.createTextNode(msg.room));
	
			let body = document.createElement("p");
			body.appendChild(title);
			body.appendChild(document.createTextNode(msg.text));
			box.appendChild(body);
		});
	}

	connect();

	</script>
</body>
</html>`
	io.WriteString(w, index)
	return nil
}
