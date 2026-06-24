package service_test

import (
	"encoding/json"
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
	saved          *model.WcChatMessage
	saveErr        error
	history        []model.WcChatMessage
	savedMentions  []uuid.UUID
	unreadCount    int64
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

func (r *fakeChatRepo) SaveMentions(_ uuid.UUID, userIDs []uuid.UUID) error {
	r.savedMentions = append(r.savedMentions, userIDs...)
	return nil
}

func (r *fakeChatRepo) UnreadMentionCount(_ uuid.UUID) (int64, error) {
	return r.unreadCount, nil
}

func (r *fakeChatRepo) MarkMentionsRead(_ uuid.UUID) error {
	return nil
}

type fakeChatHub struct {
	broadcasted []ws.ChatMessageEvent
	sentToUser  map[string][][]byte
}

func (h *fakeChatHub) BroadcastChat(event ws.ChatMessageEvent) {
	h.broadcasted = append(h.broadcasted, event)
}

func (h *fakeChatHub) SendToUser(userID uuid.UUID, data []byte) {
	if h.sentToUser == nil {
		h.sentToUser = make(map[string][][]byte)
	}
	h.sentToUser[userID.String()] = append(h.sentToUser[userID.String()], data)
}

// --- tests ---

func TestWcChatService_SendMessage_Valid(t *testing.T) {
	repo := &fakeChatRepo{}
	hub := &fakeChatHub{}
	svc := service.NewWcChatService(repo, hub)

	userID := uuid.New()
	err := svc.SendMessage(userID, "Alice", "https://avatar.url/a.png", "Hello!", nil)
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
	err := svc.SendMessage(uuid.New(), "Alice", "", "   ", nil)
	if !errors.Is(err, service.ErrChatMessageEmpty) {
		t.Errorf("expected ErrChatMessageEmpty, got %v", err)
	}
}

func TestWcChatService_SendMessage_TooLong(t *testing.T) {
	svc := service.NewWcChatService(&fakeChatRepo{}, &fakeChatHub{})
	long := strings.Repeat("a", 501)
	err := svc.SendMessage(uuid.New(), "Alice", "", long, nil)
	if !errors.Is(err, service.ErrChatMessageTooLong) {
		t.Errorf("expected ErrChatMessageTooLong, got %v", err)
	}
}

func TestWcChatService_SendMessage_DBError(t *testing.T) {
	dbErr := errors.New("db down")
	repo := &fakeChatRepo{saveErr: dbErr}
	hub := &fakeChatHub{}
	svc := service.NewWcChatService(repo, hub)

	err := svc.SendMessage(uuid.New(), "Alice", "", "Hello!", nil)
	if !errors.Is(err, dbErr) {
		t.Errorf("expected db error, got %v", err)
	}
	if len(hub.broadcasted) != 0 {
		t.Error("should not broadcast on DB error")
	}
}

func TestWcChatService_SendMessage_WithMentions(t *testing.T) {
	repo := &fakeChatRepo{}
	hub := &fakeChatHub{}
	svc := service.NewWcChatService(repo, hub)

	senderID := uuid.New()
	bobID := uuid.New()
	carolID := uuid.New()

	err := svc.SendMessage(senderID, "Alice", "", "Hey @Bob and @Carol!", []uuid.UUID{bobID, carolID})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repo.savedMentions) != 2 {
		t.Fatalf("expected 2 mention rows, got %d", len(repo.savedMentions))
	}
	// Both mentioned users should have received a targeted WS frame
	if len(hub.sentToUser) != 2 {
		t.Errorf("expected 2 users to receive WS frames, got %d", len(hub.sentToUser))
	}
	for _, msgs := range hub.sentToUser {
		var event ws.ChatMentionEvent
		if err := json.Unmarshal(msgs[0], &event); err != nil {
			t.Fatalf("failed to unmarshal mention event: %v", err)
		}
		if event.Type != "chat_mention" {
			t.Errorf("event type = %q, want %q", event.Type, "chat_mention")
		}
		if event.SenderName != "Alice" {
			t.Errorf("sender name = %q, want Alice", event.SenderName)
		}
	}
}

func TestWcChatService_SendMessage_SelfMentionSkipped(t *testing.T) {
	repo := &fakeChatRepo{}
	hub := &fakeChatHub{}
	svc := service.NewWcChatService(repo, hub)

	senderID := uuid.New()
	err := svc.SendMessage(senderID, "Alice", "", "Hey @Alice!", []uuid.UUID{senderID})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Self-mention should be silently dropped
	if len(repo.savedMentions) != 0 {
		t.Errorf("expected 0 mention rows (self-mention skipped), got %d", len(repo.savedMentions))
	}
	if len(hub.sentToUser) != 0 {
		t.Errorf("expected no WS frames sent for self-mention")
	}
}
