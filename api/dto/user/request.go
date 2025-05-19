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
	Username    string `form:"username" binding:"required,username,min=2,max=20"`
	DeleteImage bool   `form:"delete_image" binding:"boolean"`
}

type UpdateCurrentUserTagsRequest struct {
	TagIDs []int32 `json:"tag_ids" binding:"required,dive,gt=0"`
}

type InitiateResetRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ConfirmResetRequest struct {
	Email string `json:"email" binding:"required,email"`
	Code  string `json:"code" binding:"required,len=6"`
}

type CompleteResetRequest struct {
	NewPassword string `json:"new_password" binding:"required,min=6"`
}
