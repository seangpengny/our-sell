package user

type SessionRevokeResponse struct {
	Message string `json:"message"`
}

type UpdateRoleRequest struct {
	Role string `json:"role"`
}
