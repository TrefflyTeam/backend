package common

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"treffly/token"
)

const (
	defaultLat = "51.660781"
	defaultLon = "39.200296"
)

func GetUserIDFromSoftAuth(ctx *gin.Context) int32 {
	userIDStr, exists := ctx.Get("user_id")
	if !exists {
		userIDStr = -1
	}

	userID, ok := userIDStr.(int32)
	if !ok {
		userID = -1
	}

	return userID
}

func GetUserIDFromContextPayload(ctx *gin.Context) int32 {
	authPayload := ctx.MustGet(AuthorizationPayloadKey).(*token.Payload)
	userID := authPayload.UserID
	return userID
}

func GetUserLocation(ctx *gin.Context) (lat pgtype.Numeric, lon pgtype.Numeric, err error) {
	latStr := ctx.Query("user_lat")
	if latStr == "" {
		latStr = defaultLat
	}

	lonStr := ctx.Query("user_lon")
	if lonStr == ""{
		lonStr = defaultLon
	}

	if err := lat.Scan(latStr); err != nil {
		return pgtype.Numeric{}, pgtype.Numeric{}, fmt.Errorf("invalid latitude: %v", err)
	}

	if err := lon.Scan(lonStr); err != nil {
		return pgtype.Numeric{}, pgtype.Numeric{}, fmt.Errorf("invalid longitude: %v", err)
	}

	return lat, lon, nil
}
