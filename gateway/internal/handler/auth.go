package handler

import (
	"net/http"

	"github.com/adk-saugat/rally/gateway/internal/clients"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	auth *clients.AuthClient
}

func NewAuthHandler(auth *clients.AuthClient) *AuthHandler {
	return &AuthHandler{auth: auth}
}

type signUpRequest struct {
	FirstName string `json:"first_name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (handler *AuthHandler) SignUp(context *gin.Context) {
	var request signUpRequest
	if err := context.ShouldBindJSON(&request); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	userID, err := handler.auth.SignUp(context.Request.Context(), request.FirstName, request.Email, request.Password)
	if err != nil {
		writeError(context, err)
		return
	}

	context.JSON(http.StatusOK, gin.H{"user_id": userID})
}

func (handler *AuthHandler) Login(context *gin.Context) {
	var request loginRequest
	if err := context.ShouldBindJSON(&request); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	token, err := handler.auth.Login(context.Request.Context(), request.Email, request.Password)
	if err != nil {
		writeError(context, err)
		return
	}

	context.JSON(http.StatusOK, gin.H{"token": token})
}
