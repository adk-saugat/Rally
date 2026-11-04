package handler

import (
	"net/http"

	"github.com/adk-saugat/rally/gateway/internal/clients"
	"github.com/adk-saugat/rally/gateway/internal/utils/bearer"
	"github.com/adk-saugat/rally/gateway/internal/utils/httperr"
	"github.com/gin-gonic/gin"
)

type GroupHandler struct {
	groups *clients.GroupClient
}

func NewGroupHandler(groups *clients.GroupClient) *GroupHandler {
	return &GroupHandler{groups: groups}
}

type createGroupRequest struct {
	Name string `json:"name"`
}

type joinGroupRequest struct {
	InviteCode string `json:"invite_code"`
}

type memberResponse struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
}

func (handler *GroupHandler) CreateGroup(ginContext *gin.Context) {
	ctx, ok := bearer.Context(ginContext)
	if !ok {
		return
	}

	var request createGroupRequest
	if err := ginContext.ShouldBindJSON(&request); err != nil {
		ginContext.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	groupID, inviteCode, err := handler.groups.CreateGroup(ctx, request.Name)
	if err != nil {
		httperr.Write(ginContext, err)
		return
	}

	ginContext.JSON(http.StatusOK, gin.H{
		"group_id":    groupID,
		"invite_code": inviteCode,
	})
}

func (handler *GroupHandler) JoinGroup(ginContext *gin.Context) {
	ctx, ok := bearer.Context(ginContext)
	if !ok {
		return
	}

	var request joinGroupRequest
	if err := ginContext.ShouldBindJSON(&request); err != nil {
		ginContext.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	groupID, name, err := handler.groups.JoinGroup(ctx, request.InviteCode)
	if err != nil {
		httperr.Write(ginContext, err)
		return
	}

	ginContext.JSON(http.StatusOK, gin.H{
		"group_id": groupID,
		"name":     name,
	})
}

func (handler *GroupHandler) ListMembers(ginContext *gin.Context) {
	ctx, ok := bearer.Context(ginContext)
	if !ok {
		return
	}

	groupID := ginContext.Param("group_id")
	members, err := handler.groups.ListMembers(ctx, groupID)
	if err != nil {
		httperr.Write(ginContext, err)
		return
	}

	response := make([]memberResponse, 0, len(members))
	for _, member := range members {
		response = append(response, memberResponse{
			UserID:    member.UserID,
			Email:     member.Email,
			FirstName: member.FirstName,
		})
	}

	ginContext.JSON(http.StatusOK, gin.H{"members": response})
}
