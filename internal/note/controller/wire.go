package controller

import (
	"github.com/goforj/wire"
	// WARN: Integration event handler (Kafka DocumentCommitted) is disabled.
	// TODO: Uncomment once event handler is stable and integrate with app.Start()
	// "github.com/notopia-uit/notopia/internal/note/controller/event"
	"github.com/notopia-uit/notopia/internal/note/controller/grpc"
	"github.com/notopia-uit/notopia/internal/note/controller/http"
)

var ProviderSet = wire.NewSet(
	// NOTE: event.ProviderSet currently disabled - integration event router not started
	// event.ProviderSet,
	grpc.ProviderSet,
	http.ProviderSet,
)
