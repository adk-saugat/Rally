package clients

import (
	"context"

	groupv1 "github.com/adk-saugat/rally/gen/group/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Member struct {
	UserID    string
	Email     string
	FirstName string
}

type GroupClient struct {
	connection *grpc.ClientConn
	groups     groupv1.GroupServiceClient
}

func NewGroupClient(address string) (*GroupClient, error) {
	connection, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &GroupClient{
		connection: connection,
		groups:     groupv1.NewGroupServiceClient(connection),
	}, nil
}

func (client *GroupClient) CreateGroup(ctx context.Context, name string) (groupID, inviteCode string, err error) {
	response, err := client.groups.CreateGroup(outgoingWithAuthorization(ctx), &groupv1.CreateGroupRequest{
		Name: name,
	})
	if err != nil {
		return "", "", err
	}
	return response.GetGroupId(), response.GetInviteCode(), nil
}

func (client *GroupClient) JoinGroup(ctx context.Context, inviteCode string) (groupID, name string, err error) {
	response, err := client.groups.JoinGroup(outgoingWithAuthorization(ctx), &groupv1.JoinGroupRequest{
		InviteCode: inviteCode,
	})
	if err != nil {
		return "", "", err
	}
	return response.GetGroupId(), response.GetName(), nil
}

func (client *GroupClient) ListMembers(ctx context.Context, groupID string) ([]Member, error) {
	response, err := client.groups.ListMembers(outgoingWithAuthorization(ctx), &groupv1.ListMembersRequest{
		GroupId: groupID,
	})
	if err != nil {
		return nil, err
	}

	members := make([]Member, 0, len(response.GetMembers()))
	for _, member := range response.GetMembers() {
		members = append(members, Member{
			UserID:    member.GetUserId(),
			Email:     member.GetEmail(),
			FirstName: member.GetFirstName(),
		})
	}
	return members, nil
}

func (client *GroupClient) Close() error {
	return client.connection.Close()
}
