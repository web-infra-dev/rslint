package main

import (
	"encoding/json"
	"log"
	"os"
)

func main() {
	entries := collectRuleSchemas()
	if err := json.NewEncoder(os.Stdout).Encode(entries); err != nil {
		log.Fatal(err)
	}
}
