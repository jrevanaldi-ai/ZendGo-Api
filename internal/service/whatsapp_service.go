package service

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/zendgo/zendgo-api/internal/config"
	"github.com/zendgo/zendgo-api/internal/repository"
	"github.com/zendgo/zendgo-api/models"
	"github.com/zendgo/zendgo-api/pkg/whatsapp"
)

type WhatsAppService struct {
	db              *repository.Database
	sessionRepo     *repository.SessionRepository
	waSessionRepo   *repository.WASessionRepository
	messageRepo     *repository.MessageRepository
	config          *config.Config
	clients         map[uuid.UUID]*whatsapp.WhatsAppClient
	mu              sync.RWMutex
	webhookClient   *http.Client
}

func NewWhatsAppService(
	db *repository.Database,
	sessionRepo *repository.SessionRepository,
	waSessionRepo *repository.WASessionRepository,
	messageRepo *repository.MessageRepository,
	config *config.Config,
) *WhatsAppService {
	return &WhatsAppService{
		db:            db,
		sessionRepo:   sessionRepo,
		waSessionRepo: waSessionRepo,
		messageRepo:   messageRepo,
		config:        config,
		clients:       make(map[uuid.UUID]*whatsapp.WhatsAppClient),
		webhookClient: &http.Client{Timeout: 30000000000},
	}
}

func (s *WhatsAppService) generateAPIKey() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func (s *WhatsAppService) CreateSession(ctx context.Context, phone, webhookURL string) (*models.Session, error) {
	apiKey, err := s.generateAPIKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate API key: %w", err)
	}

	session := &models.Session{
		ID:         uuid.New(),
		Phone:      phone,
		Status:     models.SessionStatusConnecting,
		WebhookURL: webhookURL,
		APIKey:     apiKey,
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	client := whatsapp.NewWhatsAppClient(whatsapp.WhatsAppClientConfig{
		SessionID:      session.ID,
		Database:       s.db,
		SessionRepo:    s.sessionRepo,
		WASessionRepo:  s.waSessionRepo,
		MessageRepo:    s.messageRepo,
		Config:         s.config,
		WebhookHandler: s.handleWebhook,
	})

	if err := client.Initialize(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize WhatsApp client: %w", err)
	}

	if err := client.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	s.mu.Lock()
	s.clients[session.ID] = client
	s.mu.Unlock()

	return session, nil
}

func (s *WhatsAppService) GetSession(ctx context.Context, id uuid.UUID) (*models.Session, error) {
	return s.sessionRepo.GetByID(ctx, id)
}

func (s *WhatsAppService) GetSessionByAPIKey(ctx context.Context, apiKey string) (*models.Session, error) {
	return s.sessionRepo.GetByAPIKey(ctx, apiKey)
}

func (s *WhatsAppService) GetAllSessions(ctx context.Context) ([]*models.Session, error) {
	return s.sessionRepo.GetAll(ctx)
}

func (s *WhatsAppService) DeleteSession(ctx context.Context, id uuid.UUID) error {
	s.mu.Lock()
	if client, ok := s.clients[id]; ok {
		client.Disconnect()
		delete(s.clients, id)
	}
	s.mu.Unlock()

	return s.sessionRepo.Delete(ctx, id)
}

func (s *WhatsAppService) GetQRCode(ctx context.Context, id uuid.UUID) (<-chan string, error) {
	s.mu.RLock()
	client, ok := s.clients[id]
	s.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("session not found or not connected")
	}

	return client.GetQRChannel(ctx)
}

func (s *WhatsAppService) PairPhone(ctx context.Context, id uuid.UUID, phone string, hasCountryCode bool, clientType, deviceName string) (string, error) {
	s.mu.RLock()
	client, ok := s.clients[id]
	s.mu.RUnlock()

	if !ok {
		return "", fmt.Errorf("session not found")
	}

	return client.PairPhone(ctx, phone, hasCountryCode, clientType, deviceName)
}

func (s *WhatsAppService) SendTextMessage(ctx context.Context, sessionID uuid.UUID, recipient, message string) (*models.Message, error) {
	s.mu.RLock()
	client, ok := s.clients[sessionID]
	s.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("session not found")
	}

	msg := &models.Message{
		ID:          uuid.New(),
		SessionID:   sessionID,
		Recipient:   recipient,
		MessageType: models.MessageTypeText,
		Content:     message,
		Status:      models.MessageStatusQueued,
	}

	if err := s.messageRepo.Create(ctx, msg); err != nil {
		return nil, fmt.Errorf("failed to save message: %w", err)
	}

	waMessageID, err := client.SendTextMessage(ctx, recipient, message)
	if err != nil {
		s.messageRepo.UpdateStatus(ctx, msg.ID, models.MessageStatusFailed, "", err.Error())
		return nil, err
	}

	s.messageRepo.UpdateStatus(ctx, msg.ID, models.MessageStatusSent, waMessageID, "")
	msg.Status = models.MessageStatusSent
	msg.WAMessageID = waMessageID

	return msg, nil
}

func (s *WhatsAppService) SendImageMessage(ctx context.Context, sessionID uuid.UUID, recipient, imageURL, caption string) (*models.Message, error) {
	s.mu.RLock()
	client, ok := s.clients[sessionID]
	s.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("session not found")
	}

	msg := &models.Message{
		ID:          uuid.New(),
		SessionID:   sessionID,
		Recipient:   recipient,
		MessageType: models.MessageTypeImage,
		Content:     caption,
		MediaURL:    imageURL,
		Status:      models.MessageStatusQueued,
	}

	if err := s.messageRepo.Create(ctx, msg); err != nil {
		return nil, fmt.Errorf("failed to save message: %w", err)
	}

	waMessageID, err := client.SendImageMessage(ctx, recipient, imageURL, caption)
	if err != nil {
		s.messageRepo.UpdateStatus(ctx, msg.ID, models.MessageStatusFailed, "", err.Error())
		return nil, err
	}

	s.messageRepo.UpdateStatus(ctx, msg.ID, models.MessageStatusSent, waMessageID, "")
	msg.Status = models.MessageStatusSent
	msg.WAMessageID = waMessageID

	return msg, nil
}

func (s *WhatsAppService) SendDocumentMessage(ctx context.Context, sessionID uuid.UUID, recipient, documentURL, fileName, caption string) (*models.Message, error) {
	s.mu.RLock()
	client, ok := s.clients[sessionID]
	s.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("session not found")
	}

	msg := &models.Message{
		ID:          uuid.New(),
		SessionID:   sessionID,
		Recipient:   recipient,
		MessageType: models.MessageTypeDocument,
		Content:     caption,
		MediaURL:    documentURL,
		Status:      models.MessageStatusQueued,
	}

	if err := s.messageRepo.Create(ctx, msg); err != nil {
		return nil, fmt.Errorf("failed to save message: %w", err)
	}

	waMessageID, err := client.SendDocumentMessage(ctx, recipient, documentURL, fileName, caption)
	if err != nil {
		s.messageRepo.UpdateStatus(ctx, msg.ID, models.MessageStatusFailed, "", err.Error())
		return nil, err
	}

	s.messageRepo.UpdateStatus(ctx, msg.ID, models.MessageStatusSent, waMessageID, "")
	msg.Status = models.MessageStatusSent
	msg.WAMessageID = waMessageID

	return msg, nil
}

func (s *WhatsAppService) SendAudioMessage(ctx context.Context, sessionID uuid.UUID, recipient, audioURL string) (*models.Message, error) {
	s.mu.RLock()
	client, ok := s.clients[sessionID]
	s.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("session not found")
	}

	msg := &models.Message{
		ID:          uuid.New(),
		SessionID:   sessionID,
		Recipient:   recipient,
		MessageType: models.MessageTypeAudio,
		MediaURL:    audioURL,
		Status:      models.MessageStatusQueued,
	}

	if err := s.messageRepo.Create(ctx, msg); err != nil {
		return nil, fmt.Errorf("failed to save message: %w", err)
	}

	waMessageID, err := client.SendAudioMessage(ctx, recipient, audioURL)
	if err != nil {
		s.messageRepo.UpdateStatus(ctx, msg.ID, models.MessageStatusFailed, "", err.Error())
		return nil, err
	}

	s.messageRepo.UpdateStatus(ctx, msg.ID, models.MessageStatusSent, waMessageID, "")
	msg.Status = models.MessageStatusSent
	msg.WAMessageID = waMessageID

	return msg, nil
}

func (s *WhatsAppService) SendVideoMessage(ctx context.Context, sessionID uuid.UUID, recipient, videoURL, caption string) (*models.Message, error) {
	s.mu.RLock()
	client, ok := s.clients[sessionID]
	s.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("session not found")
	}

	msg := &models.Message{
		ID:          uuid.New(),
		SessionID:   sessionID,
		Recipient:   recipient,
		MessageType: models.MessageTypeVideo,
		Content:     caption,
		MediaURL:    videoURL,
		Status:      models.MessageStatusQueued,
	}

	if err := s.messageRepo.Create(ctx, msg); err != nil {
		return nil, fmt.Errorf("failed to save message: %w", err)
	}

	waMessageID, err := client.SendVideoMessage(ctx, recipient, videoURL, caption)
	if err != nil {
		s.messageRepo.UpdateStatus(ctx, msg.ID, models.MessageStatusFailed, "", err.Error())
		return nil, err
	}

	s.messageRepo.UpdateStatus(ctx, msg.ID, models.MessageStatusSent, waMessageID, "")
	msg.Status = models.MessageStatusSent
	msg.WAMessageID = waMessageID

	return msg, nil
}

func (s *WhatsAppService) SendLocationMessage(ctx context.Context, sessionID uuid.UUID, recipient string, latitude, longitude float64, name, address string) (*models.Message, error) {
	s.mu.RLock()
	client, ok := s.clients[sessionID]
	s.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("session not found")
	}

	msg := &models.Message{
		ID:          uuid.New(),
		SessionID:   sessionID,
		Recipient:   recipient,
		MessageType: models.MessageTypeLocation,
		Content:     fmt.Sprintf("%s, %s", name, address),
		Status:      models.MessageStatusQueued,
	}

	if err := s.messageRepo.Create(ctx, msg); err != nil {
		return nil, fmt.Errorf("failed to save message: %w", err)
	}

	waMessageID, err := client.SendLocationMessage(ctx, recipient, latitude, longitude, name, address)
	if err != nil {
		s.messageRepo.UpdateStatus(ctx, msg.ID, models.MessageStatusFailed, "", err.Error())
		return nil, err
	}

	s.messageRepo.UpdateStatus(ctx, msg.ID, models.MessageStatusSent, waMessageID, "")
	msg.Status = models.MessageStatusSent
	msg.WAMessageID = waMessageID

	return msg, nil
}

func (s *WhatsAppService) SendContactMessage(ctx context.Context, sessionID uuid.UUID, recipient, displayName, phoneNumber, organization string) (*models.Message, error) {
	s.mu.RLock()
	client, ok := s.clients[sessionID]
	s.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("session not found")
	}

	msg := &models.Message{
		ID:          uuid.New(),
		SessionID:   sessionID,
		Recipient:   recipient,
		MessageType: models.MessageTypeContact,
		Content:     fmt.Sprintf("%s - %s", displayName, phoneNumber),
		Status:      models.MessageStatusQueued,
	}

	if err := s.messageRepo.Create(ctx, msg); err != nil {
		return nil, fmt.Errorf("failed to save message: %w", err)
	}

	waMessageID, err := client.SendContactMessage(ctx, recipient, displayName, phoneNumber, organization)
	if err != nil {
		s.messageRepo.UpdateStatus(ctx, msg.ID, models.MessageStatusFailed, "", err.Error())
		return nil, err
	}

	s.messageRepo.UpdateStatus(ctx, msg.ID, models.MessageStatusSent, waMessageID, "")
	msg.Status = models.MessageStatusSent
	msg.WAMessageID = waMessageID

	return msg, nil
}

func (s *WhatsAppService) SendCTAButtonMessage(ctx context.Context, sessionID uuid.UUID, recipient, text, buttonText, url string) (*models.Message, error) {
	s.mu.RLock()
	client, ok := s.clients[sessionID]
	s.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("session not found")
	}

	msg := &models.Message{
		ID:          uuid.New(),
		SessionID:   sessionID,
		Recipient:   recipient,
		MessageType: models.MessageTypeText,
		Content:     fmt.Sprintf("%s - %s", text, url),
		Status:      models.MessageStatusQueued,
	}

	if err := s.messageRepo.Create(ctx, msg); err != nil {
		return nil, fmt.Errorf("failed to save message: %w", err)
	}

	waMessageID, err := client.SendCTAButtonMessage(ctx, recipient, text, buttonText, url)
	if err != nil {
		s.messageRepo.UpdateStatus(ctx, msg.ID, models.MessageStatusFailed, "", err.Error())
		return nil, err
	}

	s.messageRepo.UpdateStatus(ctx, msg.ID, models.MessageStatusSent, waMessageID, "")
	msg.Status = models.MessageStatusSent
	msg.WAMessageID = waMessageID

	return msg, nil
}

func (s *WhatsAppService) CreateGroup(ctx context.Context, sessionID uuid.UUID, name string, participants []string) (*whatsapp.GroupInfo, error) {
	s.mu.RLock()
	client, ok := s.clients[sessionID]
	s.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("session not found")
	}

	return client.CreateGroup(ctx, name, participants)
}

func (s *WhatsAppService) ListGroups(ctx context.Context, sessionID uuid.UUID) ([]*whatsapp.GroupInfo, error) {
	s.mu.RLock()
	client, ok := s.clients[sessionID]
	s.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("session not found")
	}

	return client.ListGroups(ctx)
}

func (s *WhatsAppService) handleWebhook(sessionID uuid.UUID, msg *models.IncomingMessage) error {
	if err := s.messageRepo.SaveIncomingMessage(context.Background(), msg); err != nil {
		return fmt.Errorf("failed to save incoming message: %w", err)
	}

	session, err := s.sessionRepo.GetByID(context.Background(), sessionID)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	if session.WebhookURL == "" {
		return nil
	}

	go s.sendWebhook(session.WebhookURL, msg)

	return nil
}

func (s *WhatsAppService) sendWebhook(url string, msg *models.IncomingMessage) {
	payload := map[string]interface{}{
		"session_id":   msg.SessionID.String(),
		"sender":       msg.Sender,
		"message_type": msg.MessageType,
		"content":      msg.Content,
		"caption":      msg.Caption,
		"timestamp":    msg.Timestamp,
		"is_group":     msg.IsGroup,
		"from":         msg.From,
		"push_name":    msg.PushName,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("Failed to marshal webhook payload: %v\n", err)
		return
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Failed to create webhook request: %v\n", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.webhookClient.Do(req)
	if err != nil {
		fmt.Printf("Failed to send webhook: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Webhook sent to %s, status: %d\n", url, resp.StatusCode)
}
