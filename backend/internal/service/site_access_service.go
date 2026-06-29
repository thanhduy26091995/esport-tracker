package service

import (
	"crypto/sha256"
	"fmt"
	"strings"

	"github.com/duyb/esport-score-tracker/internal/repository"
)

type SiteAccessService struct {
	repo *repository.SiteAccessRepository
}

func NewSiteAccessService(repo *repository.SiteAccessRepository) *SiteAccessService {
	return &SiteAccessService{repo: repo}
}

func hashSiteAnswer(answer string) string {
	h := sha256.Sum256([]byte(strings.TrimSpace(strings.ToLower(answer))))
	return fmt.Sprintf("%x", h)
}

func (s *SiteAccessService) GetQuestion() (question string, enabled bool, err error) {
	cfg, err := s.repo.Get()
	if err != nil {
		return "", false, err
	}
	return cfg.Question, cfg.Enabled, nil
}

// Validate checks the answer and returns the token (= answer hash) if correct.
func (s *SiteAccessService) Validate(answer string) (string, error) {
	cfg, err := s.repo.Get()
	if err != nil {
		return "", err
	}
	if !cfg.Enabled {
		return "", nil
	}
	if hashSiteAnswer(answer) != cfg.AnswerHash {
		return "", fmt.Errorf("incorrect answer")
	}
	return cfg.AnswerHash, nil
}

// Update saves question, answer (optional), and enabled flag.
// If answer is empty, keeps the existing hash.
func (s *SiteAccessService) Update(question, answer string, enabled bool) error {
	answerHash := ""
	if answer != "" {
		answerHash = hashSiteAnswer(answer)
	} else {
		if cfg, err := s.repo.Get(); err == nil {
			answerHash = cfg.AnswerHash
		}
	}
	return s.repo.Save(question, answerHash, enabled)
}
