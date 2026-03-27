package authorization

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"connectrpc.com/connect"
	"connectrpc.com/otelconnect"
	"connectrpc.com/validate"
	"github.com/notopia-uit/notopia/internal/authorization/errs"
	"github.com/notopia-uit/notopia/pkg/pb/pbconnect"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/trace"
)

func toConectRPCError(err error) error {
	if cerr, ok := errors.AsType[*errs.Err](err); ok {
		switch cerr.Code() {
		case errs.CodeCasbinInternalError,
			errs.CodeCasbinEnforcerError,
			errs.CodeGetWorkspaceMembersGetFailed,
			errs.CodeInternal:
			return connect.NewError(connect.CodeInternal, cerr)
		case errs.CodeCasbinPolicySignatureInvalid,
			errs.CodeErrInvalidUserFormat,
			errs.CodeInvalid:
			return connect.NewError(connect.CodeInvalidArgument, cerr)
		case errs.CodeMemberHasNoPermission,
			errs.CodeForbidden:
			return connect.NewError(connect.CodePermissionDenied, cerr)
		case errs.CodeCreateWorkspaceExists:
			return connect.NewError(connect.CodeAlreadyExists, cerr)
		case errs.CodeUnimplemented:
			return connect.NewError(connect.CodeUnimplemented, cerr)
		}
	}
	return err
}

func newErrorInterceptor() connect.UnaryInterceptorFunc {
	interceptor := func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			resp, err := next(ctx, req)
			if err != nil {
				return nil, toConectRPCError(err)
			}
			return resp, nil
		}
	}
	return connect.UnaryInterceptorFunc(interceptor)
}

type GRPCServer struct {
	*http.Server
}

func NewGRPCServer(
	ctx context.Context,
	handler pbconnect.AuthorizationServiceHandler,
	cfg *ServerConfig,
	traceProvider *trace.TracerProvider,
	meterProvider *metric.MeterProvider,
	logger *slog.Logger,
) (*GRPCServer, func(), error) {
	otelInterceptor, err := otelconnect.NewInterceptor(
		otelconnect.WithTracerProvider(traceProvider),
		otelconnect.WithMeterProvider(meterProvider),
		otelconnect.WithTrustRemote(),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create otel interceptor: %w", err)
	}
	errInterceptor := newErrorInterceptor()
	validateInterceptor := validate.NewInterceptor()
	Path, Handler := pbconnect.NewAuthorizationServiceHandler(
		handler,
		connect.WithInterceptors(
			otelInterceptor,
			validateInterceptor,
			errInterceptor,
		),
	)
	mux := http.NewServeMux()
	mux.Handle(Path, Handler)
	protocol := new(http.Protocols)
	protocol.SetHTTP1(true)
	protocol.SetUnencryptedHTTP2(true)
	server := &GRPCServer{
		Server: &http.Server{
			Addr:      cfg.GRPC.Address(),
			Handler:   mux,
			Protocols: protocol,
		},
	}
	cleanup := func() {
		if err := server.Shutdown(ctx); err != nil {
			logger.ErrorContext(ctx, "failed to shutdown grpc server", slog.String("error", err.Error()))
		}
	}
	return server, cleanup, nil
}

var ProvideGRPCServer = NewGRPCServer

func (s *GRPCServer) Run() error {
	if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start grpc server: %w", err)
	}
	return nil
}
