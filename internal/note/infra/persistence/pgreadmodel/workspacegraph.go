package pgreadmodel

import (
	"context"

	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/internal/note/infra/persistence/pgsqlc"
)

type WorkspaceGraph struct {
	queries *pgsqlc.Queries
}

var _ app.GetWorkspaceGraphReadModel = (*WorkspaceGraph)(nil)

func GetWorkspaceGraph(queries *pgsqlc.Queries) *WorkspaceGraph {
	return &WorkspaceGraph{queries: queries}
}

var ProvideWorkspaceGraph = GetWorkspaceGraph

func (h *WorkspaceGraph) Handle(ctx context.Context, p *app.GetWorkspaceGraphReadModelParams) (app.Graph, error) {
	notes, err := h.queries.ReadGetNotesInWorkspace(ctx, pgsqlc.ReadGetNotesInWorkspaceParams{
		WorkspaceID:  p.ID,
		ExcludeTrash: true,
	})
	if err != nil {
		return app.Graph{}, toErr(err)
	}

	links, err := h.queries.ReadGetNoteLinksInWorkspace(ctx, p.ID)
	if err != nil {
		return app.Graph{}, toErr(err)
	}

	reachableIDs := make(map[string]bool)

	if p.IgnoreOrphans {
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
