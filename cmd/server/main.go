package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/zendgo/zendgo-api/internal/config"
	"github.com/zendgo/zendgo-api/internal/handler"
	"github.com/zendgo/zendgo-api/internal/middleware"
	"github.com/zendgo/zendgo-api/internal/repository"
	"github.com/zendgo/zendgo-api/internal/service"
	"github.com/zendgo/zendgo-api/routes"
)

func main() {
	log.Println("Starting ZendGo API...")

	cfg := config.Load()

	db, err := repository.NewDatabase(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("Connected to PostgreSQL database")

	if err := db.Migrate(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	log.Println("Database migrations completed")

	sessionRepo := repository.NewSessionRepository(db.DB)
	waSessionRepo := repository.NewWASessionRepository(db.DB)
	messageRepo := repository.NewMessageRepository(db.DB)

	whatsappService := service.NewWhatsAppService(db, sessionRepo, waSessionRepo, messageRepo, cfg)

	sessionHandler := handler.NewSessionHandler(whatsappService)
	messageHandler := handler.NewMessageHandler(whatsappService)

	authMiddleware := middleware.NewAuthMiddleware(whatsappService)
	loggingMiddleware := middleware.NewLoggingMiddleware()
	corsMiddleware := middleware.NewCORSMiddleware()

	routes := routes.NewRoutes(sessionHandler, messageHandler, authMiddleware, loggingMiddleware, corsMiddleware)
	handler := routes.Setup()

	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	serverErrors := make(chan error, 1)

	go func() {
		log.Printf("Server listening on %s:%s", cfg.Server.Host, cfg.Server.Port)
		serverErrors <- server.ListenAndServe()
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		log.Fatalf("Server error: %v", err)

	case sig := <-shutdown:
		log.Printf("Received signal %v, shutting down...", sig)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Fatalf("Server shutdown error: %v", err)
		}

		log.Println("Server stopped gracefully")
	}
}
