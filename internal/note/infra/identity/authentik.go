package identity

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/internal/note/errs"
	commonconfig "github.com/notopia-uit/notopia/pkg/common/config"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	semconv "go.opentelemetry.io/otel/semconv/v1.39.0"
	"go.opentelemetry.io/otel/trace"
	"goauthentik.io/api/v3"
	"golang.org/x/sync/errgroup"
)

type Authentik struct {
	client *api.APIClient
	token  string
}

func NewAuthentik(
	cfg *commonconfig.Authentik,
	logger *slog.Logger,
) *Authentik {
	authentikCfg := api.NewConfiguration()
	authentikCfg.Host = cfg.Host
	authentikCfg.Scheme = cfg.Scheme
	authentikCfg.Servers = api.ServerConfigurations{
		{
			URL: cfg.URL,
		},
	}
	otelTransport := otelhttp.NewTransport(
		http.DefaultTransport,
		otelhttp.WithSpanOptions(trace.WithAttributes(
			semconv.ServicePeerName("authentik"),
		)),
	)
	logTransport := &AuthentikLogRoundTripper{
		next:   otelTransport,
		logger: logger,
	}
	authentikCfg.HTTPClient = &http.Client{
		Transport: logTransport,
	}
	client := api.NewAPIClient(authentikCfg)
	return &Authentik{
		client: client,
		token:  cfg.Token,
	}
}

var ProvideAuthentik = NewAuthentik

var _ app.IdentitySvc = (*Authentik)(nil)

// Due to limitation of Authentik API, we have to retrieve users one by one
func (a *Authentik) GetUsersByIDs(ctx context.Context, ids []string) ([]app.User, error) {
	ctx = context.WithValue(ctx, api.ContextAccessToken, a.token)
	result := make([]app.User, len(ids))

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(10)

	for i, id := range ids {
		g.Go(func() error {
			userID, err := strconv.Atoi(id)
			if err != nil {
				return errs.NewIdentityUserIDInvalid(id, err)
			}

			user, _, err := a.client.CoreApi.CoreUsersRetrieve(ctx, int32(userID)).Execute()
			if err != nil {
				return err
			}

			result[i] = app.User{
				ID:     id,
				Name:   user.GetName(),
				Email:  user.GetEmail(),
				Groups: user.GetGroups(),
				Roles:  user.GetRoles(),
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return result, nil
}
