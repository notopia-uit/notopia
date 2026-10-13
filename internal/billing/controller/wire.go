package controller

import (
	"github.com/goforj/wire"
	"github.com/notopia-uit/notopia/internal/billing/controller/event"
	"github.com/notopia-uit/notopia/internal/billing/controller/grpc"
	"github.com/notopia-uit/notopia/internal/billing/controller/health"
	"github.com/notopia-uit/notopia/internal/billing/controller/http"
)

var ProviderSet = wire.NewSet(
	event.ProviderSet,
	grpc.ProviderSet,
	http.ProviderSet,
	health.ProviderSet,
)
