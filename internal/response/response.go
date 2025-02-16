package response

import "time"

type APIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type CreateAccountResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	AccountID string `json:"account_id"`
}

type ServiceResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Logo        *string `json:"logo,omitempty"`
	UsersCount  int64   `json:"users_count"`
}

type ServiceListResponse struct {
	Success  bool              `json:"success"`
	Message  string            `json:"message"`
	Services []ServiceResponse `json:"services"`
	Total    int64             `json:"total"`
}

type ServiceUserResponse struct {
	UserID     string    `json:"user_id"`
	Username   string    `json:"username"`
	Email      string    `json:"email"`
	IsVerified bool      `json:"is_verified"`
	JoinedAt   time.Time `json:"joined_at"`
}

type ServiceUsersListResponse struct {
	Success bool                  `json:"success"`
	Message string                `json:"message"`
	Users   []ServiceUserResponse `json:"users"`
	Total   int64                 `json:"total"`
}
