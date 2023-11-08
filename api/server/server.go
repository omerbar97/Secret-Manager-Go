package main

import (
	"context"
	"golang-secret-manager/api/server/handler"
	"golang-secret-manager/api/server/middleware"
	"golang-secret-manager/utils/storage"
	"log"
	"net/http"
	"time"
)

type HttpServer struct {
	ctx  context.Context
	Addr string
}

func NewHttpServer(Addr string, ctx context.Context) *HttpServer {
	return &HttpServer{
		Addr: Addr,
		ctx:  ctx,
	}
}

func (s *HttpServer) Start() error {
	// Loading Routes
	http.HandleFunc("/secrets", func(w http.ResponseWriter, r *http.Request) {
		// applying middileware
		middleware.GetAllSecretsMiddleware(handler.MakeHTTPHandleFuncDecoder(handler.GetAllSecretsHandlers))(w, r.WithContext(s.ctx))
	})

	http.HandleFunc("/reports", func(w http.ResponseWriter, r *http.Request) {
		// applying middileware
	})

	log.Println("SERVER: Starting Server on port", s.Addr)
	return http.ListenAndServe(s.Addr, nil)
}

func main() {
	// Setting up the cache system
	// persist cache
	persistCache := storage.NewPersistCache("./persist-cache/")

	// FastCache
	cacheDuration := 5 * time.Minute
	fastCache := storage.NewFastCache(cacheDuration)

	fastCache.SetCacheLayer(persistCache, true)
	fastCache.ActivateLayerSavingRuntime(20 * time.Second)
	// End setting up cache system

	ctx := context.Background()

	httpServer := NewHttpServer(":8080", ctx)
	if err := httpServer.Start(); err != nil {
		log.Println("SERVER: failed to create server")
	}

}
