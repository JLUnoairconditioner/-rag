// Package service 包含了应用的业务逻辑层。
package service

import (
	"context"
	"errors"
	"fmt"
	"pai-smart-go/internal/model"
	"pai-smart-go/internal/repository"
	"strings"
	"time"

	"gorm.io/gorm"
)

// UserListResponse 定义了用户列表 API 的响应结构。
type UserListResponse struct {
	Content       []UserDetailResponse `json:"content"`
	TotalElements int64                `json:"totalElements"`
	TotalPages    int                  `json:"totalPages"`
	Size          int                  `json:"size"`
	Number        int                  `json:"number"`
}

// UserDetailResponse 定义了用户列表项的详细结构。
type UserDetailResponse struct {
	UserID     uint            `json:"userId"`
	Username   string          `json:"username"`
	Role       string          `json:"role"`
	OrgTags    []OrgTagDetail  `json:"orgTags"`
	PrimaryOrg string          `json:"primaryOrg"`
	Status     int             `json:"status"`
	CreatedAt  model.LocalTime `json:"createdAt"`
}

// OrgTagDetail 定义了组织标签的详细信息。
type OrgTagDetail struct {
	TagID string `json:"tagId"`
	Name  string `json:"name"`
}

// AdminService 接口定义了所有管理员相关的业务操作。
type AdminService interface {
	// Organization Tag Management
	CreateOrganizationTag(tagID, name, description, parentTag string, creator *model.User) (*model.OrganizationTag, error)
	ListOrganizationTags() ([]model.OrganizationTag, error)
	GetOrganizationTagTree() ([]*model.OrganizationTagNode, error)
	UpdateOrganizationTag(tagID string, name, description, parentTag string) (*model.OrganizationTag, error)
	DeleteOrganizationTag(tagID string) error

	// User Management
	AssignOrgTagsToUser(userID uint, orgTags []string) error
	ListUsers(page, size int) (*UserListResponse, error)
	GetAllConversations(ctx context.Context, userID *uint, startTime, endTime *time.Time) ([]map[string]interface{}, error)
}

// adminService 是 AdminService 接口的实现。
type adminService struct {
	orgTagRepo         repository.OrgTagRepository
	userRepo           repository.UserRepository
	conversationRepo   repository.ConversationRepository
	persistentConvRepo repository.PersistentConversationRepository
}

// NewAdminService 创建一个新的 AdminService 实例。
func NewAdminService(orgTagRepo repository.OrgTagRepository, userRepo repository.UserRepository, conversationRepo repository.ConversationRepository, persistentConvRepo repository.PersistentConversationRepository) AdminService {
	return &adminService{
		orgTagRepo:         orgTagRepo,
		userRepo:           userRepo,
		conversationRepo:   conversationRepo,
		persistentConvRepo: persistentConvRepo,
	}
}

// CreateOrganizationTag 处理创建新组织标签的逻辑。
func (s *adminService) CreateOrganizationTag(tagID, name, description, parentTag string, creator *model.User) (*model.OrganizationTag, error) {
	// 检查 TagID 是否已存在
	_, err := s.orgTagRepo.FindByID(tagID)
	if err == nil {
		// 如果 err 为 nil，说明找到了记录，因此 TagID 已存在
		return nil, errors.New("tagID 已存在")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		// 如果是其他类型的数据库错误，则直接返回
		return nil, err
	}

	tag := &model.OrganizationTag{
		TagID:       tagID,
		Name:        name,
		Description: description,
		CreatedBy:   creator.ID,
	}
	if parentTag != "" {
		tag.ParentTag = &parentTag
	}

	if err := s.orgTagRepo.Create(tag); err != nil {
		return nil, err
	}
	return tag, nil
}

// GetOrganizationTagTree retrieves all tags and organizes them into a tree structure.
func (s *adminService) GetOrganizationTagTree() ([]*model.OrganizationTagNode, error) {
	tags, err := s.orgTagRepo.FindAll()
	if err != nil {
		return nil, err
	}

	nodes := make(map[string]*model.OrganizationTagNode)
	var tree []*model.OrganizationTagNode

	for _, tag := range tags {
		nodes[tag.TagID] = &model.OrganizationTagNode{
			TagID:       tag.TagID,
			Name:        tag.Name,
			Description: tag.Description,
			ParentTag:   tag.ParentTag,
			Children:    []*model.OrganizationTagNode{},
		}
	}

	for _, node := range nodes {
		if node.ParentTag != nil && *node.ParentTag != "" {
			if parent, ok := nodes[*node.ParentTag]; ok {
				parent.Children = append(parent.Children, node)
			}
		} else {
			tree = append(tree, node)
		}
	}
	return tree, nil
}

// ListOrganizationTags 返回所有组织标签的列表。
func (s *adminService) ListOrganizationTags() ([]model.OrganizationTag, error) {
	return s.orgTagRepo.FindAll()
}

// UpdateOrganizationTag updates an existing organization tag.
func (s *adminService) UpdateOrganizationTag(tagID string, name, description, parentTag string) (*model.OrganizationTag, error) {
	tag, err := s.orgTagRepo.FindByID(tagID)
	if err != nil {
		return nil, errors.New("tag not found")
	}

	tag.Name = name
	tag.Description = description
	if parentTag != "" {
		tag.ParentTag = &parentTag
	} else {
		tag.ParentTag = nil
	}

	if err := s.orgTagRepo.Update(tag); err != nil {
		return nil, err
	}
	return tag, nil
}

// DeleteOrganizationTag deletes an organization tag by its ID.
func (s *adminService) DeleteOrganizationTag(tagID string) error {
	return s.orgTagRepo.Delete(tagID)
}

// AssignOrgTagsToUser 为指定用户分配一组组织标签。
func (s *adminService) AssignOrgTagsToUser(userID uint, orgTags []string) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return err
	}
	user.OrgTags = strings.Join(orgTags, ",")
	return s.userRepo.Update(user)
}

// ListUsers 以分页的形式返回用户列表
func (s *adminService) ListUsers(page, size int) (*UserListResponse, error) {
	offset := (page - 1) * size
	users, total, err := s.userRepo.FindWithPagination(offset, size)
	if err != nil {
		return nil, err
	}

	var userResponses []UserDetailResponse
	for _, u := range users {
		// 获取组织标签详情
		orgTagDetails := make([]OrgTagDetail, 0) // 初始化为空数组，而不是 nil
		if u.OrgTags != "" {
			tagIDs := strings.Split(u.OrgTags, ",")
			for _, tagID := range tagIDs {
				tag, err := s.orgTagRepo.FindByID(tagID)
				if err != nil { // 忽略找不到的标签
					continue
				}
				orgTagDetails = append(orgTagDetails, OrgTagDetail{
					TagID: tag.TagID,
					Name:  tag.Name,
				})
			}
		}

		// 转换角色为状态码
		status := 1 // 默认为 USER
		if u.Role == "ADMIN" {
			status = 0
		}

		userResponses = append(userResponses, UserDetailResponse{
			UserID:     u.ID,
			Username:   u.Username,
			Role:       u.Role,
			OrgTags:    orgTagDetails,
			PrimaryOrg: u.PrimaryOrg,
			Status:     status,
			CreatedAt:  model.LocalTime(u.CreatedAt),
		})
	}

	totalPages := 0
	if total > 0 && size > 0 {
		totalPages = (int(total) + size - 1) / size
	}

	response := &UserListResponse{
		Content:       userResponses,
		TotalElements: total,
		TotalPages:    totalPages,
		Size:          size,
		Number:        page,
	}
	return response, nil
}

// GetAllConversations retrieves conversation histories for all or a specific user, with optional date filtering.
func (s *adminService) GetAllConversations(ctx context.Context, userID *uint, startTime, endTime *time.Time) ([]map[string]interface{}, error) {
	var records []*model.Conversation
	var err error

	if userID != nil {
		if _, err = s.userRepo.FindByID(*userID); err != nil {
			return nil, errors.New("user not found")
		}
		records, err = s.persistentConvRepo.FindByUserID(ctx, *userID, startTime, endTime)
	} else {
		records, err = s.persistentConvRepo.FindAll(ctx, startTime, endTime)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query conversations: %w", err)
	}

	// 批量查询用户名
	userIDSet := make(map[uint]struct{})
	for _, r := range records {
		userIDSet[r.UserID] = struct{}{}
	}
	usernameMap := make(map[uint]string)
	for uid := range userIDSet {
		u, ferr := s.userRepo.FindByID(uid)
		if ferr == nil {
			usernameMap[uid] = u.Username
		}
	}

	result := make([]map[string]interface{}, 0, len(records)*2)
	for _, r := range records {
		username := usernameMap[r.UserID]
		ts := r.CreatedAt.Format("2006-01-02T15:04:05")
		result = append(result,
			map[string]interface{}{"username": username, "role": "user", "content": r.Question, "timestamp": ts},
			map[string]interface{}{"username": username, "role": "assistant", "content": r.Answer, "timestamp": ts},
		)
	}
	return result, nil
}
