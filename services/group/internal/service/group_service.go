package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strings"

	"github.com/adk-saugat/rally/services/group/internal/repository"
	"github.com/adk-saugat/rally/shared/types"
)

type UserDirectory interface {
	GetUsers(ctx context.Context, userIDs []string) ([]*types.User, error)
}

type GroupService interface {
	CreateGroup(ctx context.Context, name, userID string) (*types.Group, error)
	JoinGroup(ctx context.Context, inviteCode, userID string) (*types.Group, error)
	ListMembers(ctx context.Context, groupID, userID string) ([]*types.User, error)
}

type groupService struct {
	repo  repository.GroupRepository
	users UserDirectory
}

func NewGroupService(repo repository.GroupRepository, users UserDirectory) GroupService {
	return &groupService{repo: repo, users: users}
}

func (service *groupService) CreateGroup(ctx context.Context, name, userID string) (*types.Group, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, types.ErrUnauthorized
	}

	name = strings.TrimSpace(name)
	if name == "" {
		return nil, types.ErrInvalidInput
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
		return nil, types.ErrUnauthorized
	}

	inviteCode = strings.TrimSpace(inviteCode)
	if inviteCode == "" {
		return nil, types.ErrInvalidInput
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
		return nil, types.ErrUnauthorized
	}

	groupID = strings.TrimSpace(groupID)
	if groupID == "" {
		return nil, types.ErrInvalidInput
	}

	members, err := service.repo.ListMembers(ctx, groupID)
	if err != nil {
		return nil, err
	}

	isMember := false
	userIDs := make([]string, 0, len(members))
	for _, member := range members {
		userIDs = append(userIDs, member.ID)
		if member.ID == userID {
			isMember = true
		}
	}
	if !isMember {
		return nil, types.ErrPermissionDenied
	}

	profiles, err := service.users.GetUsers(ctx, userIDs)
	if err != nil {
		return nil, err
	}

	profilesByID := make(map[string]*types.User, len(profiles))
	for _, profile := range profiles {
		profilesByID[profile.ID] = profile
	}

	for _, member := range members {
		profile, ok := profilesByID[member.ID]
		if !ok {
			continue
		}
		member.Email = profile.Email
		member.FirstName = profile.FirstName
	}

	return members, nil
}

func generateInviteCode() (string, error) {
	buffer := make([]byte, 4)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return hex.EncodeToString(buffer), nil
}
