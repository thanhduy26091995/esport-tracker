package ws

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// TokenVerifier validates a JWT string and returns the user's identity.
type TokenVerifier interface {
	VerifyToken(tokenStr string) (userID uuid.UUID, userName string, err error)
}

// UserAvatarFetcher returns the avatar URL for a given user ID.
type UserAvatarFetcher interface {
	GetAvatarURL(userID uuid.UUID) string
}

// ChatHandler upgrades HTTP connections to WebSocket for the live chat room.
type ChatHandler struct {
	hub          *Hub
	tokenVerify  TokenVerifier
	avatarFetch  UserAvatarFetcher
	chatService  ChatService
}

// ChatService is the subset of WcChatService used by ChatHandler.
type ChatService interface {
	SendMessage(userID uuid.UUID, userName, avatarURL, text string) error
}

func NewChatHandler(hub *Hub, tokenVerify TokenVerifier, avatarFetch UserAvatarFetcher, chatSvc ChatService) *ChatHandler {
	return &ChatHandler{
		hub:         hub,
		tokenVerify: tokenVerify,
		avatarFetch: avatarFetch,
		chatService: chatSvc,
	}
}

func (h *ChatHandler) Handle(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("ws/chat: upgrade error: %v", err)
		return
	}

	cl := &client{
		conn: conn,
		send: make(chan []byte, 64),
	}

	// Authenticate via optional ?token= query param
	if token := c.Query("token"); token != "" {
		userID, userName, verifyErr := h.tokenVerify.VerifyToken(token)
		if verifyErr != nil {
			conn.WriteMessage(websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, "invalid token"))
			conn.Close()
			return
		}
		cl.userID = userID
		cl.userName = userName
		cl.avatarURL = h.avatarFetch.GetAvatarURL(userID)
	}

	h.hub.register <- cl

	go cl.writePump()
	go h.chatReadPump(cl)
}

func (h *ChatHandler) chatReadPump(cl *client) {
	defer func() {
		h.hub.unregister <- cl
		cl.conn.Close()
	}()

	cl.conn.SetReadLimit(maxMessageSize * 4) // chat messages up to ~2KB
	cl.conn.SetReadDeadline(time.Now().Add(pongWait))
	cl.conn.SetPongHandler(func(string) error {
		cl.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, raw, err := cl.conn.ReadMessage()
		if err != nil {
			break
		}

		var frame ChatSendFrame
		if jsonErr := json.Unmarshal(raw, &frame); jsonErr != nil || frame.Type != "chat_send" {
			continue
		}

		// Guest clients cannot send messages
		if cl.userID == (uuid.UUID{}) {
			h.sendError(cl, "unauthenticated")
			continue
		}

		if svcErr := h.chatService.SendMessage(cl.userID, cl.userName, cl.avatarURL, frame.Message); svcErr != nil {
			h.sendError(cl, svcErr.Error())
		}
	}
}

func (h *ChatHandler) sendError(cl *client, msg string) {
	data, _ := json.Marshal(ChatErrorFrame{Type: "error", Message: msg})
	select {
	case cl.send <- data:
	default:
	}
}

// WcAuthTokenVerifier adapts WcAuthService to the TokenVerifier interface.
type WcAuthTokenVerifier struct {
	verify func(string) (uuid.UUID, string, error)
}

func NewWcAuthTokenVerifier(fn func(string) (uuid.UUID, string, error)) *WcAuthTokenVerifier {
	return &WcAuthTokenVerifier{verify: fn}
}

func (v *WcAuthTokenVerifier) VerifyToken(tokenStr string) (uuid.UUID, string, error) {
	return v.verify(tokenStr)
}

// WcUserAvatarFetcher adapts WcUserRepository to the UserAvatarFetcher interface.
type WcUserAvatarFetcher struct {
	fetch func(uuid.UUID) string
}

func NewWcUserAvatarFetcher(fn func(uuid.UUID) string) *WcUserAvatarFetcher {
	return &WcUserAvatarFetcher{fetch: fn}
}

func (f *WcUserAvatarFetcher) GetAvatarURL(userID uuid.UUID) string {
	return f.fetch(userID)
}

