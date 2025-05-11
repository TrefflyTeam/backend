package event

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"treffly/apperror"
)

type descriptionGenerator interface {
	CreateChatCompletion(name, desc string) ([]byte, error)
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
	Name        string `json:"name" binding:"required,event_name,min=5,max=50"`
	Description string `json:"description"`
}

type GenerateDescriptionResponse struct {
	Description string `json:"description"`
	Error       string `json:"error,omitempty"`
}

func (g *GeneratorHandler) CreateChatCompletion(ctx *gin.Context) {
	var req GenerateDescriptionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(apperror.BadRequest.WithCause(err))
		return
	}

	responseData, err := g.generator.CreateChatCompletion(req.Name, req.Description)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, GenerateDescriptionResponse{
			Error: "Failed to generate description: " + err.Error(),
		})
		return
	}

	description, err := parseGeneratedDescription(responseData)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, GenerateDescriptionResponse{
			Error: "Failed to parse API response: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, GenerateDescriptionResponse{
		Description: description,
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
