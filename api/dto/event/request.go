package eventdto

import (
	"github.com/jackc/pgx/v5/pgtype"
	"time"
)

type CreateEventRequest struct {
	Name        string         `form:"name" binding:"required,event_name,min=5,max=50"`
	Description string         `form:"description" binding:"required,min=50,max=1000"`
	Capacity    int32          `form:"capacity" binding:"required,min=1,max=500"`
	Latitude    pgtype.Numeric `form:"latitude" binding:"required,latitude"`
	Longitude   pgtype.Numeric `form:"longitude" binding:"required,longitude"`
	Address     string         `form:"address" binding:"required"`
	Date        time.Time      `form:"date" binding:"required,valid_date"`
	IsPrivate   bool           `form:"is_private" binding:"boolean"`
	Tags        []int32        `form:"tags" binding:"required,min=1,max=3,dive,required,positive"`
}

type UpdateEventRequest struct {
	Name        string         `form:"name" binding:"required,event_name,min=5,max=50"`
	Description string         `form:"description" binding:"required,min=50,max=1000"`
	Capacity    int32          `form:"capacity" binding:"required,min=1,max=500"`
	Latitude    pgtype.Numeric `form:"latitude" binding:"required,latitude"`
	Longitude   pgtype.Numeric `form:"longitude" binding:"required,longitude"`
	Address     string         `form:"address" binding:"required"`
	Date        time.Time      `form:"date" binding:"required,valid_date"`
	IsPrivate   bool           `form:"is_private" binding:"boolean"`
	Tags        []int32        `form:"tags" binding:"required,min=1,max=3,dive,required,positive"`
	DeleteImage bool           `form:"delete_image" binding:"boolean"`
}

type CreatePremiumOrderRequest struct {
	EventID int32 `json:"event_id" binding:"required,min=1"`
}
