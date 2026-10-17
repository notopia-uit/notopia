package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/timeout"
	"github.com/notopia-uit/notopia/internal/billing/app"
	"github.com/notopia-uit/notopia/internal/billing/config"
	"github.com/notopia-uit/notopia/internal/billing/errs"
	"github.com/notopia-uit/notopia/pkg/otel"
	"github.com/notopia-uit/notopia/pkg/pb"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func authorizationUnaryClientErrorInterceptor() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply any,
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		err := invoker(ctx, method, req, reply, cc, opts...)
		if err != nil {
			return errs.New(errs.CodeAuthorizationService, "authorization service error", err)
		}
		return nil
	}
}

type Authorization struct {
	client pb.AuthorizationServiceClient
	conn   *grpc.ClientConn
}

var _ app.AuthorizationSvc = (*Authorization)(nil)

func NewAuthorization(
	servicesCfg *config.Services,
	logger logging.Logger,
	_ otel.Global,
) (*Authorization, func(), error) {
	conn, err := grpc.NewClient(
		servicesCfg.Authorization.URL,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
		grpc.WithChainUnaryInterceptor(logging.UnaryClientInterceptor(logger, logging.WithLogOnEvents(
			logging.StartCall,
			logging.FinishCall,
			logging.PayloadSent,
			logging.PayloadReceived,
		)), authorizationUnaryClientErrorInterceptor()),
		grpc.WithUnaryInterceptor(timeout.UnaryClientInterceptor(servicesCfg.Authorization.ConnectionTimeout)),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid authorization service client configuration: %w", err)
	}
	cleanup := func() {
		if err := conn.Close(); err != nil {
			slog.Error("failed to close authorization service connection", slog.Any("error", err))
		}
	}
	return &Authorization{
		client: pb.NewAuthorizationServiceClient(conn),
		conn:   conn,
	}, cleanup, nil
}

var ProvideAuthorization = NewAuthorization

func (a *Authorization) HasWorkspacePermission(ctx context.Context, userID string, workspaceID uuid.UUID, p app.WorkspacePermission) (bool, error) {
	perm, err := toPBWorkspacePermission(p)
	if err != nil {
		return false, err
	}
	res, err := a.client.HasWorkspacePermission(ctx, &pb.HasWorkspacePermissionRequest{
		MemberId:    userID,
		WorkspaceId: workspaceID.String(),
		Permission:  perm,
	})
	if err != nil {
		return false, err
	}
	return res.HasPermission, nil
}

func (a *Authorization) GetMemberRole(ctx context.Context, callerID string, workspaceID uuid.UUID, memberID string) (app.WorkspaceRole, error) {
	res, err := a.client.GetWorkspaceMembers(ctx, &pb.GetWorkspaceMembersRequest{
		UserId:      callerID,
		WorkspaceId: workspaceID.String(),
	})
	if err != nil {
		return app.WorkspaceRoleNone, err
	}
	for _, m := range res.Members {
		if m.Id == memberID {
			return toAppWorkspaceRole(m.Role), nil
		}
	}
	return app.WorkspaceRoleNone, nil
}

func toPBWorkspacePermission(p app.WorkspacePermission) (pb.WorkspacePermission, error) {
	switch p {
	case app.WorkspacePermissionBillingRead:
		return pb.WorkspacePermission_WORKSPACE_PERMISSION_BILLING_READ, nil
	case app.WorkspacePermissionBillingManage:
		return pb.WorkspacePermission_WORKSPACE_PERMISSION_BILLING_MANAGE, nil
	case app.WorkspacePermissionUnspecified:
	}
	return pb.WorkspacePermission_WORKSPACE_PERMISSION_UNSPECIFIED, errs.NewInternal(fmt.Sprintf("invalid workspace permission: %v", p), nil)
}

func toAppWorkspaceRole(role pb.WorkspaceRole) app.WorkspaceRole {
	switch role {
	case pb.WorkspaceRole_WORKSPACE_ROLE_OWNER:
		return app.WorkspaceRoleOwner
	case pb.WorkspaceRole_WORKSPACE_ROLE_EDITOR:
		return app.WorkspaceRoleEditor
	case pb.WorkspaceRole_WORKSPACE_ROLE_VIEWER:
		return app.WorkspaceRoleViewer
	case pb.WorkspaceRole_WORKSPACE_ROLE_UNSPECIFIED:
	}
	return app.WorkspaceRoleNone
}
