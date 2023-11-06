package main

import (
	"context"
	"fmt"
	"golang-secret-manager/server/handlers"
	"golang-secret-manager/server/middlewares"
	"golang-secret-manager/server/storage"
	"log"
	"net/http"
	"time"
)

func main() {

	// Setting up the cache system
	// persist cache
	persistCache := storage.NewPersistCache("./persist-cache/")

	// FastCache
	cacheDuration := 5 * time.Minute
	storage.NewFastCache(cacheDuration).SetCacheLayer(persistCache, true)
	// End setting up cache system

	//  Server Port
	port := ":8080"

	ctx := context.Background()

	// Adding handler
	http.HandleFunc("/secrets", func(w http.ResponseWriter, r *http.Request) {
		// applying middileware
		middlewares.GetAllSecretsMiddleware(handlers.GetAllSecretsHandlers)(w, r.WithContext(ctx))
	})

	http.HandleFunc("/reports", func(http.ResponseWriter, *http.Request) {
		// applying middileware
	})

	fmt.Println("Starting Server on port", port)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		// failed to create the server
		log.Println("SERVER: failed to start the server")
	}
}
