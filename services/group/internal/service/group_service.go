package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"

	"github.com/adk-saugat/rally/services/group/internal/repository"
	"github.com/adk-saugat/rally/shared/types"
)

var (
	ErrInvalidInput  = errors.New("invalid input")
	ErrUnauthorized  = errors.New("user id is required")
	ErrNotFound      = repository.ErrNotFound
	ErrAlreadyMember    = repository.ErrAlreadyMember
	ErrPermissionDenied = errors.New("not a member of this group")
)

type GroupService interface {
	CreateGroup(ctx context.Context, name, userID string) (*types.Group, error)
	JoinGroup(ctx context.Context, inviteCode, userID string) (*types.Group, error)
	ListMembers(ctx context.Context, groupID, userID string) ([]*types.User, error)
}

type groupService struct {
	repo repository.GroupRepository
}

func NewGroupService(repo repository.GroupRepository) GroupService {
	return &groupService{repo: repo}
}

func (service *groupService) CreateGroup(ctx context.Context, name, userID string) (*types.Group, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, ErrUnauthorized
	}

	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrInvalidInput
	}

	inviteCode, err := generateInviteCode()
	if err != nil {
		return nil, err
	}

	group := &types.Group{
		Name:       name,
		InviteCode: inviteCode,
	}

	if err := service.repo.CreateGroup(ctx, group, userID); err != nil {
		return nil, err
	}

	return group, nil
}

func (service *groupService) JoinGroup(ctx context.Context, inviteCode, userID string) (*types.Group, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, ErrUnauthorized
	}

	inviteCode = strings.TrimSpace(inviteCode)
	if inviteCode == "" {
		return nil, ErrInvalidInput
	}

	group, err := service.repo.GetGroupByInviteCode(ctx, inviteCode)
	if err != nil {
		return nil, err
	}

	if err := service.repo.JoinGroup(ctx, group.ID, userID); err != nil {
		return nil, err
	}

	return group, nil
}

func (service *groupService) ListMembers(ctx context.Context, groupID, userID string) ([]*types.User, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, ErrUnauthorized
	}

	groupID = strings.TrimSpace(groupID)
	if groupID == "" {
		return nil, ErrInvalidInput
	}

	members, err := service.repo.ListMembers(ctx, groupID)
	if err != nil {
		return nil, err
	}

	for _, member := range members {
		if member.ID == userID {
			return members, nil
		}
	}

	return nil, ErrPermissionDenied
}

func generateInviteCode() (string, error) {
	buffer := make([]byte, 4)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return hex.EncodeToString(buffer), nil
}
