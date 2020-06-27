package main

import (
	"log"
	"net/http"
	"time"

	"github.com/alexflint/go-arg"
)

func main() {
	var args struct {
		Addr string `help:"address"`
	}
	arg.MustParse(&args)

	addr := args.Addr
	if addr == "" {
		addr = "localhost:8082"
	}
	log.Println("start server:", addr)

	server := &http.Server{
		Addr:         addr,
		Handler:      rootHandler(),
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  30 * time.Second,
	}
	if err := server.ListenAndServe(); err != nil {
		log.Panicln("listen fail:", err)
	}
}
