package identity

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/internal/note/errs"
	commonconfig "github.com/notopia-uit/notopia/pkg/common/config"
	"github.com/notopia-uit/notopia/pkg/otel"
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
	_ otel.Global,
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
		Timeout:   cfg.ConnectionTimeout,
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

			result[i] = a.toAppUser(user)
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return result, nil
}

func (a *Authentik) SearchUsers(ctx context.Context, params *app.IdentitySvcSearchUsersParams) ([]app.User, error) {
	ctx = context.WithValue(ctx, api.ContextAccessToken, a.token)
	query := a.client.CoreApi.CoreUsersList(ctx).
		Type_([]string{string(api.USERTYPEENUM_INTERNAL), string(api.USERTYPEENUM_EXTERNAL)}).
		Search(params.Keyword).
		PageSize(int32(params.Limit))
	switch params.ActiveStatus {
	case app.IdentitySvcActiveStatusActive:
		query = query.IsActive(true)
	case app.IdentitySvcActiveStatusInactive:
		query = query.IsActive(false)
	case app.IdentitySvcActiveStatusUnspecified:
	default:
	}
	paginatedUsers, _, err := query.Execute()
	if err != nil {
		return nil, err
	}
	return a.toAppUsers(paginatedUsers.GetResults()), nil
}

func (a *Authentik) toAppPagination(pagination *api.Pagination) app.Pagination {
	currentTotal := pagination.EndIndex - pagination.StartIndex
	hasNext := pagination.Next > 0
	hasPrev := pagination.Previous > 0
	return app.Pagination{
		Page:         uint(pagination.Current),
		TotalPages:   uint(pagination.TotalPages),
		CurrentTotal: uint(currentTotal),
		Total:        uint(pagination.Count),
		HasNext:      hasNext,
		HasPrev:      hasPrev,
	}
}

func (a *Authentik) toAppUser(user *api.User) app.User {
	return app.User{
		ID:       strconv.Itoa(int(user.GetPk())),
		UserName: user.GetUsername(),
		Name:     user.GetName(),
		Email:    user.GetEmail(),
		Roles:    user.GetRoles(),
	}
}

func (a *Authentik) toAppUsers(users []api.User) []app.User {
	result := make([]app.User, len(users))
	for i, user := range users {
		result[i] = a.toAppUser(&user)
	}
	return result
}
