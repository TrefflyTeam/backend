package event

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
	"treffly/api/models"
	"treffly/apperror"
)

type descriptionGenerator interface {
	CreateChatCompletion(name, desc string, maxCharacters int) ([]byte, error)
}

type GeneratorHandler struct {
	generator descriptionGenerator
}

func NewGenerator(generator descriptionGenerator) *GeneratorHandler {
	return &GeneratorHandler{
		generator: generator,
	}
}

type GenerateDescriptionRequest struct {
	Name          string `json:"name" binding:"required,event_name,min=5,max=50"`
	Description   string `json:"description"`
	MaxCharacters int    `json:"max_characters" biding:"required,min=100,max=1000"`
}

type GenerateDescriptionResponse struct {
	Description string    `json:"description"`
	Remaining   int       `json:"remaining"`
	ResetAt     time.Time `json:"reset_at"`
}

func (g *GeneratorHandler) CreateChatCompletion(ctx *gin.Context) {
	var req GenerateDescriptionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(apperror.BadRequest.WithCause(err))
		return
	}

	var result models.RateLimitResult
	if val, exists := ctx.Get("rate_limit"); exists {
		result = val.(models.RateLimitResult)
	}

	responseData, err := g.generator.CreateChatCompletion(req.Name, req.Description, req.MaxCharacters)
	if err != nil {
		ctx.Error(apperror.BadGateway.WithCause(err))
		return
	}

	description, err := parseGeneratedDescription(responseData)
	if err != nil {
		ctx.Error(apperror.BadGateway.WithCause(err))
		return
	}

	ctx.JSON(http.StatusOK, GenerateDescriptionResponse{
		Description: description,
		Remaining:   result.Remaining,
		ResetAt:     result.ResetAt,
	})
}

func parseGeneratedDescription(response []byte) (string, error) {
	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(response, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if result.Error != nil {
		return "", fmt.Errorf("API error: %s", result.Error.Message)
	}

	if len(result.Choices) == 0 || result.Choices[0].Message.Content == "" {
		return "", fmt.Errorf("empty content in response")
	}

	return result.Choices[0].Message.Content, nil
}
