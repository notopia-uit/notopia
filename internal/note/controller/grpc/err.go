package grpc

import (
	"context"
	"errors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/notopia-uit/notopia/internal/note/errs"
)

func toGRPCError(err error) error {
	if cerr, ok := errors.AsType[*errs.Err](err); ok {
		switch cerr.Code() {
		case errs.CodeUnauthorized:
			return status.Error(codes.Unauthenticated, cerr.Error())
		case errs.CodeForbidden:
			return status.Error(codes.PermissionDenied, cerr.Error())
		case errs.CodeInvalid,
			errs.CodeEmptyFolderName,
			errs.CodePersistenceInvalid,
			errs.CodeInvalidWorkspaceName,
			errs.CodeInvalidWorkspaceSlug,
			errs.CodeFoldersNotInWorkspace,
			errs.CodeDestinationFolderNotInWorkspace,
			errs.CodeNotesNotInWorkspace,
			errs.CodeWorkspaceMembersCannotBeEmpty,
			errs.CodeWorkspaceMustHaveAtLeastOneOwner,
			errs.CodeIdentityUserIDInvalid,
			errs.CodeCannotMoveFolderToItOwnSubfolder:
			return status.Error(codes.InvalidArgument, cerr.Error())
		case errs.CodeUnimplemented:
			return status.Error(codes.Unimplemented, cerr.Error())
		case errs.CodeInternal,
			errs.CodeNoteFailToMarshalDocumentContent,
			errs.CodePersistenceInternal,
			errs.CodeWorkspaceEventPubSubFailedToCreateMessage,
			errs.CodeWorkspaceEventPubSubPublishFailed,
			errs.CodeWorkspaceEventPubSubSubscribeFailed,
			errs.CodeInternalGenerateID:
			return status.Error(codes.Internal, cerr.Error())
		case errs.CodeFolderNotFound,
			errs.CodeNoteNotFound,
			errs.CodeWorkspaceNotFound,
			errs.CodeWorkspaceBySlugNotFound,
			errs.CodeWorkspaceRootFolderNotFound:
			return status.Error(codes.NotFound, cerr.Error())
		case errs.CodeFolderAlreadyExisted,
			errs.CodeFolderAlreadyTrashed,
			errs.CodeNoteAlreadyTrashed,
			errs.CodeWorkspaceSlugAlreadyExists:
			return status.Error(codes.AlreadyExists, cerr.Error())
		case errs.CodeAuthorizationServiceInternalError:
			return status.Error(codes.Internal, cerr.Error())
		default:
			return status.Error(codes.Unknown, cerr.Error())
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
