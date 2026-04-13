package myprofile

type UpdateMyProfileRequest struct {
	FirstName *string `json:"first_name"`
	LastName  *string `json:"last_name"`
	Email     *string `json:"email"`
	Phone     *string `json:"phone"`
	AvatarURL *string `json:"avatar_url"`
}

type ChangeMyPasswordRequest struct {
	OldPassword *string `json:"old_password"`
	NewPassword *string `json:"new_password"`
}
