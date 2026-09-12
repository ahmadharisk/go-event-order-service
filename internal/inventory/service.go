package inventory

import "context"

type Service struct{ repo *Repo }

func NewService(repo *Repo) *Service { return &Service{repo: repo} }

func (s *Service) Reserve(ctx context.Context, orderID string) error { return s.repo.Reserve(ctx, orderID) }
func (s *Service) Release(ctx context.Context, orderID string) error { return s.repo.Release(ctx, orderID) }
