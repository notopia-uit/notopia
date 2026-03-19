package commongrpc

import (
	"context"

	"connectrpc.com/connect"
	commonerror "github.com/notopia-uit/notopia/pkg/common/error"
)

func NewErrorInterceptor() connect.UnaryInterceptorFunc {
	interceptor := func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			resp, err := next(ctx, req)
			if err != nil {
				return nil, commonerror.ToConnectRPC(err)
			}
			return resp, nil
		}
	}
	return connect.UnaryInterceptorFunc(interceptor)
}

func NewClientErrorInterceptor() connect.UnaryInterceptorFunc {
	interceptor := func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			resp, err := next(ctx, req)
			if err != nil {
				return nil, commonerror.FromConnectRPC(err)
			}
			return resp, nil
		}
	}
	return connect.UnaryInterceptorFunc(interceptor)
}
