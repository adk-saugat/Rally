package handler

import (
	"context"

	groupv1 "github.com/adk-saugat/rally/gen/group/v1"
	"github.com/adk-saugat/rally/services/group/internal/service"
	"github.com/adk-saugat/rally/shared/types"
	"github.com/adk-saugat/rally/shared/utils"
)

type GroupHandler struct {
	groupv1.UnimplementedGroupServiceServer
	groups service.GroupService
}

func NewGroupHandler(groups service.GroupService) *GroupHandler {
	return &GroupHandler{groups: groups}
}

func (handler *GroupHandler) CreateGroup(ctx context.Context, request *groupv1.CreateGroupRequest) (*groupv1.CreateGroupResponse, error) {
	if request == nil {
		return nil, types.GRPCError(types.ErrInvalidInput)
	}

	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	group, err := handler.groups.CreateGroup(ctx, request.GetName(), userID)
	if err != nil {
		return nil, types.GRPCError(err)
	}

	return &groupv1.CreateGroupResponse{
		GroupId:    group.ID,
		InviteCode: group.InviteCode,
	}, nil
}

func (handler *GroupHandler) JoinGroup(ctx context.Context, request *groupv1.JoinGroupRequest) (*groupv1.JoinGroupResponse, error) {
	if request == nil {
		return nil, types.GRPCError(types.ErrInvalidInput)
	}

	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	group, err := handler.groups.JoinGroup(ctx, request.GetInviteCode(), userID)
	if err != nil {
		return nil, types.GRPCError(err)
	}

	return &groupv1.JoinGroupResponse{
		GroupId: group.ID,
		Name:    group.Name,
	}, nil
}

func (handler *GroupHandler) ListMembers(ctx context.Context, request *groupv1.ListMembersRequest) (*groupv1.ListMembersResponse, error) {
	if request == nil {
		return nil, types.GRPCError(types.ErrInvalidInput)
	}

	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	members, err := handler.groups.ListMembers(ctx, request.GetGroupId(), userID)
	if err != nil {
		return nil, types.GRPCError(err)
	}

	responseMembers := make([]*groupv1.User, 0, len(members))
	for _, member := range members {
		responseMembers = append(responseMembers, &groupv1.User{
			UserId:    member.ID,
			Email:     member.Email,
			FirstName: member.FirstName,
		})
	}

	return &groupv1.ListMembersResponse{Members: responseMembers}, nil
}

func (handler *GroupHandler) IsMember(ctx context.Context, request *groupv1.IsMemberRequest) (*groupv1.IsMemberResponse, error) {
	if request == nil {
		return nil, types.GRPCError(types.ErrInvalidInput)
	}

	userID, err := userIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	isMember, err := handler.groups.IsMember(ctx, request.GetGroupId(), userID)
	if err != nil {
		return nil, types.GRPCError(err)
	}

	return &groupv1.IsMemberResponse{IsMember: isMember}, nil
}

func userIDFromContext(ctx context.Context) (string, error) {
	userID := utils.UserIDFromContext(ctx)
	if userID == "" {
		return "", types.GRPCError(types.ErrUnauthorized)
	}
	return userID, nil
}
