package repository

import (
	"time"

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

// ListMessages returns up to limit messages ordered oldest → newest.
// If before is non-nil, only messages with created_at < before are returned (cursor pagination).
func (r *WcChatRepository) ListMessages(limit int, before *time.Time) ([]model.WcChatMessage, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	q := r.db.Order("created_at DESC").Limit(limit)
	if before != nil {
		q = q.Where("created_at < ?", before)
	}
	var msgs []model.WcChatMessage
	if err := q.Find(&msgs).Error; err != nil {
		return nil, err
	}
	// Reverse so oldest is first in the returned slice
	for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	}
	return msgs, nil
}
