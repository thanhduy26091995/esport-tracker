package ws

import (
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// ActivityEvent is the JSON payload pushed to all connected clients when a bet is placed.
type ActivityEvent struct {
	Type      string `json:"type"`       // "bet_placed"
	UserID    string `json:"user_id"`    // sender's ID — clients use this for self-suppression
	UserName  string `json:"user_name"`
	BetType   string `json:"bet_type"`   // "handicap" | "exact_score" | "ou"
	Selection string `json:"selection"`  // e.g. "Bồ Đào Nha", "1 - 0", "Tài"
	Stake     int    `json:"stake"`
	MatchID   string `json:"match_id"`
	TeamHome  string `json:"team_home"`
	TeamAway  string `json:"team_away"`
}

// HubBroadcaster is the interface WcService depends on — allows injection of a no-op in tests.
type HubBroadcaster interface {
	Broadcast(event ActivityEvent)
}

// ChatSendFrame is a client→server frame for chat messages.
type ChatSendFrame struct {
	Type    string `json:"type"`    // "chat_send"
	Message string `json:"message"` // 1–500 chars
}

// ChatMessageEvent is the server→all-clients broadcast payload for a chat message.
type ChatMessageEvent struct {
	Type      string `json:"type"`       // "chat_message"
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	UserName  string `json:"user_name"`
	AvatarURL string `json:"avatar_url"`
	Message   string `json:"message"`
	CreatedAt string `json:"created_at"` // RFC3339
}

// ChatErrorFrame is sent back to a client when its request cannot be processed.
type ChatErrorFrame struct {
	Type    string `json:"type"`    // "error"
	Message string `json:"message"`
}

// client represents one connected WebSocket peer.
type client struct {
	conn      *websocket.Conn
	send      chan []byte
	userID    uuid.UUID // zero value = guest
	userName  string
	avatarURL string
}
