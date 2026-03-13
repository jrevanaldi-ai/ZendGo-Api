package handler

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/zendgo/zendgo-api/internal/service"
	"github.com/zendgo/zendgo-api/models"
	"github.com/zendgo/zendgo-api/pkg/whatsapp"
)

type MessageHandler struct {
	service *service.WhatsAppService
}

func NewMessageHandler(service *service.WhatsAppService) *MessageHandler {
	return &MessageHandler{service: service}
}

type SendTextRequest struct {
	Recipient string `json:"recipient"`
	Message   string `json:"message"`
}

type SendImageRequest struct {
	Recipient string `json:"recipient"`
	ImageURL  string `json:"image_url"`
	Caption   string `json:"caption"`
}

type SendDocumentRequest struct {
	Recipient    string `json:"recipient"`
	DocumentURL  string `json:"document_url"`
	FileName     string `json:"file_name"`
	Caption      string `json:"caption"`
}

type SendAudioRequest struct {
	Recipient string `json:"recipient"`
	AudioURL  string `json:"audio_url"`
}

type SendVideoRequest struct {
	Recipient string `json:"recipient"`
	VideoURL  string `json:"video_url"`
	Caption   string `json:"caption"`
}

type SendLocationRequest struct {
	Recipient string  `json:"recipient"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Name      string  `json:"name"`
	Address   string  `json:"address"`
}

type SendContactRequest struct {
	Recipient    string `json:"recipient"`
	DisplayName  string `json:"display_name"`
	PhoneNumber  string `json:"phone_number"`
	Organization string `json:"organization"`
}

type SendCTAButtonRequest struct {
	Recipient  string `json:"recipient"`
	Text       string `json:"text"`
	ButtonText string `json:"button_text"`
	URL        string `json:"url"`
}

type MessageResponse struct {
	Success bool            `json:"success"`
	Data    *models.Message `json:"data"`
	Message string          `json:"message,omitempty"`
}

type CreateGroupRequest struct {
	Name         string   `json:"name"`
	Participants []string `json:"participants"`
}

type GroupResponse struct {
	Success bool                   `json:"success"`
	Data    *whatsapp.GroupInfo    `json:"data"`
	Message string                 `json:"message,omitempty"`
}

type GroupsResponse struct {
	Success bool                    `json:"success"`
	Data    []*whatsapp.GroupInfo   `json:"data"`
	Message string                  `json:"message,omitempty"`
}

func (h *MessageHandler) SendTextMessage(w http.ResponseWriter, r *http.Request) {
	sessionID, ok := r.Context().Value("session_id").(uuid.UUID)
	if !ok {
		http.Error(w, `{"success":false,"message":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var req SendTextRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"success":false,"message":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Recipient == "" {
		http.Error(w, `{"success":false,"message":"Recipient is required"}`, http.StatusBadRequest)
		return
	}

	if req.Message == "" {
		http.Error(w, `{"success":false,"message":"Message is required"}`, http.StatusBadRequest)
		return
	}

	msg, err := h.service.SendTextMessage(r.Context(), sessionID, req.Recipient, req.Message)
	if err != nil {
		http.Error(w, jsonEncode(&MessageResponse{Success: false, Message: err.Error()}), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(&MessageResponse{
		Success: true,
		Data:    msg,
		Message: "Message sent successfully",
	})
}

func (h *MessageHandler) SendImageMessage(w http.ResponseWriter, r *http.Request) {
	sessionID, ok := r.Context().Value("session_id").(uuid.UUID)
	if !ok {
		http.Error(w, `{"success":false,"message":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var req SendImageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"success":false,"message":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Recipient == "" {
		http.Error(w, `{"success":false,"message":"Recipient is required"}`, http.StatusBadRequest)
		return
	}

	if req.ImageURL == "" {
		http.Error(w, `{"success":false,"message":"Image URL is required"}`, http.StatusBadRequest)
		return
	}

	msg, err := h.service.SendImageMessage(r.Context(), sessionID, req.Recipient, req.ImageURL, req.Caption)
	if err != nil {
		http.Error(w, jsonEncode(&MessageResponse{Success: false, Message: err.Error()}), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(&MessageResponse{
		Success: true,
		Data:    msg,
		Message: "Image sent successfully",
	})
}

func (h *MessageHandler) SendDocumentMessage(w http.ResponseWriter, r *http.Request) {
	sessionID, ok := r.Context().Value("session_id").(uuid.UUID)
	if !ok {
		http.Error(w, `{"success":false,"message":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var req SendDocumentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"success":false,"message":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Recipient == "" {
		http.Error(w, `{"success":false,"message":"Recipient is required"}`, http.StatusBadRequest)
		return
	}

	if req.DocumentURL == "" {
		http.Error(w, `{"success":false,"message":"Document URL is required"}`, http.StatusBadRequest)
		return
	}

	msg, err := h.service.SendDocumentMessage(r.Context(), sessionID, req.Recipient, req.DocumentURL, req.FileName, req.Caption)
	if err != nil {
		http.Error(w, jsonEncode(&MessageResponse{Success: false, Message: err.Error()}), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(&MessageResponse{
		Success: true,
		Data:    msg,
		Message: "Document sent successfully",
	})
}

func (h *MessageHandler) SendAudioMessage(w http.ResponseWriter, r *http.Request) {
	sessionID, ok := r.Context().Value("session_id").(uuid.UUID)
	if !ok {
		http.Error(w, `{"success":false,"message":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var req SendAudioRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"success":false,"message":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Recipient == "" {
		http.Error(w, `{"success":false,"message":"Recipient is required"}`, http.StatusBadRequest)
		return
	}

	if req.AudioURL == "" {
		http.Error(w, `{"success":false,"message":"Audio URL is required"}`, http.StatusBadRequest)
		return
	}

	msg, err := h.service.SendAudioMessage(r.Context(), sessionID, req.Recipient, req.AudioURL)
	if err != nil {
		http.Error(w, jsonEncode(&MessageResponse{Success: false, Message: err.Error()}), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(&MessageResponse{
		Success: true,
		Data:    msg,
		Message: "Audio sent successfully",
	})
}

func (h *MessageHandler) SendVideoMessage(w http.ResponseWriter, r *http.Request) {
	sessionID, ok := r.Context().Value("session_id").(uuid.UUID)
	if !ok {
		http.Error(w, `{"success":false,"message":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var req SendVideoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"success":false,"message":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Recipient == "" {
		http.Error(w, `{"success":false,"message":"Recipient is required"}`, http.StatusBadRequest)
		return
	}

	if req.VideoURL == "" {
		http.Error(w, `{"success":false,"message":"Video URL is required"}`, http.StatusBadRequest)
		return
	}

	msg, err := h.service.SendVideoMessage(r.Context(), sessionID, req.Recipient, req.VideoURL, req.Caption)
	if err != nil {
		http.Error(w, jsonEncode(&MessageResponse{Success: false, Message: err.Error()}), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(&MessageResponse{
		Success: true,
		Data:    msg,
		Message: "Video sent successfully",
	})
}

func (h *MessageHandler) SendLocationMessage(w http.ResponseWriter, r *http.Request) {
	sessionID, ok := r.Context().Value("session_id").(uuid.UUID)
	if !ok {
		http.Error(w, `{"success":false,"message":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var req SendLocationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"success":false,"message":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Recipient == "" {
		http.Error(w, `{"success":false,"message":"Recipient is required"}`, http.StatusBadRequest)
		return
	}

	msg, err := h.service.SendLocationMessage(r.Context(), sessionID, req.Recipient, req.Latitude, req.Longitude, req.Name, req.Address)
	if err != nil {
		http.Error(w, jsonEncode(&MessageResponse{Success: false, Message: err.Error()}), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(&MessageResponse{
		Success: true,
		Data:    msg,
		Message: "Location sent successfully",
	})
}

func (h *MessageHandler) SendContactMessage(w http.ResponseWriter, r *http.Request) {
	sessionID, ok := r.Context().Value("session_id").(uuid.UUID)
	if !ok {
		http.Error(w, `{"success":false,"message":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var req SendContactRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"success":false,"message":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Recipient == "" {
		http.Error(w, `{"success":false,"message":"Recipient is required"}`, http.StatusBadRequest)
		return
	}

	if req.DisplayName == "" || req.PhoneNumber == "" {
		http.Error(w, `{"success":false,"message":"Display name and phone number are required"}`, http.StatusBadRequest)
		return
	}

	msg, err := h.service.SendContactMessage(r.Context(), sessionID, req.Recipient, req.DisplayName, req.PhoneNumber, req.Organization)
	if err != nil {
		http.Error(w, jsonEncode(&MessageResponse{Success: false, Message: err.Error()}), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(&MessageResponse{
		Success: true,
		Data:    msg,
		Message: "Contact sent successfully",
	})
}

func (h *MessageHandler) SendCTAButtonMessage(w http.ResponseWriter, r *http.Request) {
	sessionID, ok := r.Context().Value("session_id").(uuid.UUID)
	if !ok {
		http.Error(w, `{"success":false,"message":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var req SendCTAButtonRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"success":false,"message":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Recipient == "" {
		http.Error(w, `{"success":false,"message":"Recipient is required"}`, http.StatusBadRequest)
		return
	}

	if req.Text == "" || req.URL == "" {
		http.Error(w, `{"success":false,"message":"Text and URL are required"}`, http.StatusBadRequest)
		return
	}

	msg, err := h.service.SendCTAButtonMessage(r.Context(), sessionID, req.Recipient, req.Text, req.ButtonText, req.URL)
	if err != nil {
		http.Error(w, jsonEncode(&MessageResponse{Success: false, Message: err.Error()}), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(&MessageResponse{
		Success: true,
		Data:    msg,
		Message: "CTA button sent successfully",
	})
}

func (h *MessageHandler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	sessionID, ok := r.Context().Value("session_id").(uuid.UUID)
	if !ok {
		http.Error(w, `{"success":false,"message":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var req CreateGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"success":false,"message":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, `{"success":false,"message":"Group name is required"}`, http.StatusBadRequest)
		return
	}

	if len(req.Participants) == 0 {
		http.Error(w, `{"success":false,"message":"At least one participant is required"}`, http.StatusBadRequest)
		return
	}

	group, err := h.service.CreateGroup(r.Context(), sessionID, req.Name, req.Participants)
	if err != nil {
		http.Error(w, jsonEncode(&GroupResponse{Success: false, Message: err.Error()}), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(&GroupResponse{
		Success: true,
		Data:    group,
		Message: "Group created successfully",
	})
}

func (h *MessageHandler) ListGroups(w http.ResponseWriter, r *http.Request) {
	sessionID, ok := r.Context().Value("session_id").(uuid.UUID)
	if !ok {
		http.Error(w, `{"success":false,"message":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	groups, err := h.service.ListGroups(r.Context(), sessionID)
	if err != nil {
		http.Error(w, jsonEncode(&GroupsResponse{Success: false, Message: err.Error()}), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&GroupsResponse{
		Success: true,
		Data:    groups,
	})
}
