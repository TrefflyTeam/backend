package userservice

import (
	"github.com/jackc/pgx/v5/pgtype"
	db "treffly/db/sqlc"
)

func convertTags(dbTags []db.Tag) []Tag {
	tags := make([]Tag, len(dbTags))
	for i, t := range dbTags {
		tags[i] = Tag{
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

func ConvertUser(dbUser db.User) User {
	return User{
		ID:       dbUser.ID,
		Username: dbUser.Username,
		Email:    dbUser.Email,
		CreatedAt: dbUser.CreatedAt,
	}
}

func ConvertUserWithTags(dbUser db.UserWithTagsView) UserWithTags {
	return UserWithTags{
		User: User{
			ID: dbUser.ID,
			Username: dbUser.Username,
			Email: dbUser.Email,
			CreatedAt: dbUser.CreatedAt,
		},
		Tags: convertTags(dbUser.Tags),
		ImagePath: safeString(dbUser.ImagePath),
	}
}
