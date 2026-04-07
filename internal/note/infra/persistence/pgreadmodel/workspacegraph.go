package pgreadmodel

import (
	"context"

	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/internal/note/infra/persistence/pgsqlc"
)

type GetWorkspaceGraph struct {
	queries *pgsqlc.Queries
}

var _ app.GetWorkspaceGraphReadModel = (*GetWorkspaceGraph)(nil)

func NewGetWorkspaceGraph(queries *pgsqlc.Queries) *GetWorkspaceGraph {
	return &GetWorkspaceGraph{queries: queries}
}

var ProvideGetWorkspaceGraph = NewGetWorkspaceGraph

func (h *GetWorkspaceGraph) GetWorkspaceGraph(ctx context.Context, q *app.GetWorkspaceGraph) (*app.Graph, error) {
	notes, err := h.queries.ReadGetNotesInWorkspace(ctx, pgsqlc.ReadGetNotesInWorkspaceParams{
		WorkspaceID:  q.ID,
		ExcludeTrash: true,
	})
	if err != nil {
		return nil, toErr(err)
	}

	links, err := h.queries.ReadGetNoteLinksInWorkspace(ctx, q.ID)
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
