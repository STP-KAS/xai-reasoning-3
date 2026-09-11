package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	cfg := loadConfig()
	store, err := openStore(cfg.DataDir)
	if err != nil {
		log.Fatal(err)
	}
	s := newServer(cfg, store)
	log.Printf("desk %s listening on http://%s (pid %d)", cfg.Name, cfg.Bind, os.Getpid())
	log.Printf("does not hold funds. 402 refuses unverified kaspa txids.")
	if err := http.ListenAndServe(cfg.Bind, s.routes()); err != nil {
		log.Fatal(err)
	}
}
