package userdto

import (
	"treffly/api/common"
	userservice "treffly/api/service/user"
)

type UserConverter struct {
	env    string
	domain string
}

func NewUserConverter(env, domain string) *UserConverter {
	return &UserConverter{
		env:    env,
		domain: domain,
	}
}

func (c *UserConverter) ToUserResponse(user *userservice.User) UserResponse {
	return UserResponse{
		Username: user.Username,
		Email:    user.Email,
		CreatedAt: user.CreatedAt,
	}
}

func (c *UserConverter) ToUserWithTagsResponse(user *userservice.UserWithTags) UserWithTagsResponse {
	return UserWithTagsResponse{
		UserResponse: c.ToUserResponse(&user.User),
		Tags: c.convertTagsToResponse(user.Tags),
		ImageURL: common.ImageURL(c.env, c.domain, user.ImagePath),
	}
}

func (c *UserConverter) convertTagsToResponse(tags []userservice.Tag) []TagResponse {
	result := make([]TagResponse, len(tags))
	for i, t := range tags {
		result[i] = TagResponse{
			ID:   t.ID,
			Name: t.Name,
		}
	}
	return result
}