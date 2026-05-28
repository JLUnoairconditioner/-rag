package repository

import (
	"context"
	"pai-smart-go/internal/model"
	"time"

	"gorm.io/gorm"
)

// PersistentConversationRepository 提供对话记录的 MySQL 持久化操作。
type PersistentConversationRepository interface {
	Save(ctx context.Context, userID uint, question, answer string) error
	FindByUserID(ctx context.Context, userID uint, startTime, endTime *time.Time) ([]*model.Conversation, error)
	FindAll(ctx context.Context, startTime, endTime *time.Time) ([]*model.Conversation, error)
}

type mysqlConversationRepository struct {
	db *gorm.DB
}

func NewPersistentConversationRepository(db *gorm.DB) PersistentConversationRepository {
	return &mysqlConversationRepository{db: db}
}

func (r *mysqlConversationRepository) Save(ctx context.Context, userID uint, question, answer string) error {
	record := &model.Conversation{
		UserID:   userID,
		Question: question,
		Answer:   answer,
	}
	return r.db.WithContext(ctx).Create(record).Error
}

func (r *mysqlConversationRepository) FindByUserID(ctx context.Context, userID uint, startTime, endTime *time.Time) ([]*model.Conversation, error) {
	q := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at ASC")
	if startTime != nil {
		q = q.Where("created_at >= ?", *startTime)
	}
	if endTime != nil {
		q = q.Where("created_at <= ?", *endTime)
	}
	var records []*model.Conversation
	return records, q.Find(&records).Error
}

func (r *mysqlConversationRepository) FindAll(ctx context.Context, startTime, endTime *time.Time) ([]*model.Conversation, error) {
	q := r.db.WithContext(ctx).Order("created_at ASC")
	if startTime != nil {
		q = q.Where("created_at >= ?", *startTime)
	}
	if endTime != nil {
		q = q.Where("created_at <= ?", *endTime)
	}
	var records []*model.Conversation
	return records, q.Find(&records).Error
}
