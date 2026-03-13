package whatsapp

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	"github.com/zendgo/zendgo-api/internal/config"
	"github.com/zendgo/zendgo-api/internal/repository"
	"github.com/zendgo/zendgo-api/models"
)

type WhatsAppClient struct {
	client           *whatsmeow.Client
	sessionID        uuid.UUID
	db               *repository.Database
	sessionRepo      *repository.SessionRepository
	waSessionRepo    *repository.WASessionRepository
	messageRepo      *repository.MessageRepository
	config           *config.Config
	eventHandlers    []func(interface{})
	mu               sync.RWMutex
	isConnected      bool
	webhookHandler   WebhookHandler
}

type WebhookHandler func(sessionID uuid.UUID, msg *models.IncomingMessage) error

type WhatsAppClientConfig struct {
	SessionID      uuid.UUID
	Database       *repository.Database
	SessionRepo    *repository.SessionRepository
	WASessionRepo  *repository.WASessionRepository
	MessageRepo    *repository.MessageRepository
	Config         *config.Config
	WebhookHandler WebhookHandler
}

func NewWhatsAppClient(cfg WhatsAppClientConfig) *WhatsAppClient {
	return &WhatsAppClient{
		sessionID:      cfg.SessionID,
		db:             cfg.Database,
		sessionRepo:    cfg.SessionRepo,
		waSessionRepo:  cfg.WASessionRepo,
		messageRepo:    cfg.MessageRepo,
		config:         cfg.Config,
		webhookHandler: cfg.WebhookHandler,
		eventHandlers:  make([]func(interface{}), 0),
	}
}

func (w *WhatsAppClient) Initialize(ctx context.Context) error {
	logLevel := zerolog.InfoLevel
	if w.config.WhatsApp.LogLevel == "debug" {
		logLevel = zerolog.DebugLevel
	}

	dbLog := waLog.Stdout("Database", logLevel.String(), true)
	container, err := sqlstore.New(ctx, "postgres", w.db.Config.DSN(), dbLog)
	if err != nil {
		return fmt.Errorf("failed to create store container: %w", err)
	}

	deviceStore, err := container.GetFirstDevice(ctx)
	if err != nil {
		return fmt.Errorf("failed to get device: %w", err)
	}

	if deviceStore == nil {
		deviceStore = container.NewDevice()
	}

	clientLog := waLog.Stdout("Client", logLevel.String(), true)
	w.client = whatsmeow.NewClient(deviceStore, clientLog)

	w.setupEventHandlers()

	return nil
}

func (w *WhatsAppClient) setupEventHandlers() {
	w.client.AddEventHandler(w.handleEvents)
}

func (w *WhatsAppClient) handleEvents(evt interface{}) {
	w.mu.RLock()
	handlers := make([]func(interface{}), len(w.eventHandlers))
	copy(handlers, w.eventHandlers)
	w.mu.RUnlock()

	for _, handler := range handlers {
		handler(evt)
	}

	switch v := evt.(type) {
	case *events.Message:
		w.handleIncomingMessage(v)
	case *events.QR:
		w.handleQRCode(v)
	case *events.PairSuccess:
		w.handlePairSuccess(v)
	case *events.Disconnected:
		w.handleDisconnected()
	case *events.Connected:
		w.handleConnected()
	}
}

func (w *WhatsAppClient) handleIncomingMessage(msg *events.Message) {
	if msg.Info.IsFromMe {
		return
	}

	messageType := models.MessageTypeText
	content := ""
	caption := ""

	if msg.Message.GetConversation() != "" {
		content = msg.Message.GetConversation()
	} else if msg.Message.ExtendedTextMessage != nil {
		content = msg.Message.ExtendedTextMessage.GetText()
		caption = msg.Message.ExtendedTextMessage.GetText()
	}

	if msg.Message.ImageMessage != nil {
		messageType = models.MessageTypeImage
		content = "[Image]"
	} else if msg.Message.DocumentMessage != nil {
		messageType = models.MessageTypeDocument
		content = "[Document]"
	} else if msg.Message.AudioMessage != nil {
		messageType = models.MessageTypeAudio
		content = "[Audio]"
	} else if msg.Message.VideoMessage != nil {
		messageType = models.MessageTypeVideo
		content = "[Video]"
	}

	incomingMsg := &models.IncomingMessage{
		ID:          uuid.New(),
		SessionID:   w.sessionID,
		Sender:      msg.Info.Sender.String(),
		MessageType: messageType,
		Content:     content,
		Caption:     caption,
		Timestamp:   msg.Info.Timestamp.Unix(),
		IsGroup:     msg.Info.IsGroup,
		From:        msg.Info.Chat.String(),
		PushName:    msg.Info.PushName,
		CreatedAt:   time.Now(),
	}

	if w.webhookHandler != nil {
		w.webhookHandler(w.sessionID, incomingMsg)
	}
}

func (w *WhatsAppClient) handleQRCode(qr *events.QR) {
	if len(qr.Codes) > 0 {
		fmt.Printf("[Session %s] QR Code: %s\n", w.sessionID, qr.Codes[0])
	}
}

func (w *WhatsAppClient) handlePairSuccess(pair *events.PairSuccess) {
	fmt.Printf("[Session %s] Pair success! JID: %s\n", w.sessionID, pair.ID.String())
	
	ctx := context.Background()
	if err := w.sessionRepo.UpdateStatus(ctx, w.sessionID, models.SessionStatusPaired); err != nil {
		fmt.Printf("Failed to update session status: %v\n", err)
	}
}

func (w *WhatsAppClient) handleConnected() {
	fmt.Printf("[Session %s] Connected to WhatsApp\n", w.sessionID)
	w.mu.Lock()
	w.isConnected = true
	w.mu.Unlock()
}

func (w *WhatsAppClient) handleDisconnected() {
	fmt.Printf("[Session %s] Disconnected from WhatsApp\n", w.sessionID)
	w.mu.Lock()
	w.isConnected = false
	w.mu.Unlock()
}

func (w *WhatsAppClient) Connect() error {
	if w.client == nil {
		return fmt.Errorf("client not initialized")
	}
	
	err := w.client.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	return nil
}

func (w *WhatsAppClient) Disconnect() {
	if w.client != nil {
		w.client.Disconnect()
	}
	w.mu.Lock()
	w.isConnected = false
	w.mu.Unlock()
}

func (w *WhatsAppClient) IsConnected() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.isConnected
}

func (w *WhatsAppClient) GetQRChannel(ctx context.Context) (<-chan string, error) {
	if w.client == nil {
		return nil, fmt.Errorf("client not initialized")
	}
	
	qrChan := make(chan string, 1)
	
	ch, err := w.client.GetQRChannel(ctx)
	if err != nil {
		return nil, err
	}
	
	go func() {
		for evt := range ch {
			if evt.Event == "code" {
				qrChan <- evt.Code
				close(qrChan)
				return
			}
		}
		close(qrChan)
	}()
	
	return qrChan, nil
}

func (w *WhatsAppClient) PairPhone(ctx context.Context, phone string, hasCountryCode bool, pairingCode, deviceName string) (string, error) {
	if w.client == nil {
		return "", fmt.Errorf("client not initialized")
	}
	
	code, err := w.client.PairPhone(ctx, phone, hasCountryCode, whatsmeow.PairClientChrome, deviceName)
	if err != nil {
		return "", err
	}
	return code, nil
}

func (w *WhatsAppClient) AddEventHandler(handler func(interface{})) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.eventHandlers = append(w.eventHandlers, handler)
}

func (w *WhatsAppClient) downloadMedia(url string) ([]byte, error) {
	if url == "" {
		return nil, fmt.Errorf("media URL is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}
