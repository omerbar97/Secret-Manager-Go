package main

import (
	"context"
	"golang-secret-manager/api/server/handler"
	"golang-secret-manager/api/server/middleware"
	"golang-secret-manager/utils/storage"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type HttpServer struct {
	ctx    context.Context
	server *http.Server
}

func NewHttpServer(addr string, ctx context.Context) *HttpServer {
	return &HttpServer{
		ctx: ctx,
		server: &http.Server{
			Addr: addr,
		},
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
		middleware.GetReportMiddleware(handler.MakeHTTPHandleFuncDecoder(handler.GetReportsHandler))(w, r.WithContext(s.ctx))
	})

	log.Println("SERVER: Starting Server on port", s.server.Addr)
	return http.ListenAndServe(s.server.Addr, nil)
}

func (s *HttpServer) ShutDown() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		log.Printf("SERVER: Error shutting down server: %s", err)
	} else {
		log.Println("SERVER: Server gracefully stopped")
	}
}

func main() {
	// Setting up the cache system
	// persist cache
	persistCache := storage.NewPersistCache("./persist-cache/")

	// FastCache
	cacheDuration := 5 * time.Minute
	fastCache := storage.NewFastCache(cacheDuration)

	if err := fastCache.SetCacheLayer(persistCache, true); err != nil {
		log.Fatalln("failed to init fast cache")
	}
	fastCache.ActivateLayerSavingRuntime(20 * time.Second)
	// End setting up cache system

	ctx := context.Background()

	httpServer := NewHttpServer(":8080", ctx)

	// Thread that handle the Ctrl + C signal
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT)
	go func() {
		<-ch
		// exiting program
		log.Println("Shuting down server...")
		httpServer.ShutDown()
	}()

	if err := httpServer.Start(); err != nil {
		log.Println("SERVER: failed to create server")
	}

}
