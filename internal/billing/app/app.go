package app

import (
	"log/slog"

	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
	"go.opentelemetry.io/otel/trace"
)

const tracerName = "billing-app"

type Cmds struct {
	ArchivePrice               ArchivePriceCmd
	CancelNow                  CancelNowCmd
	CancelProviderSubscription CancelProviderSubscriptionCmd
	CompleteCheckout           CompleteCheckoutCmd
	CreatePrice                CreatePriceCmd
	EnsureDefaults             EnsureDefaultsCmd
	EnsurePortalConfiguration  EnsurePortalConfigurationCmd
	HandlePayerRemoved         HandlePayerRemovedCmd
	HandleWorkspaceDeleted     HandleWorkspaceDeletedCmd
	PruneProviderEvents        PruneProviderEventsCmd
	ReconcileSubscriptions     ReconcileSubscriptionsCmd
	SubmitRefund               SubmitRefundCmd
	SyncRefund                 SyncRefundCmd
	SyncSubscription           SyncSubscriptionCmd
	UpdatePlan                 UpdatePlanCmd
	UpdatePolicy               UpdatePolicyCmd
}

func NewCmds(
	traceProvider trace.TracerProvider,
	logger *slog.Logger,
	archivePriceHandler *ArchivePriceHandler,
	cancelNowHandler *CancelNowHandler,
	cancelProviderSubscriptionHandler *CancelProviderSubscriptionHandler,
	completeCheckoutHandler *CompleteCheckoutHandler,
	createPriceHandler *CreatePriceHandler,
	ensureDefaultsHandler *EnsureDefaultsHandler,
	ensurePortalConfigurationHandler *EnsurePortalConfigurationHandler,
	handlePayerRemovedHandler *HandlePayerRemovedHandler,
	handleWorkspaceDeletedHandler *HandleWorkspaceDeletedHandler,
	pruneProviderEventsHandler *PruneProviderEventsHandler,
	reconcileSubscriptionsHandler *ReconcileSubscriptionsHandler,
	submitRefundHandler *SubmitRefundHandler,
	syncRefundHandler *SyncRefundHandler,
	syncSubscriptionHandler *SyncSubscriptionHandler,
	updatePlanHandler *UpdatePlanHandler,
	updatePolicyHandler *UpdatePolicyHandler,
) *Cmds {
	d := commonhandler.NewDecorator(
		commonhandler.WithTracer(traceProvider.Tracer(tracerName)),
		commonhandler.WithLogger(logger),
	)
	return &Cmds{
		ArchivePrice:               d.Cmd(archivePriceHandler),
		CancelNow:                  d.Cmd(cancelNowHandler),
		CancelProviderSubscription: d.Cmd(cancelProviderSubscriptionHandler),
		CompleteCheckout:           d.Cmd(completeCheckoutHandler),
		CreatePrice:                d.Cmd(createPriceHandler),
		EnsureDefaults:             d.Cmd(ensureDefaultsHandler),
		EnsurePortalConfiguration:  d.Cmd(ensurePortalConfigurationHandler),
		HandlePayerRemoved:         d.Cmd(handlePayerRemovedHandler),
		HandleWorkspaceDeleted:     d.Cmd(handleWorkspaceDeletedHandler),
		PruneProviderEvents:        d.Cmd(pruneProviderEventsHandler),
		ReconcileSubscriptions:     d.Cmd(reconcileSubscriptionsHandler),
		SubmitRefund:               d.Cmd(submitRefundHandler),
		SyncRefund:                 d.Cmd(syncRefundHandler),
		SyncSubscription:           d.Cmd(syncSubscriptionHandler),
		UpdatePlan:                 d.Cmd(updatePlanHandler),
		UpdatePolicy:               d.Cmd(updatePolicyHandler),
	}
}

var ProvideCmds = NewCmds

type Queries struct {
	BatchGetEntitlements BatchGetEntitlementsQuery
	ClaimPayer           ClaimPayerQuery
	GetEntitlements      GetEntitlementsQuery
	GetPlan              GetPlanQuery
	GetPolicy            GetPolicyQuery
	GetPrice             GetPriceQuery
	GetWorkspaceBilling  GetWorkspaceBillingQuery
	GetWorkspacePayer    GetWorkspacePayerQuery
	ListPlans            ListPlansQuery
	ListPrices           ListPricesQuery
	PreviewCancelNow     PreviewCancelNowQuery
	StartCheckout        StartCheckoutQuery
	StartPortal          StartPortalQuery
}

func NewQueries(
	traceProvider trace.TracerProvider,
	logger *slog.Logger,
	batchGetEntitlementsHandler *BatchGetEntitlementsHandler,
	claimPayerHandler *ClaimPayerHandler,
	getEntitlementsHandler *GetEntitlementsHandler,
	getPlanHandler *GetPlanHandler,
	getPolicyHandler *GetPolicyHandler,
	getPriceHandler *GetPriceHandler,
	getWorkspaceBillingHandler *GetWorkspaceBillingHandler,
	getWorkspacePayerHandler *GetWorkspacePayerHandler,
	listPlansHandler *ListPlansHandler,
	listPricesHandler *ListPricesHandler,
	previewCancelNowHandler *PreviewCancelNowHandler,
	startCheckoutHandler *StartCheckoutHandler,
	startPortalHandler *StartPortalHandler,
) *Queries {
	d := commonhandler.NewDecorator(
		commonhandler.WithTracer(traceProvider.Tracer(tracerName)),
		commonhandler.WithLogger(logger),
	)
	return &Queries{
		BatchGetEntitlements: d.Query(batchGetEntitlementsHandler),
		ClaimPayer:           d.Query(claimPayerHandler),
		GetEntitlements:      d.Query(getEntitlementsHandler),
		GetPlan:              d.Query(getPlanHandler),
		GetPolicy:            d.Query(getPolicyHandler),
		GetPrice:             d.Query(getPriceHandler),
		GetWorkspaceBilling:  d.Query(getWorkspaceBillingHandler),
		GetWorkspacePayer:    d.Query(getWorkspacePayerHandler),
		ListPlans:            d.Query(listPlansHandler),
		ListPrices:           d.Query(listPricesHandler),
		PreviewCancelNow:     d.Query(previewCancelNowHandler),
		StartCheckout:        d.Query(startCheckoutHandler),
		StartPortal:          d.Query(startPortalHandler),
	}
}

var ProvideQueries = NewQueries

type Server struct {
	Cmds    *Cmds
	Queries *Queries
}
