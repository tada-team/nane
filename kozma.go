package main

import (
	"log"
	"math/rand"
	"time"

	"github.com/getsentry/sentry-go"

	"github.com/tada-team/kozma"
	"github.com/tada-team/nane/nane"
)

func startKozma() {
	time.Sleep(1 * time.Second) // wait for ws
	for {
		sender := nane.Sender{Username: kozma.Name}
		err := addMessage(sender, &nane.Message{
			Room: "kozma",
			Text: kozma.Say(),
		})
		if err != nil {
			sentry.CaptureException(err)
			log.Panicln("kozma fail:", err)
		}
		time.Sleep(time.Duration(rand.Intn(60)) * time.Second)
	}
}
