package service

import (
	"encoding/json"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/duyb/esport-score-tracker/internal/model"
	"github.com/duyb/esport-score-tracker/internal/ws"
	"github.com/google/uuid"
)

// ChatRepository is the persistence interface for WcChatService.
type ChatRepository interface {
	Save(msg *model.WcChatMessage) error
	ListMessages(limit int, before *time.Time) ([]model.WcChatMessage, error)
	SaveMentions(messageID uuid.UUID, userIDs []uuid.UUID) error
	UnreadMentionCount(userID uuid.UUID) (int64, error)
	MarkMentionsRead(userID uuid.UUID) error
}

// ChatHubBroadcaster is the interface WcChatService depends on for broadcasting.
type ChatHubBroadcaster interface {
	BroadcastChat(event ws.ChatMessageEvent)
	SendToUser(userID uuid.UUID, data []byte)
}

type WcChatService struct {
	repo ChatRepository
	hub  ChatHubBroadcaster
}

func NewWcChatService(repo ChatRepository, hub ChatHubBroadcaster) *WcChatService {
	return &WcChatService{repo: repo, hub: hub}
}

var (
	ErrChatMessageEmpty   = errors.New("message cannot be empty")
	ErrChatMessageTooLong = errors.New("message exceeds 500 characters")
)

// SendMessage validates, persists, broadcasts a chat message and sends targeted mention notifications.
func (s *WcChatService) SendMessage(userID uuid.UUID, userName, avatarURL, text string, mentions []uuid.UUID) error {
	text = strings.TrimSpace(text)
	if len(text) == 0 {
		return ErrChatMessageEmpty
	}
	if len([]rune(text)) > 500 {
		return ErrChatMessageTooLong
	}

	msg := &model.WcChatMessage{
		UserID:    userID,
		UserName:  userName,
		AvatarURL: avatarURL,
		Message:   text,
		CreatedAt: time.Now().UTC(),
	}
	if err := s.repo.Save(msg); err != nil {
		return err
	}

	// Save mention rows (deduplicate + skip self-mentions)
	uniqueMentions := deduplicateMentions(userID, mentions)
	if len(uniqueMentions) > 0 {
		if err := s.repo.SaveMentions(msg.ID, uniqueMentions); err != nil {
			log.Printf("wc_chat: failed to save mentions: %v", err)
		}
	}

	s.hub.BroadcastChat(ws.ChatMessageEvent{
		Type:      "chat_message",
		ID:        msg.ID.String(),
		UserID:    userID.String(),
		UserName:  userName,
		AvatarURL: avatarURL,
		Message:   text,
		CreatedAt: msg.CreatedAt.Format(time.RFC3339),
	})

	// Send targeted mention notification to each mentioned user
	for _, mentionedID := range uniqueMentions {
		event := ws.ChatMentionEvent{
			Type:       "chat_mention",
			MessageID:  msg.ID.String(),
			SenderID:   userID.String(),
			SenderName: userName,
			Message:    text,
		}
		data, err := json.Marshal(event)
		if err != nil {
			continue
		}
		s.hub.SendToUser(mentionedID, data)
	}

	return nil
}

// ListHistory returns up to limit messages, optionally before a cursor timestamp.
func (s *WcChatService) ListHistory(limit int, before *time.Time) ([]model.WcChatMessage, error) {
	return s.repo.ListMessages(limit, before)
}

// UnreadMentionCount returns the number of unread mentions for a user.
func (s *WcChatService) UnreadMentionCount(userID uuid.UUID) (int64, error) {
	return s.repo.UnreadMentionCount(userID)
}

// MarkMentionsRead marks all unread mentions for a user as read.
func (s *WcChatService) MarkMentionsRead(userID uuid.UUID) error {
	return s.repo.MarkMentionsRead(userID)
}

func deduplicateMentions(senderID uuid.UUID, mentions []uuid.UUID) []uuid.UUID {
	seen := map[uuid.UUID]struct{}{senderID: {}} // skip self-mentions
	out := make([]uuid.UUID, 0, len(mentions))
	for _, id := range mentions {
		if _, exists := seen[id]; !exists {
			seen[id] = struct{}{}
			out = append(out, id)
		}
	}
	return out
}
