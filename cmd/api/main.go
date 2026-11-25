package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nabilfikrisp/cicd-demo/internal/httpserver"
	"github.com/nabilfikrisp/cicd-demo/internal/modules/service/handler"
	"github.com/nabilfikrisp/cicd-demo/internal/modules/service/repository/memory"
	"github.com/nabilfikrisp/cicd-demo/internal/modules/service/usecase"
)

var version = "dev"
var commit = "unknown"

func main() {
	log.SetOutput(os.Stdout)
	server := newServer()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	run(ctx, server)
}

func newServer() *http.Server {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3003"
	}
	serviceRepository := memory.NewServiceRepository()
	serviceUsecase := usecase.NewServiceUsecase(serviceRepository)
	serviceHandler := handler.NewServiceHandler(serviceUsecase)
	router := httpserver.NewRouter(serviceHandler, version, commit)

	return &http.Server{
		Addr:              ":" + port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
	}
}

func run(ctx context.Context, server *http.Server) {
	go func() {
		log.Printf("listening on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server: %v", err)
		}
	}()
	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}
