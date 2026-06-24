package service

import (
	"errors"
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
}

// ChatHubBroadcaster is the interface WcChatService depends on for broadcasting.
type ChatHubBroadcaster interface {
	BroadcastChat(event ws.ChatMessageEvent)
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

// SendMessage validates, persists, and broadcasts a chat message.
func (s *WcChatService) SendMessage(userID uuid.UUID, userName, avatarURL, text string) error {
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

	s.hub.BroadcastChat(ws.ChatMessageEvent{
		Type:      "chat_message",
		ID:        msg.ID.String(),
		UserID:    userID.String(),
		UserName:  userName,
		AvatarURL: avatarURL,
		Message:   text,
		CreatedAt: msg.CreatedAt.Format(time.RFC3339),
	})
	return nil
}

// ListHistory returns up to limit messages, optionally before a cursor timestamp.
func (s *WcChatService) ListHistory(limit int, before *time.Time) ([]model.WcChatMessage, error) {
	return s.repo.ListMessages(limit, before)
}
