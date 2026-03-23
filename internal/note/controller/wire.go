package controller

import (
	"github.com/goforj/wire"
	"github.com/notopia-uit/notopia/internal/note/controller/grpc"
	"github.com/notopia-uit/notopia/internal/note/controller/health"
	"github.com/notopia-uit/notopia/internal/note/controller/http"
	"github.com/notopia-uit/notopia/internal/note/controller/integrationevent"
)

var ProviderSet = wire.NewSet(
	integrationevent.ProviderSet,
	grpc.ProviderSet,
	http.ProviderSet,
	health.ProviderSet,
)
