package whatsapp

import (
	"context"
	"fmt"
	"time"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
)

type GroupInfo struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Topic        string    `json:"topic"`
	Owner        string    `json:"owner"`
	Participants []string  `json:"participants"`
	CreatedAt    time.Time `json:"created_at"`
}

func (w *WhatsAppClient) CreateGroup(ctx context.Context, name string, participants []string) (*GroupInfo, error) {
	if w.client == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	jids := make([]types.JID, 0, len(participants))
	for _, p := range participants {
		jid, err := w.parseJID(p)
		if err != nil {
			return nil, fmt.Errorf("invalid participant %s: %w", p, err)
		}
		jids = append(jids, jid)
	}

	req := whatsmeow.ReqCreateGroup{
		Name:         name,
		Participants: jids,
	}

	groupInfo, err := w.client.CreateGroup(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create group: %w", err)
	}

	return groupInfoToDTO(groupInfo), nil
}

func (w *WhatsAppClient) GetGroupInfo(ctx context.Context, groupJID string) (*GroupInfo, error) {
	if w.client == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	jid, err := types.ParseJID(groupJID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse group JID: %w", err)
	}

	info, err := w.client.GetGroupInfo(ctx, jid)
	if err != nil {
		return nil, fmt.Errorf("failed to get group info: %w", err)
	}

	return groupInfoToDTO(info), nil
}

func (w *WhatsAppClient) ListGroups(ctx context.Context) ([]*GroupInfo, error) {
	if w.client == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	groups, err := w.client.GetJoinedGroups(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list groups: %w", err)
	}

	result := make([]*GroupInfo, 0, len(groups))
	for _, g := range groups {
		result = append(result, groupInfoToDTO(g))
	}

	return result, nil
}

func (w *WhatsAppClient) AddGroupParticipants(ctx context.Context, groupJID string, participants []string) error {
	if w.client == nil {
		return fmt.Errorf("client not initialized")
	}

	jids := make([]types.JID, 0, len(participants))
	for _, p := range participants {
		jid, err := w.parseJID(p)
		if err != nil {
			return fmt.Errorf("invalid participant %s: %w", p, err)
		}
		jids = append(jids, jid)
	}

	_, err := w.client.UpdateGroupParticipants(ctx, types.NewJID(groupJID, types.GroupServer), jids, whatsmeow.ParticipantChangeAdd)
	return err
}

func (w *WhatsAppClient) RemoveGroupParticipants(ctx context.Context, groupJID string, participants []string) error {
	if w.client == nil {
		return fmt.Errorf("client not initialized")
	}

	jids := make([]types.JID, 0, len(participants))
	for _, p := range participants {
		jid, err := w.parseJID(p)
		if err != nil {
			return fmt.Errorf("invalid participant %s: %w", p, err)
		}
		jids = append(jids, jid)
	}

	_, err := w.client.UpdateGroupParticipants(ctx, types.NewJID(groupJID, types.GroupServer), jids, whatsmeow.ParticipantChangeRemove)
	return err
}

func (w *WhatsAppClient) SetGroupName(ctx context.Context, groupJID, name string) error {
	if w.client == nil {
		return fmt.Errorf("client not initialized")
	}

	err := w.client.SetGroupName(ctx, types.NewJID(groupJID, types.GroupServer), name)
	return err
}

func (w *WhatsAppClient) SetGroupDescription(ctx context.Context, groupJID, description string) error {
	if w.client == nil {
		return fmt.Errorf("client not initialized")
	}

	err := w.client.SetGroupDescription(ctx, types.NewJID(groupJID, types.GroupServer), description)
	return err
}

func (w *WhatsAppClient) SetGroupTopic(ctx context.Context, groupJID, previousID, newID, topic string) error {
	if w.client == nil {
		return fmt.Errorf("client not initialized")
	}

	err := w.client.SetGroupTopic(ctx, types.NewJID(groupJID, types.GroupServer), previousID, newID, topic)
	return err
}

func (w *WhatsAppClient) GetGroupInviteLink(ctx context.Context, groupJID string, forceReset bool) (string, error) {
	if w.client == nil {
		return "", fmt.Errorf("client not initialized")
	}

	link, err := w.client.GetGroupInviteLink(ctx, types.NewJID(groupJID, types.GroupServer), forceReset)
	if err != nil {
		return "", err
	}

	return link, nil
}

func (w *WhatsAppClient) JoinGroupWithLink(ctx context.Context, link string) (string, error) {
	if w.client == nil {
		return "", fmt.Errorf("client not initialized")
	}

	groupJID, err := w.client.JoinGroupWithLink(ctx, link)
	if err != nil {
		return "", err
	}

	return groupJID.String(), nil
}

func (w *WhatsAppClient) LeaveGroup(ctx context.Context, groupJID string) error {
	if w.client == nil {
		return fmt.Errorf("client not initialized")
	}

	err := w.client.LeaveGroup(ctx, types.NewJID(groupJID, types.GroupServer))
	return err
}

func (w *WhatsAppClient) SendPresence(ctx context.Context, presence string) error {
	if w.client == nil {
		return fmt.Errorf("client not initialized")
	}

	var presenceType types.Presence
	switch presence {
	case "available":
		presenceType = types.PresenceAvailable
	case "unavailable":
		presenceType = types.PresenceUnavailable
	default:
		presenceType = types.PresenceAvailable
	}

	err := w.client.SendPresence(ctx, presenceType)
	return err
}

func (w *WhatsAppClient) SendChatPresence(ctx context.Context, chatJID string, state string) error {
	if w.client == nil {
		return fmt.Errorf("client not initialized")
	}

	jid, err := types.ParseJID(chatJID)
	if err != nil {
		return err
	}

	var chatPresenceType types.ChatPresence
	switch state {
	case "composing":
		chatPresenceType = types.ChatPresenceComposing
	case "paused":
		chatPresenceType = types.ChatPresencePaused
	default:
		chatPresenceType = types.ChatPresenceComposing
	}

	err = w.client.SendChatPresence(ctx, jid, chatPresenceType, types.ChatPresenceMediaText)
	return err
}

func (w *WhatsAppClient) MarkRead(ctx context.Context, messageIDs []string, chatJID string) error {
	if w.client == nil {
		return fmt.Errorf("client not initialized")
	}

	jid, err := types.ParseJID(chatJID)
	if err != nil {
		return err
	}

	err = w.client.MarkRead(ctx, messageIDs, time.Now(), jid, types.EmptyJID)
	return err
}

func (w *WhatsAppClient) GetProfilePicture(ctx context.Context, jidStr string) (string, error) {
	if w.client == nil {
		return "", fmt.Errorf("client not initialized")
	}

	jid, err := types.ParseJID(jidStr)
	if err != nil {
		return "", err
	}

	pic, err := w.client.GetProfilePictureInfo(ctx, jid, nil)
	if err != nil {
		return "", err
	}

	if pic == nil {
		return "", fmt.Errorf("no profile picture found")
	}

	return pic.URL, nil
}

func (w *WhatsAppClient) GetContactName(ctx context.Context, jidStr string) (string, error) {
	if w.client == nil {
		return "", fmt.Errorf("client not initialized")
	}

	jid, err := types.ParseJID(jidStr)
	if err != nil {
		return "", err
	}

	contact, err := w.client.Store.Contacts.GetContact(ctx, jid)
	if err != nil {
		return "", err
	}

	if contact.FullName != "" {
		return contact.FullName, nil
	}

	return contact.FirstName, nil
}

func (w *WhatsAppClient) IsOnWhatsApp(ctx context.Context, phoneNumbers []string) ([]types.IsOnWhatsAppResponse, error) {
	if w.client == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	responses, err := w.client.IsOnWhatsApp(ctx, phoneNumbers)
	if err != nil {
		return nil, err
	}

	return responses, nil
}

func (w *WhatsAppClient) GetBusinessProfile(ctx context.Context, jidStr string) (map[string]interface{}, error) {
	if w.client == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	jid, err := types.ParseJID(jidStr)
	if err != nil {
		return nil, err
	}

	profile, err := w.client.GetBusinessProfile(ctx, jid)
	if err != nil {
		return nil, err
	}

	result := make(map[string]interface{})
	if profile.Address != "" {
		result["address"] = profile.Address
	}
	if profile.Email != "" {
		result["email"] = profile.Email
	}
	if profile.BusinessHours != nil && len(profile.BusinessHours) > 0 {
		result["business_hours"] = profile.BusinessHours
	}
	if profile.Categories != nil && len(profile.Categories) > 0 {
		categories := make([]string, 0)
		for _, cat := range profile.Categories {
			categories = append(categories, cat.Name)
		}
		result["categories"] = categories
	}

	return result, nil
}

func (w *WhatsAppClient) DeleteMessage(ctx context.Context, jidStr, messageID string, forEveryone bool) error {
	if w.client == nil {
		return fmt.Errorf("client not initialized")
	}

	jid, err := types.ParseJID(jidStr)
	if err != nil {
		return err
	}

	revokeMsg := w.client.BuildRevoke(jid, types.EmptyJID, messageID)
	_, err = w.client.SendMessage(ctx, jid, revokeMsg)
	return err
}

func (w *WhatsAppClient) ReactMessage(ctx context.Context, jidStr, messageID, reaction string) error {
	if w.client == nil {
		return fmt.Errorf("client not initialized")
	}

	jid, err := types.ParseJID(jidStr)
	if err != nil {
		return err
	}

	reactionMsg := w.client.BuildReaction(jid, types.EmptyJID, messageID, reaction)
	_, err = w.client.SendMessage(ctx, jid, reactionMsg)
	return err
}

func (w *WhatsAppClient) BuildPoll(ctx context.Context, name string, options []string, maxSelections int) (*waE2E.Message, error) {
	if w.client == nil {
		return nil, fmt.Errorf("client not initialized")
	}

	msg := w.client.BuildPollCreation(name, options, maxSelections)
	return msg, nil
}

func (w *WhatsAppClient) SendPoll(ctx context.Context, recipient string, name string, options []string, maxSelections int) (string, error) {
	if w.client == nil {
		return "", fmt.Errorf("client not initialized")
	}

	jid, err := w.parseJID(recipient)
	if err != nil {
		return "", fmt.Errorf("failed to parse JID: %w", err)
	}

	pollMsg := w.client.BuildPollCreation(name, options, maxSelections)
	resp, err := w.client.SendMessage(ctx, jid, pollMsg)
	if err != nil {
		return "", fmt.Errorf("failed to send poll: %w", err)
	}

	return resp.ID, nil
}

func groupInfoToDTO(info *types.GroupInfo) *GroupInfo {
	participants := make([]string, 0, len(info.Participants))
	for _, p := range info.Participants {
		participants = append(participants, p.JID.String())
	}

	owner := ""
	if !info.OwnerJID.IsEmpty() {
		owner = info.OwnerJID.String()
	}

	return &GroupInfo{
		ID:           info.JID.String(),
		Name:         info.Name,
		Topic:        info.Topic,
		Owner:        owner,
		Participants: participants,
		CreatedAt:    info.GroupCreated,
	}
}
