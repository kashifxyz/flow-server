package models

import (
	"encoding/json"
	"time"
)

type Workspace struct {
	ID                   string          `json:"id"`
	Name                 string          `json:"name"`
	Slug                 *string         `json:"slug,omitempty"`
	Description          string          `json:"description"`
	Icon                 *string         `json:"icon,omitempty"`
	Color                *string         `json:"color,omitempty"`
	Status               string          `json:"status"`
	Settings             json.RawMessage `json:"settings"`
	OwnerUserID          *string         `json:"owner_user_id,omitempty"`
	PublicSharingEnabled bool            `json:"public_sharing_enabled"`
	GuestAccessEnabled   bool            `json:"guest_access_enabled"`
	CreatedAt            time.Time       `json:"created_at"`
	UpdatedAt            time.Time       `json:"updated_at"`
}

type CreateWorkspaceRequest struct {
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	Color       string `json:"color"`
}

type UpdateWorkspaceRequest struct {
	Name                 *string `json:"name"`
	Description          *string `json:"description"`
	Icon                 *string `json:"icon"`
	Color                *string `json:"color"`
	PublicSharingEnabled *bool   `json:"public_sharing_enabled"`
	GuestAccessEnabled   *bool   `json:"guest_access_enabled"`
}

type Usage struct {
	StorageUsedBytes  int64  `json:"storage_used_bytes"`
	StorageQuotaBytes *int64 `json:"storage_quota_bytes,omitempty"`
	MemberCount       int    `json:"member_count"`
	MemberQuota       *int   `json:"member_quota,omitempty"`
	GuestCount        int    `json:"guest_count"`
	GuestQuota        *int   `json:"guest_quota,omitempty"`
}

type Member struct {
	ID          string     `json:"id"`
	UserID      string     `json:"user_id"`
	Email       string     `json:"email"`
	DisplayName string     `json:"display_name"`
	Role        string     `json:"role"`
	RoleID      *string    `json:"role_id,omitempty"`
	SeatType    string     `json:"seat_type"`
	Status      string     `json:"status"`
	JoinedAt    *time.Time `json:"joined_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

type AddMemberRequest struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

type UpdateMemberRequest struct {
	Role     *string `json:"role"`
	SeatType *string `json:"seat_type"`
	Status   *string `json:"status"`
}

type Invite struct {
	ID         string     `json:"id"`
	Email      string     `json:"email"`
	Role       string     `json:"role"`
	GroupID    *string    `json:"group_id,omitempty"`
	SeatType   string     `json:"seat_type"`
	ExpiresAt  time.Time  `json:"expires_at"`
	AcceptedAt *time.Time `json:"accepted_at,omitempty"`
	RevokedAt  *time.Time `json:"revoked_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

type CreateInviteRequest struct {
	Email    string `json:"email"`
	Role     string `json:"role"`
	GroupID  string `json:"group_id"`
	SeatType string `json:"seat_type"`
}

type Role struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	IsSystem    bool            `json:"is_system"`
	Permissions json.RawMessage `json:"permissions"`
	CreatedAt   time.Time       `json:"created_at"`
}

type CreateRoleRequest struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Permissions json.RawMessage `json:"permissions"`
}

type UpdateRoleRequest struct {
	Name        *string         `json:"name"`
	Description *string         `json:"description"`
	Permissions json.RawMessage `json:"permissions"`
}

type Group struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type CreateGroupRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type UpdateGroupRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

type GroupMember struct {
	UserID      string    `json:"user_id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name"`
	CreatedAt   time.Time `json:"created_at"`
}

type AddGroupMemberRequest struct {
	UserID string `json:"user_id"`
}

type Domain struct {
	ID         string     `json:"id"`
	Domain     string     `json:"domain"`
	VerifiedAt *time.Time `json:"verified_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

type CreateDomainRequest struct {
	Domain string `json:"domain"`
}

type OKResponse struct {
	OK bool `json:"ok"`
}
