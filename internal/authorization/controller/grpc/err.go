package grpc

import (
	"context"
	"errors"

	"github.com/notopia-uit/notopia/internal/authorization/errs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func toGRPCError(err error) error {
	if err, ok := errors.AsType[errs.Error](err); ok {
		switch err.Code() {
		case errs.CodeCasbinInternalError,
			errs.CodeCasbinEnforcerError,
			errs.CodeGetWorkspaceMembersGetFailed,
			errs.CodePublishIntegrationEventsFailed,
			errs.CodeInternal:
			return status.Error(codes.Internal, err.Message())
		case errs.CodeCasbinPolicySignatureInvalid,
			errs.CodeErrInvalidUserFormat,
			errs.CodeInvalidWorkspaceRoleFormat,
			errs.CodeInvalid:
			return status.Error(codes.InvalidArgument, err.Message())
		case errs.CodeMemberHasNoPermission,
			errs.CodeUserIsOnlyOwner,
			errs.CodeForbidden:
			return status.Error(codes.PermissionDenied, err.Message())
		case errs.CodeCreateWorkspaceExists:
			return status.Error(codes.AlreadyExists, err.Message())
		case errs.CodeUnimplemented:
			return status.Error(codes.Unimplemented, err.Message())
		}
	}
	return err
}

func unaryErrorInterceptor(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (any, error) {
	resp, err := handler(ctx, req)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return resp, nil
}
