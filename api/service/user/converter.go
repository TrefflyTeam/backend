package userservice

import (
	"github.com/jackc/pgx/v5/pgtype"
	"treffly/api/models"
	db "treffly/db/sqlc"
)

func convertTags(dbTags []db.Tag) []models.Tag {
	tags := make([]models.Tag, len(dbTags))
	for i, t := range dbTags {
		tags[i] = models.Tag{
			ID:   t.ID,
			Name: t.Name,
		}
	}
	return tags
}

func safeString(s pgtype.Text) string {
	if s.Valid {
		return s.String
	}
	return ""
}

func ConvertUser(dbUser db.User) models.User {
	return models.User{
		ID:        dbUser.ID,
		Username:  dbUser.Username,
		Email:     dbUser.Email,
		CreatedAt: dbUser.CreatedAt,
		IsAdmin:   dbUser.IsAdmin,
	}
}

func ConvertUserWithTags(dbUser db.UserWithTagsView) models.UserWithTags {
	return models.UserWithTags{
		User: models.User{
			ID:        dbUser.ID,
			Username:  dbUser.Username,
			Email:     dbUser.Email,
			CreatedAt: dbUser.CreatedAt,
		},
		Tags:      convertTags(dbUser.Tags),
		ImagePath: safeString(dbUser.ImagePath),
	}
}
