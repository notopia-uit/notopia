package pg

import (
	"context"

	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/internal/note/infra/persistence/pgsqlc"
)

type GetWorkspaceGraphReadModel struct {
	queries *pgsqlc.Queries
}

var _ app.GetWorkspaceGraphReadModel = (*GetWorkspaceGraphReadModel)(nil)

func NewGetWorkspaceGraphReadModel(queries *pgsqlc.Queries) *GetWorkspaceGraphReadModel {
	return &GetWorkspaceGraphReadModel{queries: queries}
}

var ProvideGetWorkspaceGraphReadModel = NewGetWorkspaceGraphReadModel

func (h *GetWorkspaceGraphReadModel) GetWorkspaceGraph(ctx context.Context, q *app.GetWorkspaceGraph) (*app.Graph, error) {
	notes, err := h.queries.GetNotesInWorkspace(ctx, &pgsqlc.GetNotesInWorkspaceParams{
		WorkspaceID:  q.ID,
		TrashedBy:    nil,
		IsNotTrashed: true,
	})
	if err != nil {
		return nil, toErr(err)
	}

	links, err := h.queries.GetNoteLinksInWorkspace(ctx, q.ID)
	if err != nil {
		return nil, toErr(err)
	}

	reachableIDs := make(map[string]bool)

	if q.IgnoreOrphans {
		adj := make(map[string]bool)
		for _, l := range links {
			adj[l.SourceID.String()] = true
			adj[l.TargetID.String()] = true
		}
		for _, n := range notes {
			if len(n.Tags) > 0 || adj[n.ID.String()] {
				reachableIDs[n.ID.String()] = true
				for _, tag := range n.Tags {
					reachableIDs["#"+tag] = true
				}
			}
		}
	} else {
		for _, n := range notes {
			reachableIDs[n.ID.String()] = true
			for _, tag := range n.Tags {
				reachableIDs["#"+tag] = true
			}
		}
	}

	return buildGraph(notes, links, reachableIDs), nil
}
