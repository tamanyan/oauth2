package main

import (
	"log"
	"os"
	"time"

	"github.com/tamanyan/oauth2"
	"github.com/tamanyan/oauth2/models"
	"github.com/tamanyan/oauth2/store""
)


func main() {
	store, err := store.NewTokenStore(&store.TokenConfig{
		DSN:         ":memory:",
		DBType:      "sqlite3",
		TableName:   "oauth2_token",
		MaxLifetime: time.Second * 1,
	}, 600)
	time.Sleep(time.Second * 5)
	log.Println(store)
	store.Close()
}
