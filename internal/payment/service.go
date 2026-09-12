package payment

import (
	"context"
	"math/rand"
)

type Service struct{ repo *Repo }

func NewService(repo *Repo) *Service { return &Service{repo: repo} }

func (s *Service) Process(ctx context.Context, orderID string) (string, error) {
	amt, err := s.repo.FindOrderAmount(ctx, orderID)
	if err != nil {
		return "failed", err
	}
	status := "succeeded"
	if rand.Float64() < 0.15 {
		status = "failed"
	}
	if err := s.repo.Insert(ctx, orderID, amt, status); err != nil {
		return status, err
	}
	return status, nil
}
