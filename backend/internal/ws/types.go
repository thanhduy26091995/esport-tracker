package ws

import "github.com/gorilla/websocket"

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

// client represents one connected WebSocket peer.
type client struct {
	conn *websocket.Conn
	send chan []byte
}
