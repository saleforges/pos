package user

import (
	"github.com/saleforge/pos/services/internal/iam/domain"
	"github.com/saleforge/pos/services/internal/iam/transport/http/common"
)

func toUserResponse(u domain.User) common.UserResponse {
	r := common.UserResponse{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		Type:      string(u.Type),
		Status:    string(u.Status),
		CreatedAt: u.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: u.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if u.SystemRole != nil {
		r.Role = u.SystemRole.Name
	}
	return r
}
