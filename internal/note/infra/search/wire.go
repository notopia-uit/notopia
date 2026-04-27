package search

import (
	"github.com/goforj/wire"
	"github.com/notopia-uit/notopia/internal/note/app"
)

var MeilisearchProviderSet = wire.NewSet(
	ProvideMeilisearch,
	wire.Bind(new(app.SearchSvc), new(*Meilisearch)),
)
