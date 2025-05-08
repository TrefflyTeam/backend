package geo

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"treffly/api/common"
	geodto "treffly/api/dto/geo"
	"treffly/apperror"
	"treffly/util"
)

type geoService interface {
	GetSuggestions(query string, lat, lon, radius float64) ([]byte, error)
	ReverseGeocode(address string) ([]byte, error)
	Geocode(lat, lon float64) ([]byte, error)
}

type Handler struct {
	mapService geoService
}

func NewGeoHandler(service geoService) *Handler {
	return &Handler{
		mapService: service,
	}
}

func (h *Handler) Suggest(ctx *gin.Context) {
	query := ctx.Query("text")
	if query == "" {
		ctx.Error(apperror.BadRequest.WithCause(fmt.Errorf("text parameter is required")))
		return
	}

	lat, lon, err := common.GetUserLocation(ctx)

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

	rawData, err := h.mapService.GetSuggestions(query, latFloat, lonFloat, radius)
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

func (h *Handler) Geocode(ctx *gin.Context) {
	lat, err := strconv.ParseFloat(ctx.Query("lat"), 64)
	if err != nil {
		ctx.Error(apperror.BadRequest.WithCause(err))
		return
	}
	if lat > 90.0 || lat < -90.0 {
		err = fmt.Errorf("invalid lat %f", lat)
		ctx.Error(apperror.BadRequest.WithCause(err))
		return
	}

	lon, err := strconv.ParseFloat(ctx.Query("lon"), 64)
	if err != nil {
		ctx.Error(apperror.BadRequest.WithCause(err))
		return
	}
	if lon > 180.0 || lon < -180.0 {
		err = fmt.Errorf("invalid lon %f", lon)
		ctx.Error(apperror.BadRequest.WithCause(err))
		return
	}

	body, err := h.mapService.Geocode(lat, lon)
	if err != nil {
		ctx.Error(apperror.BadRequest.WithCause(err))
		return
	}

	response, err := parseGeocodeResponse(body)

	ctx.JSON(http.StatusOK, response)
}

func (h *Handler) ReverseGeocode(ctx *gin.Context) {
	address := ctx.Query("address")
	if address == "" {
		ctx.Error(apperror.BadRequest.WithCause(fmt.Errorf("address parameter is required")))
		return
	}

	rawData, err := h.mapService.ReverseGeocode(address)
	if err != nil {
		ctx.Error(apperror.BadRequest.WithCause(err))
		return
	}

	response, err := ParseReverseGeocodeResponse(rawData)
	if err != nil {
		ctx.Error(apperror.BadRequest.WithCause(err))
		return
	}

	ctx.JSON(http.StatusOK, response)
}

func parseGeocodeResponse(data []byte) (*geodto.LocationResult, error) {
	var response geodto.GeoResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, err
	}

	if len(response.Response.GeoObjectCollection.FeatureMember) == 0 {
		return nil, fmt.Errorf("no results found")
	}

	firstResult := response.Response.GeoObjectCollection.FeatureMember[0].GeoObject

	coords := strings.Split(firstResult.Point.Pos, " ")
	if len(coords) != 2 {
		return nil, fmt.Errorf("invalid coordinates format")
	}

	lon, err := strconv.ParseFloat(coords[0], 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse longitude: %v", err)
	}

	lat, err := strconv.ParseFloat(coords[1], 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse latitude: %v", err)
	}

	address := firstResult.MetaDataProperty.GeocoderMetaData.Address.Formatted
	if address == "" {
		address = firstResult.MetaDataProperty.GeocoderMetaData.Text
	}

	parts := strings.SplitN(address, ",", 2)
	if len(parts) < 2 {
		return &geodto.LocationResult{
			Address: address,
			Lat:     lat,
			Lon:     lon,
		}, nil
	}

	address = strings.TrimSpace(parts[1])
	return &geodto.LocationResult{
		Address: address,
		Lat:     lat,
		Lon:     lon,
	}, nil
}

func ParseReverseGeocodeResponse(data []byte) (*geodto.LocationResult, error) {
	return parseGeocodeResponse(data)
}

func ParseSuggestResponse(data []byte) ([]geodto.SuggestItem, error) {
	var response geodto.SuggestResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse suggest response: %w", err)
	}

	items := make([]geodto.SuggestItem, 0, len(response.Results))
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
		items = append(items, geodto.SuggestItem{
			ID:      uuid.New().String(),
			Title:   res.Title.Text,
			Address: fmt.Sprintf("%s, %s", kind, res.Address.FormattedAddress),
		})
	}

	return items, nil
}
