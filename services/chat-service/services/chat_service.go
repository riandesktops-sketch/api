package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"zodiac-ai-backend/services/chat-service/models"
	"zodiac-ai-backend/services/chat-service/repositories"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	ErrSessionNotFound = errors.New("chat session not found")
	ErrAIServiceDown   = errors.New("AI service unavailable")
)

// ChatService handles chat business logic
type ChatService struct {
	sessionRepo  *repositories.ChatSessionRepository
	messageRepo  *repositories.MessageRepository
	aiServiceURL string
}

// NewChatService creates a new chat service
func NewChatService(
	sessionRepo *repositories.ChatSessionRepository,
	messageRepo *repositories.MessageRepository,
	aiServiceURL string,
) *ChatService {
	return &ChatService{
		sessionRepo:  sessionRepo,
		messageRepo:  messageRepo,
		aiServiceURL: aiServiceURL,
	}
}

// CreateSession creates a new chat session
func (s *ChatService) CreateSession(ctx context.Context, userID, title string) (*models.ChatSession, error) {
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	if title == "" {
		title = "New Chat - " + time.Now().Format("Jan 02, 2006")
	}

	return s.sessionRepo.Create(ctx, userObjID, title)
}

// SendMessage sends a message and gets AI response
func (s *ChatService) SendMessage(ctx context.Context, sessionID, userID, zodiacSign, message string) (*models.MessageResponse, error) {
	sessionObjID, err := primitive.ObjectIDFromHex(sessionID)
	if err != nil {
		return nil, err
	}

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	// Verify session exists and belongs to user
	session, err := s.sessionRepo.FindByID(ctx, sessionObjID)
	if err != nil {
		return nil, ErrSessionNotFound
	}

	if session.UserID != userObjID {
		return nil, errors.New("unauthorized access to session")
	}

	// Save user message
	userMessage := &models.Message{
		SessionID: sessionObjID,
		UserID:    userObjID,
		Sender:    models.SenderUser,
		Content:   message,
	}

	if err := s.messageRepo.Create(ctx, userMessage); err != nil {
		return nil, err
	}

	// Call AI service to get response
	aiResponse, err := s.callAIService(ctx, zodiacSign, message)
	if err != nil {
		return nil, err
	}

	// Save AI message
	aiMessage := &models.Message{
		SessionID: sessionObjID,
		UserID:    userObjID,
		Sender:    models.SenderAI,
		Content:   aiResponse,
	}

	if err := s.messageRepo.Create(ctx, aiMessage); err != nil {
		return nil, err
	}

	return &models.MessageResponse{
		UserMessage: userMessage,
		AIMessage:   aiMessage,
	}, nil
}

// GetMessages gets chat history with cursor-based pagination
func (s *ChatService) GetMessages(ctx context.Context, sessionID, userID, cursor string, limit int) ([]*models.Message, string, error) {
	sessionObjID, err := primitive.ObjectIDFromHex(sessionID)
	if err != nil {
		return nil, "", err
	}

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, "", err
	}

	// Verify session belongs to user
	session, err := s.sessionRepo.FindByID(ctx, sessionObjID)
	if err != nil {
		return nil, "", ErrSessionNotFound
	}

	if session.UserID != userObjID {
		return nil, "", errors.New("unauthorized access to session")
	}

	if limit <= 0 || limit > 50 {
		limit = 20 // Default limit
	}

	return s.messageRepo.FindBySessionID(ctx, sessionObjID, cursor, limit)
}

// GenerateInsight generates insight from chat history
func (s *ChatService) GenerateInsight(ctx context.Context, sessionID, userID string) (string, error) {
	sessionObjID, err := primitive.ObjectIDFromHex(sessionID)
	if err != nil {
		return "", err
	}

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return "", err
	}

	// Verify session belongs to user
	session, err := s.sessionRepo.FindByID(ctx, sessionObjID)
	if err != nil {
		return "", ErrSessionNotFound
	}

	if session.UserID != userObjID {
		return "", errors.New("unauthorized access to session")
	}

	// Get all messages in session
	messages, err := s.messageRepo.GetAllBySessionID(ctx, sessionObjID)
	if err != nil {
		return "", err
	}

	if len(messages) == 0 {
		return "", errors.New("no messages in session")
	}

	// Build chat history
	var chatHistory strings.Builder
	for _, msg := range messages {
		sender := "User"
		if msg.Sender == models.SenderAI {
			sender = "AI"
		}
		chatHistory.WriteString(fmt.Sprintf("%s: %s\n", sender, msg.Content))
	}

	// Call AI service to generate insight
	return s.callAIInsightService(ctx, chatHistory.String())
}

// GetUserSessions gets all sessions for a user
func (s *ChatService) GetUserSessions(ctx context.Context, userID string) ([]*models.ChatSession, error) {
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	return s.sessionRepo.FindByUserID(ctx, userObjID)
}

// callAIService calls AI service to get chat response
func (s *ChatService) callAIService(ctx context.Context, zodiacSign, userMessage string) (string, error) {
	reqBody := map[string]string{
		"zodiac_sign":  zodiacSign,
		"user_message": userMessage,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.aiServiceURL+"/api/v1/ai/chat", strings.NewReader(string(jsonData)))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 35 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", ErrAIServiceDown
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", ErrAIServiceDown
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result struct {
		Success bool `json:"success"`
		Data    struct {
			Response string `json:"response"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	if !result.Success {
		return "", ErrAIServiceDown
	}

	return result.Data.Response, nil
}

// callAIInsightService calls AI service to generate insight
func (s *ChatService) callAIInsightService(ctx context.Context, chatHistory string) (string, error) {
	reqBody := map[string]string{
		"chat_history": chatHistory,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.aiServiceURL+"/api/v1/ai/insight", strings.NewReader(string(jsonData)))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 35 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", ErrAIServiceDown
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", ErrAIServiceDown
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result struct {
		Success bool `json:"success"`
		Data    struct {
			Insight string `json:"insight"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	if !result.Success {
		return "", ErrAIServiceDown
	}

	return result.Data.Insight, nil
}
