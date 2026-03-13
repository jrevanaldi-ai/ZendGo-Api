package handler

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/zendgo/zendgo-api/internal/service"
	"github.com/zendgo/zendgo-api/models"
)

type SessionHandler struct {
	service *service.WhatsAppService
}

func NewSessionHandler(service *service.WhatsAppService) *SessionHandler {
	return &SessionHandler{service: service}
}

type CreateSessionRequest struct {
	Phone      string `json:"phone"`
	WebhookURL string `json:"webhook_url,omitempty"`
}

type CreateSessionResponse struct {
	Success bool             `json:"success"`
	Data    *models.Session  `json:"data"`
	Message string           `json:"message,omitempty"`
}

type SessionResponse struct {
	Success bool             `json:"success"`
	Data    *models.Session  `json:"data"`
	Message string           `json:"message,omitempty"`
}

type SessionsResponse struct {
	Success bool              `json:"success"`
	Data    []*models.Session `json:"data"`
	Message string            `json:"message,omitempty"`
}

type QRResponse struct {
	Success bool   `json:"success"`
	Data    string `json:"data"`
	Message string `json:"message,omitempty"`
}

type PairRequest struct {
	Phone         string `json:"phone"`
	HasCountryCode *bool  `json:"has_country_code,omitempty"`
	ClientType    string `json:"client_type,omitempty"`
	DeviceName     string `json:"device_name,omitempty"`
}

type PairResponse struct {
	Success bool   `json:"success"`
	Data    string `json:"data"`
	Message string `json:"message,omitempty"`
}

func (h *SessionHandler) CreateSession(w http.ResponseWriter, r *http.Request) {
	var req CreateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"success":false,"message":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Phone == "" {
		http.Error(w, `{"success":false,"message":"Phone number is required"}`, http.StatusBadRequest)
		return
	}

	session, err := h.service.CreateSession(r.Context(), req.Phone, req.WebhookURL)
	if err != nil {
		http.Error(w, jsonEncode(&CreateSessionResponse{Success: false, Message: err.Error()}), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(&CreateSessionResponse{
		Success: true,
		Data:    session,
		Message: "Session created successfully. Scan QR code to connect.",
	})
}

func (h *SessionHandler) GetSession(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, `{"success":false,"message":"Session ID is required"}`, http.StatusBadRequest)
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, `{"success":false,"message":"Invalid session ID"}`, http.StatusBadRequest)
		return
	}

	session, err := h.service.GetSession(r.Context(), id)
	if err != nil {
		http.Error(w, jsonEncode(&SessionResponse{Success: false, Message: err.Error()}), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&SessionResponse{
		Success: true,
		Data:    session,
	})
}

func (h *SessionHandler) GetAllSessions(w http.ResponseWriter, r *http.Request) {
	sessions, err := h.service.GetAllSessions(r.Context())
	if err != nil {
		http.Error(w, jsonEncode(&SessionsResponse{Success: false, Message: err.Error()}), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&SessionsResponse{
		Success: true,
		Data:    sessions,
	})
}

func (h *SessionHandler) GetQRCode(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, `{"success":false,"message":"Session ID is required"}`, http.StatusBadRequest)
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, `{"success":false,"message":"Invalid session ID"}`, http.StatusBadRequest)
		return
	}

	qrChan, err := h.service.GetQRCode(r.Context(), id)
	if err != nil {
		http.Error(w, jsonEncode(&QRResponse{Success: false, Message: err.Error()}), http.StatusNotFound)
		return
	}

	qrCode := <-qrChan
	if qrCode == "" {
		http.Error(w, jsonEncode(&QRResponse{Success: false, Message: "Failed to get QR code"}), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&QRResponse{
		Success: true,
		Data:    qrCode,
	})
}

func (h *SessionHandler) PairPhone(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, `{"success":false,"message":"Session ID is required"}`, http.StatusBadRequest)
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, `{"success":false,"message":"Invalid session ID"}`, http.StatusBadRequest)
		return
	}

	var req PairRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"success":false,"message":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Phone == "" {
		http.Error(w, `{"success":false,"message":"Phone number is required"}`, http.StatusBadRequest)
		return
	}

	hasCountryCode := true
	if req.HasCountryCode != nil {
		hasCountryCode = *req.HasCountryCode
	}

	clientType := req.ClientType
	if clientType == "" {
		clientType = "chrome"
	}

	deviceName := req.DeviceName
	if deviceName == "" {
		deviceName = "Chrome (Linux)"
	}

	code, err := h.service.PairPhone(r.Context(), id, req.Phone, hasCountryCode, clientType, deviceName)
	if err != nil {
		http.Error(w, jsonEncode(&PairResponse{Success: false, Message: err.Error()}), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&PairResponse{
		Success: true,
		Data:    code,
		Message: "Enter this code on your WhatsApp app",
	})
}

func (h *SessionHandler) DeleteSession(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, `{"success":false,"message":"Session ID is required"}`, http.StatusBadRequest)
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, `{"success":false,"message":"Invalid session ID"}`, http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteSession(r.Context(), id); err != nil {
		http.Error(w, `{"success":false,"message":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Session deleted successfully",
	})
}

func jsonEncode(v interface{}) string {
	data, _ := json.Marshal(v)
	return string(data)
}
