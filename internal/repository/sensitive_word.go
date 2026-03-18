package repository

import (
	"context"

	"swd-new/internal/model"

	"go.uber.org/zap"
)

type SensitiveWordRepository interface {
	List(ctx context.Context) ([]model.SensitiveWord, error)
}

type sensitiveWordRepository struct {
	*Repository
}

func NewSensitiveWordRepository(repository *Repository) (SensitiveWordRepository, error) {
	return &sensitiveWordRepository{
		Repository: repository,
	}, nil
}

func (r *sensitiveWordRepository) List(ctx context.Context) ([]model.SensitiveWord, error) {
	words := make([]model.SensitiveWord, 0, 1024)
	if err := r.db.WithContext(ctx).Order("id ASC").Find(&words).Error; err != nil {
		return nil, err
	}

	r.logger.Info("sensitive words loaded", zap.Int("count", len(words)))
	return words, nil
}
