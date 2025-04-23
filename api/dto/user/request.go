package userdto

type CreateUserRequest struct {
	Username string `json:"username" binding:"required,username,min=2,max=20"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type UpdateUserRequest struct {
	Username string `json:"username" binding:"required,username,min=2,max=20"`
}

type UpdateCurrentUserTagsRequest struct {
	TagIDs []int32 `json:"tag_ids" binding:"required,dive,gt=0"`
}