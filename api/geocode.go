package api

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"strings"
	"treffly/apperror"
)

func (server *Server) geocode(ctx *gin.Context) {
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

	body, err := server.mapsClient.GetLocation(lat, lon)
	if err != nil {
		ctx.Error(apperror.BadRequest.WithCause(err))
		return
	}

	response, err := parseGeocodeResponse(body)

	ctx.JSON(http.StatusOK, response)
}

type GeoResponse struct {
	Response struct {
		GeoObjectCollection struct {
			FeatureMember []struct {
				GeoObject struct {
					MetaDataProperty struct {
						GeocoderMetaData struct {
							Text    string `json:"text"`
							Address struct {
								Formatted string `json:"formatted"`
							} `json:"Address"`
						} `json:"GeocoderMetaData"`
					} `json:"metaDataProperty"`
					Point struct {
						Pos string `json:"pos"`
					} `json:"Point"`
				} `json:"GeoObject"`
			} `json:"featureMember"`
		} `json:"GeoObjectCollection"`
	} `json:"response"`
}

type LocationResult struct {
	Address string
	Lat     float64
	Lon     float64
}

func parseGeocodeResponse(data []byte) (*LocationResult, error) {
	var response GeoResponse
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

	return &LocationResult{
		Address: address,
		Lat:     lat,
		Lon:     lon,
	}, nil
}