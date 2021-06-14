package main

import (
	"bytes"
	"encoding/json"
	"log"

	"github.com/pkg/errors"
)

func debugJSON(v interface{}) string {
	b := new(bytes.Buffer)
	debugEncoder := json.NewEncoder(b)
	debugEncoder.SetIndent("", "    ")
	debugEncoder.SetEscapeHTML(false)
	if err := debugEncoder.Encode(v); err != nil {
		log.Panicln(errors.Wrap(err, "json marshall fail"))
	}
	return b.String()
}
