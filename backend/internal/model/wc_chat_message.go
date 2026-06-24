package model

import (
	"time"

	"github.com/google/uuid"
)

type WcChatMessage struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null"                               json:"user_id"`
	UserName  string    `gorm:"type:varchar(100);not null"                       json:"user_name"`
	AvatarURL string    `gorm:"type:text"                                        json:"avatar_url"`
	Message   string    `gorm:"type:text;not null;check:char_length(message) >= 1 AND char_length(message) <= 500" json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

func (WcChatMessage) TableName() string {
	return "wc_chat_messages"
}
