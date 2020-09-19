package main

import (
	"log"
	"math/rand"
	"time"

	"github.com/tada-team/kozma"
	"github.com/tada-team/nane/nane"
)

func init() {
	if settings.Kozma {
		go func() {
			time.Sleep(3 * time.Second)
			for {
				sender := nane.Sender{Username: kozma.Name}
				err := addMessage(sender, &nane.Message{
					Room: "kozma",
					Text: kozma.Say(),
				})
				if err != nil {
					log.Panicln("kozma fail:", err)
				}
				time.Sleep(time.Duration(rand.Intn(60)) * time.Second)
			}
		}()
	}
}
