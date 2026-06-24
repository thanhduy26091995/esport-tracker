package service

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/duyb/esport-score-tracker/internal/model"
	"github.com/duyb/esport-score-tracker/internal/repository"
	"github.com/duyb/esport-score-tracker/internal/ws"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type WcCustomBetService struct {
	repo     *repository.WcCustomBetRepository
	wcRepo   *repository.WcRepository
	userRepo *repository.WcUserRepository
	hub      ws.HubBroadcaster
}

func NewWcCustomBetService(repo *repository.WcCustomBetRepository, wcRepo *repository.WcRepository, userRepo *repository.WcUserRepository, hub ws.HubBroadcaster) *WcCustomBetService {
	return &WcCustomBetService{repo: repo, wcRepo: wcRepo, userRepo: userRepo, hub: hub}
}

type CreateCustomBetOption struct {
	Label        string  `json:"label"`
	Odds         float64 `json:"odds"`
	DisplayOrder int     `json:"display_order"`
}

func (s *WcCustomBetService) CreateCustomBet(matchID, adminID uuid.UUID, title string, line *float64, opts []CreateCustomBetOption) (*model.WcCustomBetWithOptions, error) {
	if len(opts) < 2 || len(opts) > 10 {
		return nil, fmt.Errorf("cần từ 2 đến 10 lựa chọn")
	}
	for _, o := range opts {
		if o.Label == "" {
			return nil, fmt.Errorf("label không được để trống")
		}
		if o.Odds <= 0 {
			return nil, fmt.Errorf("odds phải > 0")
		}
	}
	bet := &model.WcCustomBet{
		MatchID:   matchID,
		Title:     title,
		Line:      line,
		Status:    model.WcCustomBetStatusOpen,
		CreatedBy: &adminID,
	}
	options := make([]model.WcCustomBetOption, len(opts))
	for i, o := range opts {
		options[i] = model.WcCustomBetOption{
			Label:        o.Label,
			Odds:         o.Odds,
			DisplayOrder: o.DisplayOrder,
		}
	}
	if err := s.repo.Create(bet, options); err != nil {
		return nil, fmt.Errorf("failed to create custom bet: %w", err)
	}
	result, err := s.repo.GetOptions(bet.ID)
	if err != nil {
		return nil, err
	}
	return &model.WcCustomBetWithOptions{WcCustomBet: *bet, Options: result}, nil
}

func (s *WcCustomBetService) UpdateCustomBet(betID uuid.UUID, title *string, line *float64, status *string) (*model.WcCustomBet, error) {
	bet, err := s.repo.GetByID(betID)
	if err != nil {
		return nil, fmt.Errorf("kèo không tồn tại")
	}
	if bet.Status == model.WcCustomBetStatusSettled || bet.Status == model.WcCustomBetStatusVoid {
		return nil, fmt.Errorf("không thể cập nhật kèo đã tất toán hoặc đã huỷ")
	}
	fields := map[string]interface{}{}
	if title != nil && *title != "" {
		fields["title"] = *title
	}
	if line != nil {
		fields["line"] = *line
	}
	if status != nil {
		allowed := map[string]bool{"open": true, "closed": true}
		if !allowed[*status] {
			return nil, fmt.Errorf("status phải là 'open' hoặc 'closed'")
		}
		fields["status"] = *status
	}
	if len(fields) == 0 {
		return bet, nil
	}
	if err := s.repo.UpdateBet(nil, betID, fields); err != nil {
		return nil, err
	}
	return s.repo.GetByID(betID)
}

func (s *WcCustomBetService) ListForMatchAdmin(matchID uuid.UUID) ([]*model.WcCustomBetWithOptions, error) {
	return s.repo.ListForMatchAdmin(matchID)
}

func (s *WcCustomBetService) ListForMatchPlayer(matchID, userID uuid.UUID) ([]*model.WcCustomBetWithOptions, error) {
	return s.repo.ListForMatchWithMyEntry(matchID, userID)
}

func (s *WcCustomBetService) GetMyEntries(userID uuid.UUID) ([]model.WcCustomBetEntryHistory, error) {
	return s.repo.GetEntriesForUser(userID)
}

func (s *WcCustomBetService) PlaceEntry(betID, userID, optionID uuid.UUID, stake int, userName string) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return fmt.Errorf("user not found")
	}
	if user.IsBlocked {
		return fmt.Errorf("user is blocked from placing bets")
	}
	cfg, err := s.wcRepo.GetConfig()
	if err != nil {
		return fmt.Errorf("failed to load config")
	}
	if stake < cfg.MinPoints || stake > cfg.MaxPoints {
		return fmt.Errorf("điểm cược phải từ %d đến %d", cfg.MinPoints, cfg.MaxPoints)
	}
	bet, err := s.repo.GetByID(betID)
	if err != nil {
		return fmt.Errorf("kèo không tồn tại")
	}
	if bet.Status != model.WcCustomBetStatusOpen {
		return fmt.Errorf("kèo đã đóng")
	}
	opt, err := s.repo.GetOptionByID(optionID)
	if err != nil || opt.CustomBetID != betID {
		return fmt.Errorf("lựa chọn không hợp lệ")
	}
	entry := &model.WcCustomBetEntry{
		CustomBetID:  betID,
		OptionID:     optionID,
		WcUserID:     userID,
		Stake:        stake,
		OddsSnapshot: opt.Odds,
		Status:       model.WcCustomBetEntryStatusPending,
	}
	db := s.wcRepo.DB()
	if err := db.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.CreateEntry(tx, entry); err != nil {
			if strings.Contains(err.Error(), "idx_custom_bet_entry_dedup") {
				return fmt.Errorf("bạn đã đặt cược cho kèo này rồi")
			}
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	if s.hub != nil {
		s.hub.Broadcast(ws.ActivityEvent{
			Type:      "bet_placed",
			UserID:    userID.String(),
			UserName:  userName,
			BetType:   "custom",
			Selection: bet.Title + " - " + opt.Label,
			Stake:     stake,
			MatchID:   bet.MatchID.String(),
		})
	}

	return nil
}

func (s *WcCustomBetService) CancelEntry(entryID, userID uuid.UUID) error {
	entry, err := s.repo.GetEntry(entryID)
	if err != nil {
		return fmt.Errorf("không tìm thấy cược")
	}
	if entry.WcUserID != userID {
		return fmt.Errorf("không có quyền huỷ cược này")
	}
	if entry.Status != model.WcCustomBetEntryStatusPending {
		return fmt.Errorf("chỉ có thể huỷ cược đang chờ kết quả")
	}
	bet, err := s.repo.GetByID(entry.CustomBetID)
	if err != nil {
		return fmt.Errorf("kèo không tồn tại")
	}
	if bet.Status != model.WcCustomBetStatusOpen {
		return fmt.Errorf("kèo đã đóng, không thể huỷ")
	}
	db := s.wcRepo.DB()
	return db.Transaction(func(tx *gorm.DB) error {
		return s.repo.DeleteEntry(tx, entryID)
	})
}

func (s *WcCustomBetService) Settle(betID, winningOptionID, adminID uuid.UUID) error {
	bet, err := s.repo.GetByID(betID)
	if err != nil {
		return fmt.Errorf("kèo không tồn tại")
	}
	if bet.Status != model.WcCustomBetStatusOpen && bet.Status != model.WcCustomBetStatusClosed {
		return fmt.Errorf("kèo đã tất toán hoặc đã huỷ")
	}
	opt, err := s.repo.GetOptionByID(winningOptionID)
	if err != nil || opt.CustomBetID != betID {
		return fmt.Errorf("lựa chọn thắng không hợp lệ")
	}

	db := s.wcRepo.DB()
	return db.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.MarkOptionWinner(tx, winningOptionID); err != nil {
			return err
		}
		entries, err := s.repo.GetPendingEntries(tx, betID)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			if entry.OptionID == winningOptionID {
				payout := math.Round(float64(entry.Stake)*entry.OddsSnapshot*100) / 100
				if err := s.repo.UpdateEntryResult(tx, entry.ID, model.WcCustomBetEntryStatusWon, payout); err != nil {
					return err
				}
				// Credit net profit only (stake was not deducted at placement)
				netChange := payout - float64(entry.Stake)
				if err := s.wcRepo.UpdateWalletBalance(tx, entry.WcUserID, netChange); err != nil {
					return err
				}
			} else {
				if err := s.repo.UpdateEntryResult(tx, entry.ID, model.WcCustomBetEntryStatusLost, 0); err != nil {
					return err
				}
				// Deduct stake at settlement (matches betting system)
				if err := s.wcRepo.UpdateWalletBalance(tx, entry.WcUserID, -float64(entry.Stake)); err != nil {
					return err
				}
			}
		}
		now := time.Now()
		return s.repo.UpdateBet(tx, betID, map[string]interface{}{
			"status":     model.WcCustomBetStatusSettled,
			"settled_at": &now,
			"settled_by": &adminID,
		})
	})
}

func (s *WcCustomBetService) VoidBet(betID uuid.UUID) error {
	bet, err := s.repo.GetByID(betID)
	if err != nil {
		return fmt.Errorf("kèo không tồn tại")
	}
	if bet.Status == model.WcCustomBetStatusVoid {
		return fmt.Errorf("kèo đã huỷ rồi")
	}
	if bet.Status == model.WcCustomBetStatusSettled {
		return fmt.Errorf("kèo đã tất toán, không thể huỷ")
	}

	db := s.wcRepo.DB()
	return db.Transaction(func(tx *gorm.DB) error {
		entries, err := s.repo.GetPendingEntries(tx, betID)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			if err := s.repo.UpdateEntryResult(tx, entry.ID, model.WcCustomBetEntryStatusVoid, 0); err != nil {
				return err
			}
			// No wallet change needed — stake was never deducted at placement
		}
		return s.repo.UpdateBet(tx, betID, map[string]interface{}{
			"status": model.WcCustomBetStatusVoid,
		})
	})
}
