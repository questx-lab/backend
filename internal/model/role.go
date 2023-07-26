package model

type CreateRoleRequest struct {
	CommunityHandle string `json:"community_handle"`
	Name            string `json:"name"`
	Permissions     int64  `json:"permissions"`
	Color           string `json:"color"`
}

type CreateRoleResponse struct {
}

type UpdateRoleRequest struct {
	RoleID      string `json:"role_id"`
	Name        string `json:"name"`
	Permissions int64  `json:"permissions"`
	Priority    int    `json:"priority"`
	Color       string `json:"color"`
}

type UpdateRoleResponse struct {
}

type GetRolesRequest struct {
	CommunityHandle string `json:"community_handle"`
}

type GetRolesResponse struct {
	Roles []Role `json:"roles"`
}

type DeleteRoleRequest struct {
	RoleID string `json:"role_id"`
}

type DeleteRoleResponse struct {
}
