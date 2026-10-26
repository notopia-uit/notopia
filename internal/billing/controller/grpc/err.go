package grpc

import (
	"context"
	"errors"

	"github.com/notopia-uit/notopia/internal/billing/errs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func toGRPCError(err error) error {
	if cerr, ok := errors.AsType[errs.Error](err); ok {
		switch cerr.Code() {
		case errs.CodeUnauthorized:
			return status.Error(codes.Unauthenticated, cerr.Error())
		case errs.CodeForbidden, errs.CodeNotOwner:
			return status.Error(codes.PermissionDenied, cerr.Error())
		case errs.CodeInvalid:
			return status.Error(codes.InvalidArgument, cerr.Error())
		case errs.CodeNotFound, errs.CodeNoBillingAccount:
			return status.Error(codes.NotFound, cerr.Error())
		case errs.CodeSubscriptionAlreadyActive,
			errs.CodeNoActiveSubscription,
			errs.CodeRefundWindowExpired,
			errs.CodeAlreadyPayer,
			errs.CodePriceNotActive:
			return status.Error(codes.FailedPrecondition, cerr.Error())
		case errs.CodeWebhookSignature:
			return status.Error(codes.InvalidArgument, cerr.Error())
		case errs.CodeInternal, errs.CodePaymentProvider, errs.CodeAuthorizationService:
			return status.Error(codes.Internal, cerr.Error())
		default:
			return status.Error(codes.Internal, cerr.Error())
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
