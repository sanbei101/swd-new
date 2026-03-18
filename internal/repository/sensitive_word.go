package repository

import (
	"context"
	"fmt"
	"time"

	"swd-new/internal/model"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type SensitiveWordRepository interface {
	List(ctx context.Context) ([]model.SensitiveWord, error)
}

type sensitiveWordRepository struct {
	*Repository
	db *gorm.DB
}

func NewSensitiveWordRepository(repository *Repository, conf *viper.Viper) (SensitiveWordRepository, error) {
	dsn := conf.GetString("data.postgres.dsn")
	if dsn == "" {
		return nil, fmt.Errorf("data.postgres.dsn is required")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, err
	}

	repository.logger.Info("sensitive words postgres connected")

	return &sensitiveWordRepository{
		Repository: repository,
		db:         db,
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
