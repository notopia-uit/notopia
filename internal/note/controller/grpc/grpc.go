package grpc

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"connectrpc.com/connect"
	"connectrpc.com/otelconnect"
	"connectrpc.com/validate"
	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/internal/note/config"
	"github.com/notopia-uit/notopia/internal/note/errs"
	"github.com/notopia-uit/notopia/pkg/pb/pbconnect"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/trace"
)

func toConnectRPCError(err error) error {
	if cerr, ok := errors.AsType[*errs.Err](err); ok {
		switch cerr.Code() {
		case errs.CodeForbidden:
			return connect.NewError(connect.CodePermissionDenied, cerr)
		case errs.CodeInvalid,
			errs.CodeEmptyFolderName,
			errs.CodePersistenceInvalid,
			errs.CodeInvalidWorkspaceName,
			errs.CodeInvalidWorkspaceSlug:
			return connect.NewError(connect.CodeInvalidArgument, cerr)
		case errs.CodeUnimplemented:
			return connect.NewError(connect.CodeUnimplemented, cerr)
		case errs.CodeInternal,
			errs.CodeNoteFailToMarshalDocumentContent,
			errs.CodePersistenceInternal,
			errs.CodeWorkspaceEventPubSubFailedToCreateMessage,
			errs.CodeWorkspaceEventPubSubPublishFailed,
			errs.CodeWorkspaceEventPubSubSubscribeFailed,
			errs.CodeInternalGenerateID:
			return connect.NewError(connect.CodeInternal, cerr)
		case errs.CodeFolderNotFound,
			errs.CodeNoteNotFound,
			errs.CodeWorkspaceNotFound,
			errs.CodeWorkspaceBySlugNotFound,
			errs.CodeWorkspaceRootFolderNotFound:
			return connect.NewError(connect.CodeNotFound, cerr)
		case errs.CodeFolderAlreadyTrashed,
			errs.CodeNoteAlreadyTrashed,
			errs.CodeWorkspaceSlugAlreadyExists:
			return connect.NewError(connect.CodeAlreadyExists, cerr)
		case errs.CodeAuthorizationServiceInternalError:
			return connect.NewError(connect.CodeInternal, cerr)
		default:
			return connect.NewError(connect.CodeUnknown, cerr)
		}
	}
	return err
}

func newErrorInterceptor() connect.UnaryInterceptorFunc {
	interceptor := func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			resp, err := next(ctx, req)
			if err != nil {
				return nil, toConnectRPCError(err)
			}
			return resp, nil
		}
	}
	return connect.UnaryInterceptorFunc(interceptor)
}

type IHandler = pbconnect.NoteServiceHandler

type Handler struct {
	app app.Server
}

var _ IHandler = (*Handler)(nil)

func NewHandler(app *app.Server) *Handler {
	return &Handler{
		app: *app,
	}
}

var ProvideHandler = NewHandler

type GRPC struct {
	*http.Server
}

func New(
	ctx context.Context,
	handler IHandler,
	cfg *config.Server,
	traceProvider *trace.TracerProvider,
	meterProvider *metric.MeterProvider,
	logger *slog.Logger,
) (*GRPC, func(), error) {
	otelInterceptor, err := otelconnect.NewInterceptor(
		otelconnect.WithTracerProvider(traceProvider),
		otelconnect.WithMeterProvider(meterProvider),
		otelconnect.WithTrustRemote(),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create otel interceptor: %w", err)
	}
	validateInterceptor := validate.NewInterceptor()
	errInterceptor := newErrorInterceptor()
	Path, Handler := pbconnect.NewNoteServiceHandler(
		handler,
		connect.WithInterceptors(
			otelInterceptor,
			validateInterceptor,
			errInterceptor,
		),
	)
	mux := http.NewServeMux()
	mux.Handle(Path, Handler)

	mux.Handle("/ping", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte("pong")); err != nil {
			logger.ErrorContext(ctx, "failed to write ping response", slog.String("error", err.Error()))
		}
	}))

	protocol := new(http.Protocols)
	protocol.SetHTTP1(true)
	protocol.SetUnencryptedHTTP2(true)
	server := &GRPC{
		Server: &http.Server{
			Addr:      cfg.GRPC.Address(),
			Handler:   mux,
			Protocols: protocol,
		},
	}

	cleanup := func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.ErrorContext(ctx, "failed to shutdown grpc server", slog.String("error", err.Error()))
		}
	}
	return server, cleanup, nil
}

func (g *GRPC) Run() error {
	return g.ListenAndServe()
}

var Provide = New
