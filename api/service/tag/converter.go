package tagservice

import (
	"treffly/api/models"
	db "treffly/db/sqlc"
)

func convertTags(dbTags []db.Tag) []models.Tag {  //TODO: duplicate method
	tags := make([]models.Tag, len(dbTags))
	for i, t := range dbTags {
		tags[i] = models.Tag{
			ID:   t.ID,
			Name: t.Name,
		}
	}
	return tags
}

