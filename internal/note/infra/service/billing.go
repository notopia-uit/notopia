package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/timeout"
	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/internal/note/config"
	"github.com/notopia-uit/notopia/internal/note/errs"
	"github.com/notopia-uit/notopia/pkg/otel"
	"github.com/notopia-uit/notopia/pkg/pb"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

const billingBatchChunkSize = 500

func billingUnaryClientErrorInterceptor() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply any,
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		err := invoker(ctx, method, req, reply, cc, opts...)
		if err == nil {
			return nil
		}
		if code := status.Code(err); code == codes.Unavailable || code == codes.DeadlineExceeded {
			return errs.NewBillingServiceUnavailable(err)
		}
		return errs.NewInternalErr("billing service call failed", err)
	}
}

type Billing struct {
	client pb.BillingServiceClient
	conn   *grpc.ClientConn
}

var _ app.BillingSvc = (*Billing)(nil)

func NewBilling(
	servicesCfg *config.Services,
	logger logging.Logger,
	_ otel.Global,
) (*Billing, func(), error) {
	statsHandler := otelgrpc.NewClientHandler()
	conn, err := grpc.NewClient(
		servicesCfg.Billing.URL,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(statsHandler),
		grpc.WithChainUnaryInterceptor(logging.UnaryClientInterceptor(logger, logging.WithLogOnEvents(
			logging.StartCall,
			logging.FinishCall,
			logging.PayloadSent,
			logging.PayloadReceived,
		)), billingUnaryClientErrorInterceptor()),
		grpc.WithUnaryInterceptor(timeout.UnaryClientInterceptor(servicesCfg.Billing.ConnectionTimeout)),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid billing service client configuration: %w", err)
	}

	cleanup := func() {
		if err := conn.Close(); err != nil {
			slog.Error("failed to close billing service connection", slog.Any("error", err))
		}
	}
	return &Billing{
		client: pb.NewBillingServiceClient(conn),
		conn:   conn,
	}, cleanup, nil
}

var ProvideBilling = NewBilling

func toAppPlanTier(tier pb.PlanTier) (app.PlanTier, error) {
	switch tier {
	case pb.PlanTier_PLAN_TIER_FREE:
		return app.PlanTierFree, nil
	case pb.PlanTier_PLAN_TIER_PLUS:
		return app.PlanTierPlus, nil
	case pb.PlanTier_PLAN_TIER_PRO:
		return app.PlanTierPro, nil
	case pb.PlanTier_PLAN_TIER_UNSPECIFIED:
		return app.PlanTierUnspecified, errs.NewInternal("unspecified plan tier from billing")
	default:
		return app.PlanTierUnspecified, errs.NewInternal(fmt.Sprintf("invalid plan tier from billing: %v", tier))
	}
}

func toAppFeatures(in map[string]bool) map[app.PlanFeature]bool {
	out := make(map[app.PlanFeature]bool, len(in))
	for k, v := range in {
		if f, ok := app.ParsePlanFeature(k); ok {
			out[f] = v
		}
	}
	return out
}

func toAppLimits(in map[string]int64) map[app.PlanLimit]app.PlanQuota {
	out := make(map[app.PlanLimit]app.PlanQuota, len(in))
	for k, v := range in {
		if l, ok := app.ParsePlanLimit(k); ok {
			out[l] = app.PlanQuota(v)
		}
	}
	return out
}

func toEntitlements(e *pb.Entitlements) (app.Entitlements, error) {
	plan, err := toAppPlanTier(e.GetPlan())
	if err != nil {
		return app.Entitlements{}, err
	}
	return app.Entitlements{
		Plan:     plan,
		Features: toAppFeatures(e.GetFeatures()),
		Limits:   toAppLimits(e.GetLimits()),
	}, nil
}

func (b *Billing) GetEntitlements(ctx context.Context, workspaceID uuid.UUID) (app.Entitlements, error) {
	res, err := b.client.GetWorkspaceEntitlements(ctx, &pb.GetWorkspaceEntitlementsRequest{
		WorkspaceId: workspaceID.String(),
	})
	if err != nil {
		return app.Entitlements{}, err
	}
	return toEntitlements(res.GetEntitlements())
}

func (b *Billing) BatchGetEntitlements(ctx context.Context, workspaceIDs []uuid.UUID) (map[uuid.UUID]app.Entitlements, error) {
	out := make(map[uuid.UUID]app.Entitlements, len(workspaceIDs))
	for start := 0; start < len(workspaceIDs); start += billingBatchChunkSize {
		end := min(start+billingBatchChunkSize, len(workspaceIDs))
		ids := make([]string, 0, end-start)
		for _, id := range workspaceIDs[start:end] {
			ids = append(ids, id.String())
		}
		res, err := b.client.BatchGetWorkspaceEntitlements(ctx, &pb.BatchGetWorkspaceEntitlementsRequest{
			WorkspaceIds: ids,
		})
		if err != nil {
			return nil, err
		}
		for key, value := range res.GetEntitlements() {
			id, err := uuid.Parse(key)
			if err != nil {
				return nil, errs.NewInternalErr("invalid workspace ID from billing", err)
			}
			ent, err := toEntitlements(value)
			if err != nil {
				return nil, err
			}
			out[id] = ent
		}
	}
	return out, nil
}

func (b *Billing) GetPayer(ctx context.Context, workspaceID uuid.UUID) (*string, bool, error) {
	res, err := b.client.GetWorkspacePayer(ctx, &pb.GetWorkspacePayerRequest{
		WorkspaceId: workspaceID.String(),
	})
	if err != nil {
		return nil, false, err
	}
	return res.PayerId, res.GetRequiresPayer(), nil
}

func (b *Billing) ListPlans(ctx context.Context) ([]app.PlanEntitlements, error) {
	res, err := b.client.ListPlans(ctx, &pb.ListPlansRequest{})
	if err != nil {
		return nil, err
	}
	plans := make([]app.PlanEntitlements, 0, len(res.GetPlans()))
	for _, p := range res.GetPlans() {
		tier, err := toAppPlanTier(p.GetTier())
		if err != nil {
			return nil, err
		}
		plans = append(plans, app.PlanEntitlements{
			Tier:     tier,
			Features: toAppFeatures(p.GetFeatures()),
			Limits:   toAppLimits(p.GetLimits()),
		})
	}
	return plans, nil
}
