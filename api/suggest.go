package api

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"slices"
	"strconv"
	"treffly/apperror"
	"treffly/util"
)

type SuggestResponse struct {
	Results []struct {
		Title struct {
			Text string `json:"text"`
		} `json:"title"`
		Address struct {
			FormattedAddress string `json:"formatted_address"`
			Components []struct {
				Name string   `json:"name"`
				Kind []string `json:"kind"`
			} `json:"component"`
		} `json:"address"`
	} `json:"results"`
}

type SuggestItem struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Address string `json:"address"`
}

func ParseSuggestResponse(data []byte) ([]SuggestItem, error) {
	var response SuggestResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse suggest response: %w", err)
	}

	items := make([]SuggestItem, 0, len(response.Results))
	addressSet := make(map[string]bool)
	for _, res := range response.Results {
		if addressSet[res.Address.FormattedAddress] {
			continue
		}

		addressSet[res.Address.FormattedAddress] = true
		kind := "Неизвестно"
		for _, comp := range res.Address.Components {
			if slices.Contains(comp.Kind, "LOCALITY") {
				kind = comp.Name
				break
			}
		}
		items = append(items, SuggestItem{
			ID:      uuid.New().String(),
			Title:   res.Title.Text,
			Address: fmt.Sprintf("%s, %s", kind, res.Address.FormattedAddress),
		})
	}

	return items, nil
}

func (server *Server) suggest(ctx *gin.Context) {
	query := ctx.Query("text")
	if query == "" {
		ctx.Error(apperror.BadRequest.WithCause(fmt.Errorf("text parameter is required")))
		return
	}

	lat, lon, err := getUserLocation(ctx)

	radius, err := strconv.ParseFloat(ctx.DefaultQuery("radius", "0.2"), 64)
	if err != nil {
		ctx.Error(apperror.BadRequest.WithCause(err))
		return
	}

	latFloat, err := util.NumericToFloat64(lat)
	if err != nil {
		ctx.Error(apperror.BadRequest.WithCause(err))
		return
	}

	lonFloat, err := util.NumericToFloat64(lon)
	if err != nil {
		ctx.Error(apperror.BadRequest.WithCause(err))
		return
	}

	rawData, err := server.suggestClient.GetSuggestions(query, latFloat, lonFloat, radius)
	if err != nil {
		ctx.Error(apperror.BadGateway.WithCause(err))
		return
	}

	suggestions, err := ParseSuggestResponse(rawData)
	if err != nil {
		ctx.Error(apperror.BadGateway.WithCause(err))
		return
	}

	ctx.JSON(http.StatusOK, suggestions)
}
