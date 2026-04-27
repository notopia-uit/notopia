package search

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/meilisearch/meilisearch-go"
	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/internal/note/config"
	"github.com/notopia-uit/notopia/internal/note/errs"
)

type Meilisearch struct {
	client                   meilisearch.ServiceManager
	noteSearchKeyUID         string
	noteIndexName            string
	noteSearchExpireDuration int
}

var _ app.SearchSvc = (*Meilisearch)(nil)

func NewMeilisearch(cfg *config.Meilisearch) *Meilisearch {
	client := meilisearch.New(cfg.Host, meilisearch.WithAPIKey(cfg.APIKey))
	return &Meilisearch{
		client:                   client,
		noteSearchKeyUID:         cfg.NoteSearchKeyUID,
		noteIndexName:            cfg.NoteIndexName,
		noteSearchExpireDuration: cfg.NoteSearchExpireDuration,
	}
}

var ProvideMeilisearch = NewMeilisearch

func (m *Meilisearch) GenerateWorkspaceToken(ctx context.Context, workspaceID uuid.UUID) (app.SearchToken, error) {
	rules := map[string]any{
		m.noteIndexName: map[string]string{
			"filter": "workspaceId = " + workspaceID.String(),
		},
	}
	expiresAt := time.Now().Add(time.Duration(m.noteSearchExpireDuration) * time.Second)
	token, err := m.client.GenerateTenantToken(m.noteSearchKeyUID, rules, &meilisearch.TenantTokenOptions{
		ExpiresAt: expiresAt,
	})
	if err != nil {
		return app.SearchToken{}, errs.NewFailedToGenerateWorkspaceSearchToken(workspaceID, err)
	}
	return app.SearchToken{
		Token:     token,
		ExpiresAt: expiresAt,
	}, nil
}
