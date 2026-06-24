package service_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/duyb/esport-score-tracker/internal/model"
	"github.com/duyb/esport-score-tracker/internal/service"
	"github.com/duyb/esport-score-tracker/internal/ws"
	"github.com/google/uuid"
)

// --- fakes ---

type fakeChatRepo struct {
	saved   *model.WcChatMessage
	saveErr error
	history []model.WcChatMessage
}

func (r *fakeChatRepo) Save(msg *model.WcChatMessage) error {
	if r.saveErr != nil {
		return r.saveErr
	}
	msg.ID = uuid.New()
	r.saved = msg
	return nil
}

func (r *fakeChatRepo) ListMessages(_ int, _ *time.Time) ([]model.WcChatMessage, error) {
	return r.history, nil
}

type fakeChatHub struct {
	broadcasted []ws.ChatMessageEvent
}

func (h *fakeChatHub) BroadcastChat(event ws.ChatMessageEvent) {
	h.broadcasted = append(h.broadcasted, event)
}

// --- tests ---

func TestWcChatService_SendMessage_Valid(t *testing.T) {
	repo := &fakeChatRepo{}
	hub := &fakeChatHub{}
	svc := service.NewWcChatService(repo, hub)

	userID := uuid.New()
	err := svc.SendMessage(userID, "Alice", "https://avatar.url/a.png", "Hello!")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.saved == nil {
		t.Fatal("expected message to be saved")
	}
	if repo.saved.Message != "Hello!" {
		t.Errorf("saved message = %q, want %q", repo.saved.Message, "Hello!")
	}
	if len(hub.broadcasted) != 1 {
		t.Fatalf("expected 1 broadcast, got %d", len(hub.broadcasted))
	}
	if hub.broadcasted[0].Type != "chat_message" {
		t.Errorf("broadcast type = %q, want %q", hub.broadcasted[0].Type, "chat_message")
	}
}

func TestWcChatService_SendMessage_Empty(t *testing.T) {
	svc := service.NewWcChatService(&fakeChatRepo{}, &fakeChatHub{})
	err := svc.SendMessage(uuid.New(), "Alice", "", "   ")
	if !errors.Is(err, service.ErrChatMessageEmpty) {
		t.Errorf("expected ErrChatMessageEmpty, got %v", err)
	}
}

func TestWcChatService_SendMessage_TooLong(t *testing.T) {
	svc := service.NewWcChatService(&fakeChatRepo{}, &fakeChatHub{})
	long := strings.Repeat("a", 501)
	err := svc.SendMessage(uuid.New(), "Alice", "", long)
	if !errors.Is(err, service.ErrChatMessageTooLong) {
		t.Errorf("expected ErrChatMessageTooLong, got %v", err)
	}
}

func TestWcChatService_SendMessage_DBError(t *testing.T) {
	dbErr := errors.New("db down")
	repo := &fakeChatRepo{saveErr: dbErr}
	hub := &fakeChatHub{}
	svc := service.NewWcChatService(repo, hub)

	err := svc.SendMessage(uuid.New(), "Alice", "", "Hello!")
	if !errors.Is(err, dbErr) {
		t.Errorf("expected db error, got %v", err)
	}
	if len(hub.broadcasted) != 0 {
		t.Error("should not broadcast on DB error")
	}
}
