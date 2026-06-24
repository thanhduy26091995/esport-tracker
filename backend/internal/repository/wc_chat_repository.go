package repository

import (
	"github.com/duyb/esport-score-tracker/internal/model"
	"gorm.io/gorm"
)

type WcChatRepository struct {
	db *gorm.DB
}

func NewWcChatRepository(db *gorm.DB) *WcChatRepository {
	return &WcChatRepository{db: db}
}

func (r *WcChatRepository) Save(msg *model.WcChatMessage) error {
	return r.db.Create(msg).Error
}

// ListLast100 returns up to 100 messages ordered oldest → newest.
func (r *WcChatRepository) ListLast100() ([]model.WcChatMessage, error) {
	var msgs []model.WcChatMessage
	err := r.db.
		Order("created_at DESC").
		Limit(100).
		Find(&msgs).Error
	if err != nil {
		return nil, err
	}
	// Reverse so oldest is first
	for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	}
	return msgs, nil
}
