// Package service 包含了应用的业务逻辑层。
package service

import (
	"context"
	"pai-smart-go/internal/model"
	"pai-smart-go/internal/repository"
)

// ConversationService 定义了对话业务逻辑的接口。
type ConversationService interface {
	GetConversationHistory(ctx context.Context, userID uint) ([]model.ChatMessage, error)
	AddMessageToConversation(ctx context.Context, userID uint, message model.ChatMessage) error
}

type conversationService struct {
	repo           repository.ConversationRepository
	persistentRepo repository.PersistentConversationRepository
}

// NewConversationService 创建一个新的 ConversationService。
func NewConversationService(repo repository.ConversationRepository, persistentRepo repository.PersistentConversationRepository) ConversationService {
	return &conversationService{repo: repo, persistentRepo: persistentRepo}
}

// GetConversationHistory 从 MySQL 读取用户全量历史并转换为 ChatMessage 列表。
func (s *conversationService) GetConversationHistory(ctx context.Context, userID uint) ([]model.ChatMessage, error) {
	records, err := s.persistentRepo.FindByUserID(ctx, userID, nil, nil)
	if err != nil {
		return nil, err
	}
	messages := make([]model.ChatMessage, 0, len(records)*2)
	for _, r := range records {
		messages = append(messages,
			model.ChatMessage{Role: "user", Content: r.Question, Timestamp: r.CreatedAt},
			model.ChatMessage{Role: "assistant", Content: r.Answer, Timestamp: r.CreatedAt},
		)
	}
	return messages, nil
}

// AddMessageToConversation 将一条消息添加到用户的对话历史中（Redis 上下文窗口）。
func (s *conversationService) AddMessageToConversation(ctx context.Context, userID uint, message model.ChatMessage) error {
	conversationID, err := s.repo.GetOrCreateConversationID(ctx, userID)
	if err != nil {
		return err
	}
	history, err := s.repo.GetConversationHistory(ctx, conversationID)
	if err != nil {
		return err
	}
	history = append(history, message)
	return s.repo.UpdateConversationHistory(ctx, conversationID, history)
}
