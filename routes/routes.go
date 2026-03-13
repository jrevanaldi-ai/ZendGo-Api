package routes

import (
	"net/http"

	"github.com/zendgo/zendgo-api/internal/handler"
	"github.com/zendgo/zendgo-api/internal/middleware"
)

type Routes struct {
	sessionHandler  *handler.SessionHandler
	messageHandler  *handler.MessageHandler
	authMiddleware  *middleware.AuthMiddleware
	loggingMiddleware *middleware.LoggingMiddleware
	corsMiddleware  *middleware.CORSMiddleware
}

func NewRoutes(
	sessionHandler *handler.SessionHandler,
	messageHandler *handler.MessageHandler,
	authMiddleware *middleware.AuthMiddleware,
	loggingMiddleware *middleware.LoggingMiddleware,
	corsMiddleware *middleware.CORSMiddleware,
) *Routes {
	return &Routes{
		sessionHandler:  sessionHandler,
		messageHandler:  messageHandler,
		authMiddleware:  authMiddleware,
		loggingMiddleware: loggingMiddleware,
		corsMiddleware:  corsMiddleware,
	}
}

func (r *Routes) Setup() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success":true,"message":"ZendGo API is running"}`))
	})

	mux.HandleFunc("/api/v1/sessions/new", r.sessionHandler.CreateSession)
	mux.HandleFunc("/api/v1/sessions", r.getAllSessions)
	mux.HandleFunc("/api/v1/sessions/qr", r.sessionHandler.GetQRCode)
	mux.HandleFunc("/api/v1/sessions/pair", r.sessionHandler.PairPhone)
	mux.HandleFunc("/api/v1/sessions/delete", r.sessionHandler.DeleteSession)

	protectedMux := http.NewServeMux()
	protectedMux.HandleFunc("/api/v1/messages/text", r.messageHandler.SendTextMessage)
	protectedMux.HandleFunc("/api/v1/messages/image", r.messageHandler.SendImageMessage)
	protectedMux.HandleFunc("/api/v1/messages/document", r.messageHandler.SendDocumentMessage)
	protectedMux.HandleFunc("/api/v1/messages/audio", r.messageHandler.SendAudioMessage)
	protectedMux.HandleFunc("/api/v1/messages/video", r.messageHandler.SendVideoMessage)
	protectedMux.HandleFunc("/api/v1/messages/location", r.messageHandler.SendLocationMessage)
	protectedMux.HandleFunc("/api/v1/messages/contact", r.messageHandler.SendContactMessage)
	protectedMux.HandleFunc("/api/v1/messages/cta-button", r.messageHandler.SendCTAButtonMessage)
	protectedMux.HandleFunc("/api/v1/groups", r.messageHandler.CreateGroup)
	protectedMux.HandleFunc("/api/v1/groups/list", r.messageHandler.ListGroups)

	mux.Handle("/api/v1/messages/", r.authMiddleware.Middleware(protectedMux))
	mux.Handle("/api/v1/groups/", r.authMiddleware.Middleware(protectedMux))

	handler := r.loggingMiddleware.Middleware(mux)
	handler = r.corsMiddleware.Middleware(handler)

	return handler
}

func (r *Routes) getAllSessions(w http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodGet {
		r.sessionHandler.GetAllSessions(w, req)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
