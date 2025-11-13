package main

import (
	"log"
	"os"
)

func main() {
    addr := os.Getenv("SERVER_ADDR")
    if addr == "" {
        addr = ":9000"
    }
    storageFile := os.Getenv("STORAGE_FILE")
    if storageFile == "" {
        storageFile = "./data/logs.jsonl"
    }
    store, err := NewStore(storageFile)
    if err != nil {
        log.Fatalf("store init: %v", err)
    }
    defer store.Close()

    r := setupRouter(store)
    log.Printf("log-server listening %s, storage=%s", addr, storageFile)
    if err := r.Run(addr); err != nil {
        log.Fatalf("server exit: %v", err)
    }
}