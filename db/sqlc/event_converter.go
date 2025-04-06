package db

import (
	"github.com/jackc/pgx/v5/pgtype"
	"time"
)

type EventRow interface {
	GetID() int32
	GetName() string
	GetDescription() string
	GetCapacity() int32
	GetLatitude() pgtype.Numeric
	GetLongitude() pgtype.Numeric
	GetAddress() string
	GetDate() time.Time
	GetIsPrivate() bool
	GetIsPremium() bool
	GetCreatedAt() time.Time
	GetOwnerUsername() string
	GetTags() []Tag
	GetParticipantsCount() int64
}

func ConvertToEventRow[T EventRow](rows []T) []EventRow {
	result := make([]EventRow, len(rows))
	for i, row := range rows {
		result[i] = row
	}
	return result
}

func (r CreateEventRow) GetID() int32                  { return r.ID }
func (r CreateEventRow) GetName() string               { return r.Name }
func (r CreateEventRow) GetDescription() string        { return r.Description }
func (r CreateEventRow) GetCapacity() int32            { return r.Capacity }
func (r CreateEventRow) GetLatitude() pgtype.Numeric   { return r.Latitude }
func (r CreateEventRow) GetLongitude() pgtype.Numeric  { return r.Longitude }
func (r CreateEventRow) GetAddress() string            { return r.Address }
func (r CreateEventRow) GetDate() time.Time            { return r.Date }
func (r CreateEventRow) GetIsPrivate() bool            { return r.IsPrivate }
func (r CreateEventRow) GetIsPremium() bool            { return r.IsPremium }
func (r CreateEventRow) GetCreatedAt() time.Time       { return r.CreatedAt }
func (r CreateEventRow) GetOwnerUsername() string      { return "" }
func (r CreateEventRow) GetTags() []Tag                { return nil }
func (r CreateEventRow) GetParticipantsCount() int64   { return 0 }

func (r GetEventRow) GetID() int32                     { return r.ID }
func (r GetEventRow) GetName() string                  { return r.Name }
func (r GetEventRow) GetDescription() string           { return r.Description }
func (r GetEventRow) GetCapacity() int32               { return r.Capacity }
func (r GetEventRow) GetLatitude() pgtype.Numeric      { return r.Latitude }
func (r GetEventRow) GetLongitude() pgtype.Numeric     { return r.Longitude }
func (r GetEventRow) GetAddress() string               { return r.Address }
func (r GetEventRow) GetDate() time.Time               { return r.Date }
func (r GetEventRow) GetIsPrivate() bool               { return r.IsPrivate }
func (r GetEventRow) GetIsPremium() bool               { return r.IsPremium }
func (r GetEventRow) GetCreatedAt() time.Time          { return r.CreatedAt }
func (r GetEventRow) GetOwnerUsername() string         { return r.OwnerUsername.String }
func (r GetEventRow) GetTags() []Tag                   { return r.Tags }
func (r GetEventRow) GetParticipantsCount() int64      { return r.ParticipantsCount }

func (r GetGuestRecommendedEventsRow) GetID() int32            { return r.ID }
func (r GetGuestRecommendedEventsRow) GetName() string         { return r.Name }
func (r GetGuestRecommendedEventsRow) GetDescription() string  { return r.Description }
func (r GetGuestRecommendedEventsRow) GetCapacity() int32      { return r.Capacity }
func (r GetGuestRecommendedEventsRow) GetLatitude() pgtype.Numeric  { return r.Latitude }
func (r GetGuestRecommendedEventsRow) GetLongitude() pgtype.Numeric { return r.Longitude }
func (r GetGuestRecommendedEventsRow) GetAddress() string      { return r.Address }
func (r GetGuestRecommendedEventsRow) GetDate() time.Time      { return r.Date }
func (r GetGuestRecommendedEventsRow) GetIsPrivate() bool      { return r.IsPrivate }
func (r GetGuestRecommendedEventsRow) GetIsPremium() bool      { return r.IsPremium }
func (r GetGuestRecommendedEventsRow) GetCreatedAt() time.Time { return r.CreatedAt }
func (r GetGuestRecommendedEventsRow) GetOwnerUsername() string { return r.OwnerUsername.String }
func (r GetGuestRecommendedEventsRow) GetTags() []Tag          { return r.Tags }
func (r GetGuestRecommendedEventsRow) GetParticipantsCount() int64 { return r.ParticipantsCount }

func (r GetLatestEventsRow) GetID() int32              { return r.ID }
func (r GetLatestEventsRow) GetName() string           { return r.Name }
func (r GetLatestEventsRow) GetDescription() string    { return r.Description }
func (r GetLatestEventsRow) GetCapacity() int32        { return r.Capacity }
func (r GetLatestEventsRow) GetLatitude() pgtype.Numeric { return r.Latitude }
func (r GetLatestEventsRow) GetLongitude() pgtype.Numeric { return r.Longitude }
func (r GetLatestEventsRow) GetAddress() string        { return r.Address }
func (r GetLatestEventsRow) GetDate() time.Time        { return r.Date }
func (r GetLatestEventsRow) GetIsPrivate() bool        { return r.IsPrivate }
func (r GetLatestEventsRow) GetIsPremium() bool        { return r.IsPremium }
func (r GetLatestEventsRow) GetCreatedAt() time.Time   { return r.CreatedAt }
func (r GetLatestEventsRow) GetOwnerUsername() string  { return r.OwnerUsername.String }
func (r GetLatestEventsRow) GetTags() []Tag            { return r.Tags }
func (r GetLatestEventsRow) GetParticipantsCount() int64 { return r.ParticipantsCount }

func (r GetPopularEventsRow) GetID() int32             { return r.ID }
func (r GetPopularEventsRow) GetName() string          { return r.Name }
func (r GetPopularEventsRow) GetDescription() string   { return r.Description }
func (r GetPopularEventsRow) GetCapacity() int32       { return r.Capacity }
func (r GetPopularEventsRow) GetLatitude() pgtype.Numeric { return r.Latitude }
func (r GetPopularEventsRow) GetLongitude() pgtype.Numeric { return r.Longitude }
func (r GetPopularEventsRow) GetAddress() string       { return r.Address }
func (r GetPopularEventsRow) GetDate() time.Time       { return r.Date }
func (r GetPopularEventsRow) GetIsPrivate() bool       { return r.IsPrivate }
func (r GetPopularEventsRow) GetIsPremium() bool       { return r.IsPremium }
func (r GetPopularEventsRow) GetCreatedAt() time.Time  { return r.CreatedAt }
func (r GetPopularEventsRow) GetOwnerUsername() string { return r.OwnerUsername.String }
func (r GetPopularEventsRow) GetTags() []Tag           { return r.Tags }
func (r GetPopularEventsRow) GetParticipantsCount() int64 { return r.ParticipantsCount }

// GetPremiumEventsRow
func (r GetPremiumEventsRow) GetID() int32             { return r.ID }
func (r GetPremiumEventsRow) GetName() string          { return r.Name }
func (r GetPremiumEventsRow) GetDescription() string   { return r.Description }
func (r GetPremiumEventsRow) GetCapacity() int32       { return r.Capacity }
func (r GetPremiumEventsRow) GetLatitude() pgtype.Numeric { return r.Latitude }
func (r GetPremiumEventsRow) GetLongitude() pgtype.Numeric { return r.Longitude }
func (r GetPremiumEventsRow) GetAddress() string       { return r.Address }
func (r GetPremiumEventsRow) GetDate() time.Time       { return r.Date }
func (r GetPremiumEventsRow) GetIsPrivate() bool       { return r.IsPrivate }
func (r GetPremiumEventsRow) GetIsPremium() bool       { return r.IsPremium }
func (r GetPremiumEventsRow) GetCreatedAt() time.Time  { return r.CreatedAt }
func (r GetPremiumEventsRow) GetOwnerUsername() string { return r.OwnerUsername.String }
func (r GetPremiumEventsRow) GetTags() []Tag           { return r.Tags }
func (r GetPremiumEventsRow) GetParticipantsCount() int64 { return r.ParticipantsCount }

func (r GetUserRecommendedEventsRow) GetID() int32     { return r.ID }
func (r GetUserRecommendedEventsRow) GetName() string  { return r.Name }
func (r GetUserRecommendedEventsRow) GetDescription() string { return r.Description }
func (r GetUserRecommendedEventsRow) GetCapacity() int32 { return r.Capacity }
func (r GetUserRecommendedEventsRow) GetLatitude() pgtype.Numeric { return r.Latitude }
func (r GetUserRecommendedEventsRow) GetLongitude() pgtype.Numeric { return r.Longitude }
func (r GetUserRecommendedEventsRow) GetAddress() string { return r.Address }
func (r GetUserRecommendedEventsRow) GetDate() time.Time { return r.Date }
func (r GetUserRecommendedEventsRow) GetIsPrivate() bool { return r.IsPrivate }
func (r GetUserRecommendedEventsRow) GetIsPremium() bool { return r.IsPremium }
func (r GetUserRecommendedEventsRow) GetCreatedAt() time.Time { return r.CreatedAt }
func (r GetUserRecommendedEventsRow) GetOwnerUsername() string { return r.OwnerUsername.String }
func (r GetUserRecommendedEventsRow) GetTags() []Tag   { return r.Tags }
func (r GetUserRecommendedEventsRow) GetParticipantsCount() int64 { return r.ParticipantsCount }

func (r ListEventsRow) GetID() int32                   { return r.ID }
func (r ListEventsRow) GetName() string                { return r.Name }
func (r ListEventsRow) GetDescription() string         { return r.Description }
func (r ListEventsRow) GetCapacity() int32             { return r.Capacity }
func (r ListEventsRow) GetLatitude() pgtype.Numeric    { return r.Latitude }
func (r ListEventsRow) GetLongitude() pgtype.Numeric   { return r.Longitude }
func (r ListEventsRow) GetAddress() string             { return r.Address }
func (r ListEventsRow) GetDate() time.Time             { return r.Date }
func (r ListEventsRow) GetIsPrivate() bool             { return r.IsPrivate }
func (r ListEventsRow) GetIsPremium() bool             { return r.IsPremium }
func (r ListEventsRow) GetCreatedAt() time.Time        { return r.CreatedAt }
func (r ListEventsRow) GetOwnerUsername() string       { return r.OwnerUsername.String }
func (r ListEventsRow) GetTags() []Tag                 { return r.Tags }
func (r ListEventsRow) GetParticipantsCount() int64    { return r.ParticipantsCount }

func (r GetPastUserEventsRow) GetID() int32                   { return r.ID }
func (r GetPastUserEventsRow) GetName() string                { return r.Name }
func (r GetPastUserEventsRow) GetDescription() string         { return r.Description }
func (r GetPastUserEventsRow) GetCapacity() int32             { return r.Capacity }
func (r GetPastUserEventsRow) GetLatitude() pgtype.Numeric    { return r.Latitude }
func (r GetPastUserEventsRow) GetLongitude() pgtype.Numeric   { return r.Longitude }
func (r GetPastUserEventsRow) GetAddress() string             { return r.Address }
func (r GetPastUserEventsRow) GetDate() time.Time             { return r.Date }
func (r GetPastUserEventsRow) GetIsPrivate() bool             { return r.IsPrivate }
func (r GetPastUserEventsRow) GetIsPremium() bool             { return r.IsPremium }
func (r GetPastUserEventsRow) GetCreatedAt() time.Time        { return r.CreatedAt }
func (r GetPastUserEventsRow) GetOwnerUsername() string       { return r.OwnerUsername.String }
func (r GetPastUserEventsRow) GetTags() []Tag                 { return r.Tags }
func (r GetPastUserEventsRow) GetParticipantsCount() int64    { return r.ParticipantsCount }

func (r GetUpcomingUserEventsRow) GetID() int32                   { return r.ID }
func (r GetUpcomingUserEventsRow) GetName() string                { return r.Name }
func (r GetUpcomingUserEventsRow) GetDescription() string         { return r.Description }
func (r GetUpcomingUserEventsRow) GetCapacity() int32             { return r.Capacity }
func (r GetUpcomingUserEventsRow) GetLatitude() pgtype.Numeric    { return r.Latitude }
func (r GetUpcomingUserEventsRow) GetLongitude() pgtype.Numeric   { return r.Longitude }
func (r GetUpcomingUserEventsRow) GetAddress() string             { return r.Address }
func (r GetUpcomingUserEventsRow) GetDate() time.Time             { return r.Date }
func (r GetUpcomingUserEventsRow) GetIsPrivate() bool             { return r.IsPrivate }
func (r GetUpcomingUserEventsRow) GetIsPremium() bool             { return r.IsPremium }
func (r GetUpcomingUserEventsRow) GetCreatedAt() time.Time        { return r.CreatedAt }
func (r GetUpcomingUserEventsRow) GetOwnerUsername() string       { return r.OwnerUsername.String }
func (r GetUpcomingUserEventsRow) GetTags() []Tag                 { return r.Tags }
func (r GetUpcomingUserEventsRow) GetParticipantsCount() int64    { return r.ParticipantsCount }

func (r GetOwnedUserEventsRow) GetID() int32                   { return r.ID }
func (r GetOwnedUserEventsRow) GetName() string                { return r.Name }
func (r GetOwnedUserEventsRow) GetDescription() string         { return r.Description }
func (r GetOwnedUserEventsRow) GetCapacity() int32             { return r.Capacity }
func (r GetOwnedUserEventsRow) GetLatitude() pgtype.Numeric    { return r.Latitude }
func (r GetOwnedUserEventsRow) GetLongitude() pgtype.Numeric   { return r.Longitude }
func (r GetOwnedUserEventsRow) GetAddress() string             { return r.Address }
func (r GetOwnedUserEventsRow) GetDate() time.Time             { return r.Date }
func (r GetOwnedUserEventsRow) GetIsPrivate() bool             { return r.IsPrivate }
func (r GetOwnedUserEventsRow) GetIsPremium() bool             { return r.IsPremium }
func (r GetOwnedUserEventsRow) GetCreatedAt() time.Time        { return r.CreatedAt }
func (r GetOwnedUserEventsRow) GetOwnerUsername() string       { return r.OwnerUsername.String }
func (r GetOwnedUserEventsRow) GetTags() []Tag                 { return r.Tags }
func (r GetOwnedUserEventsRow) GetParticipantsCount() int64    { return r.ParticipantsCount }