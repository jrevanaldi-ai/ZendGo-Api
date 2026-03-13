package whatsapp

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
)

func (w *WhatsAppClient) SendTextMessage(ctx context.Context, recipient, message string) (string, error) {
	if w.client == nil {
		return "", fmt.Errorf("client not initialized")
	}

	jid, err := w.parseJID(recipient)
	if err != nil {
		return "", fmt.Errorf("failed to parse JID: %w", err)
	}

	msg := &waE2E.Message{
		Conversation: &message,
	}

	resp, err := w.client.SendMessage(ctx, jid, msg)
	if err != nil {
		return "", fmt.Errorf("failed to send message: %w", err)
	}

	return resp.ID, nil
}

func (w *WhatsAppClient) SendImageMessage(ctx context.Context, recipient, imageURL, caption string) (string, error) {
	if w.client == nil {
		return "", fmt.Errorf("client not initialized")
	}

	jid, err := w.parseJID(recipient)
	if err != nil {
		return "", fmt.Errorf("failed to parse JID: %w", err)
	}

	imgData, mimeType, err := w.downloadMediaWithMime(imageURL)
	if err != nil {
		return "", fmt.Errorf("failed to download image: %w", err)
	}

	uploaded, err := w.client.Upload(ctx, imgData, whatsmeow.MediaImage)
	if err != nil {
		return "", fmt.Errorf("failed to upload image: %w", err)
	}

	msg := &waE2E.Message{
		ImageMessage: &waE2E.ImageMessage{
			Caption:       &caption,
			Mimetype:      &mimeType,
			URL:           &uploaded.URL,
			DirectPath:    &uploaded.DirectPath,
			MediaKey:      uploaded.MediaKey,
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    protoUint64(uint64(len(imgData))),
		},
	}

	resp, err := w.client.SendMessage(ctx, jid, msg)
	if err != nil {
		return "", fmt.Errorf("failed to send image: %w", err)
	}

	return resp.ID, nil
}

func (w *WhatsAppClient) SendDocumentMessage(ctx context.Context, recipient, documentURL, fileName, caption string) (string, error) {
	if w.client == nil {
		return "", fmt.Errorf("client not initialized")
	}

	jid, err := w.parseJID(recipient)
	if err != nil {
		return "", fmt.Errorf("failed to parse JID: %w", err)
	}

	docData, mimeType, err := w.downloadMediaWithMime(documentURL)
	if err != nil {
		return "", fmt.Errorf("failed to download document: %w", err)
	}

	uploaded, err := w.client.Upload(ctx, docData, whatsmeow.MediaDocument)
	if err != nil {
		return "", fmt.Errorf("failed to upload document: %w", err)
	}

	msg := &waE2E.Message{
		DocumentMessage: &waE2E.DocumentMessage{
			Title:         &fileName,
			Caption:       &caption,
			Mimetype:      &mimeType,
			URL:           &uploaded.URL,
			DirectPath:    &uploaded.DirectPath,
			MediaKey:      uploaded.MediaKey,
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    protoUint64(uint64(len(docData))),
		},
	}

	resp, err := w.client.SendMessage(ctx, jid, msg)
	if err != nil {
		return "", fmt.Errorf("failed to send document: %w", err)
	}

	return resp.ID, nil
}

func (w *WhatsAppClient) SendAudioMessage(ctx context.Context, recipient, audioURL string) (string, error) {
	if w.client == nil {
		return "", fmt.Errorf("client not initialized")
	}

	jid, err := w.parseJID(recipient)
	if err != nil {
		return "", fmt.Errorf("failed to parse JID: %w", err)
	}

	audioData, mimeType, err := w.downloadMediaWithMime(audioURL)
	if err != nil {
		return "", fmt.Errorf("failed to download audio: %w", err)
	}

	uploaded, err := w.client.Upload(ctx, audioData, whatsmeow.MediaAudio)
	if err != nil {
		return "", fmt.Errorf("failed to upload audio: %w", err)
	}

	msg := &waE2E.Message{
		AudioMessage: &waE2E.AudioMessage{
			Mimetype:      &mimeType,
			URL:           &uploaded.URL,
			DirectPath:    &uploaded.DirectPath,
			MediaKey:      uploaded.MediaKey,
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    protoUint64(uint64(len(audioData))),
		},
	}

	resp, err := w.client.SendMessage(ctx, jid, msg)
	if err != nil {
		return "", fmt.Errorf("failed to send audio: %w", err)
	}

	return resp.ID, nil
}

func (w *WhatsAppClient) SendVideoMessage(ctx context.Context, recipient, videoURL, caption string) (string, error) {
	if w.client == nil {
		return "", fmt.Errorf("client not initialized")
	}

	jid, err := w.parseJID(recipient)
	if err != nil {
		return "", fmt.Errorf("failed to parse JID: %w", err)
	}

	videoData, mimeType, err := w.downloadMediaWithMime(videoURL)
	if err != nil {
		return "", fmt.Errorf("failed to download video: %w", err)
	}

	uploaded, err := w.client.Upload(ctx, videoData, whatsmeow.MediaVideo)
	if err != nil {
		return "", fmt.Errorf("failed to upload video: %w", err)
	}

	msg := &waE2E.Message{
		VideoMessage: &waE2E.VideoMessage{
			Caption:       &caption,
			Mimetype:      &mimeType,
			URL:           &uploaded.URL,
			DirectPath:    &uploaded.DirectPath,
			MediaKey:      uploaded.MediaKey,
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    protoUint64(uint64(len(videoData))),
		},
	}

	resp, err := w.client.SendMessage(ctx, jid, msg)
	if err != nil {
		return "", fmt.Errorf("failed to send video: %w", err)
	}

	return resp.ID, nil
}

func (w *WhatsAppClient) SendLocationMessage(ctx context.Context, recipient string, latitude, longitude float64, name, address string) (string, error) {
	if w.client == nil {
		return "", fmt.Errorf("client not initialized")
	}

	jid, err := w.parseJID(recipient)
	if err != nil {
		return "", fmt.Errorf("failed to parse JID: %w", err)
	}

	msg := &waE2E.Message{
		LocationMessage: &waE2E.LocationMessage{
			DegreesLatitude:  &latitude,
			DegreesLongitude: &longitude,
			Name:             &name,
			Address:          &address,
		},
	}

	resp, err := w.client.SendMessage(ctx, jid, msg)
	if err != nil {
		return "", fmt.Errorf("failed to send location: %w", err)
	}

	return resp.ID, nil
}

func (w *WhatsAppClient) SendContactMessage(ctx context.Context, recipient, displayName, phoneNumber, organization string) (string, error) {
	if w.client == nil {
		return "", fmt.Errorf("client not initialized")
	}

	jid, err := w.parseJID(recipient)
	if err != nil {
		return "", fmt.Errorf("failed to parse JID: %w", err)
	}

	vcard := fmt.Sprintf("BEGIN:VCARD\nVERSION:3.0\nN:%s\nFN:%s\nTEL;TYPE=CELL:%s\nORG:%s\nEND:VCARD", displayName, displayName, phoneNumber, organization)

	msg := &waE2E.Message{
		ContactMessage: &waE2E.ContactMessage{
			DisplayName: &displayName,
			Vcard:       &vcard,
		},
	}

	resp, err := w.client.SendMessage(ctx, jid, msg)
	if err != nil {
		return "", fmt.Errorf("failed to send contact: %w", err)
	}

	return resp.ID, nil
}

func (w *WhatsAppClient) parseJID(phone string) (types.JID, error) {
	if phone == "" {
		return types.EmptyJID, fmt.Errorf("phone number is required")
	}

	if phone == "status@broadcast" || phone == "news@broadcast" {
		return types.JID{User: phone, Server: "broadcast"}, nil
	}

	jid, err := types.ParseJID(phone)
	if err == nil && !jid.IsEmpty() {
		return jid, nil
	}

	if phone[0] == '+' {
		phone = phone[1:]
	}

	if len(phone) < 10 {
		return types.EmptyJID, fmt.Errorf("invalid phone number: too short")
	}

	jid = types.NewJID(phone, types.DefaultUserServer)
	return jid, nil
}

func (w *WhatsAppClient) downloadMediaWithMime(url string) ([]byte, string, error) {
	if url == "" {
		return nil, "", fmt.Errorf("media URL is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, "", err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	mimeType := resp.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	return data, mimeType, nil
}

func protoUint64(v uint64) *uint64 {
	return &v
}
