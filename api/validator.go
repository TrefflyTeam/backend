package api

import (
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgtype"
	"regexp"
	"time"
)

const (
	usernameRegexString  = "^[\\p{L}-]+$"
	eventNameRegexString = "^[\\p{L}\\p{N}\\p{P}\\p{S}\\p{Zs}]+$"
)

var validUsername validator.Func = func(fl validator.FieldLevel) bool {
	username, ok := fl.Field().Interface().(string)
	if !ok {
		return false
	}

	return regexp.MustCompile(usernameRegexString).MatchString(username)
}

var validEventName validator.Func = func(fl validator.FieldLevel) bool {
	name, ok := fl.Field().Interface().(string)
	if !ok {
		return false
	}

	return regexp.MustCompile(eventNameRegexString).MatchString(name)
}

var validLatitude validator.Func = func(fl validator.FieldLevel) bool {
	latitude, ok := fl.Field().Interface().(pgtype.Numeric)
	if !ok {
		return false
	}

	latitudeFloat, err := pgtype.Numeric.Float64Value(latitude)
	if err != nil {
		return false
	}
	if !latitudeFloat.Valid {
		return false
	}

	val := latitudeFloat.Float64

	return val >= -90.0 && val <= 90.0
}

var validLongitude validator.Func = func(fl validator.FieldLevel) bool {
	longitude, ok := fl.Field().Interface().(pgtype.Numeric)
	if !ok {
		return false
	}

	longitudeFloat, err := pgtype.Numeric.Float64Value(longitude)
	if err != nil {
		return false
	}
	if !longitudeFloat.Valid {
		return false
	}

	val := longitudeFloat.Float64

	return val >= -180.0 && val <= 180.0
}

var validDate validator.Func = func(fl validator.FieldLevel) bool {
	dateStr := fl.Field().String()

	_, err := time.Parse(time.RFC3339, dateStr)
	return err != nil
}

var validPositiveInteger validator.Func = func(fl validator.FieldLevel) bool {
	return fl.Field().Int() > 0
}

