package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/errs"
)

type GetWorkspaceMembers struct {
	ID     uuid.UUID
	UserID string
}

type GetWorkspaceMembersHandler struct {
	identitySvc      IdentitySvc
	authorizationSvc AuthorizationSvc
}

func NewGetWorkspaceMembersHandler(
	identitySvc IdentitySvc,
	authorizationSvc AuthorizationSvc,
) *GetWorkspaceMembersHandler {
	return &GetWorkspaceMembersHandler{
		identitySvc:      identitySvc,
		authorizationSvc: authorizationSvc,
	}
}

var ProvideGetWorkspaceMembersHandler = NewGetWorkspaceMembersHandler

func (h *GetWorkspaceMembersHandler) Handle(ctx context.Context, query *GetWorkspaceMembers) ([]*WorkspaceMember, error) {
	hasPermission, err := h.authorizationSvc.HasWorkspaceItemPermission(
		ctx,
		query.UserID,
		query.ID,
		WorkspaceItemPermissionRead,
	)
	if err != nil {
		return nil, err
	}
	if !hasPermission {
		return nil, errs.NewForbidden(
			fmt.Sprintf("user %s does not have permission to read workspace members of workspace %s", query.UserID, query.ID),
		)
	}
	authorizationWorkspaceMembers, err := h.authorizationSvc.GetWorkspaceMembers(ctx, query.UserID, query.ID)
	if err != nil {
		return nil, err
	}
	memberIDs := make([]string, len(authorizationWorkspaceMembers))
	for i, member := range authorizationWorkspaceMembers {
		memberIDs[i] = member.ID
	}
	members, err := h.identitySvc.GetUsersByIDs(ctx, memberIDs)
	if err != nil {
		return nil, err
	}
	memberIDMemberMap := make(map[string]*User, len(members))
	for _, member := range members {
		memberIDMemberMap[member.ID] = member
	}
	workspaceMembers := make([]*WorkspaceMember, len(authorizationWorkspaceMembers))
	for i, authorizationWorkspaceMember := range authorizationWorkspaceMembers {
		member, ok := memberIDMemberMap[authorizationWorkspaceMember.ID]
		if !ok {
			return nil, errs.NewInternal(fmt.Sprintf("user with ID %s not found", authorizationWorkspaceMember.ID))
		}
		workspaceMembers[i] = &WorkspaceMember{
			ID:   member.ID,
			Name: member.Name,
			Role: authorizationWorkspaceMember.Role,
		}
	}
	return workspaceMembers, nil
}
