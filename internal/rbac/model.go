package rbac

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token        string `json:"token"`
	Username     string `json:"username"`
	IsFirstLogin bool   `json:"is_first_login"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

type CreateUserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	PersonID *uint  `json:"person_id"`
}

type UpdateUserRequest struct {
	PersonID *uint `json:"person_id"`
	Status   *int8 `json:"status"`
}

type ResetPasswordRequest struct {
	NewPassword string `json:"new_password" binding:"required"`
}

type CreateRoleRequest struct {
	Name   string `json:"name" binding:"required"`
	Remark string `json:"remark"`
}

type UpdateRoleRequest struct {
	Name   string `json:"name"`
	Remark string `json:"remark"`
}

type AssignRolesRequest struct {
	RoleIDs []uint `json:"role_ids" binding:"required"`
}

type AssignPermissionsRequest struct {
	PermissionIDs []uint `json:"permission_ids" binding:"required"`
}
