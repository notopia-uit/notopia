package pgreadmodel

import (
	"context"

	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/internal/note/infra/persistence/pgsqlc"
)

type NoteGraph struct {
	queries *pgsqlc.Queries
}

var _ app.GetNoteGraphReadModel = (*NoteGraph)(nil)

func NewNoteGraph(queries *pgsqlc.Queries) *NoteGraph {
	return &NoteGraph{queries: queries}
}

var ProvideNoteGraph = NewNoteGraph

func (h *NoteGraph) GetNoteGraph(ctx context.Context, q *app.GetNoteGraph) (*app.Graph, error) {
	workspaceID, err := h.queries.GetWorkspaceIDByNoteID(ctx, q.ID)
	if err != nil {
		return nil, toErr(err)
	}

	notes, err := h.queries.ReadGetNotesInWorkspace(ctx, pgsqlc.ReadGetNotesInWorkspaceParams{
		WorkspaceID:  workspaceID,
		ExcludeTrash: true,
	})
	if err != nil {
		return nil, toErr(err)
	}

	links, err := h.queries.ReadGetNoteLinksInWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, toErr(err)
	}

	adj := make(map[string][]string)
	for _, n := range notes {
		for _, tag := range n.Tags {
			tagID := "#" + tag
			adj[n.ID.String()] = append(adj[n.ID.String()], tagID)
			adj[tagID] = append(adj[tagID], n.ID.String())
		}
	}
	for _, l := range links {
		adj[l.SourceID.String()] = append(adj[l.SourceID.String()], l.TargetID.String())
		adj[l.TargetID.String()] = append(adj[l.TargetID.String()], l.SourceID.String())
	}

	reachableIDs := make(map[string]bool)
	type queueItem struct {
		id    string
		depth int
	}

	startID := q.ID.String()
	queue := []queueItem{{id: startID, depth: 0}}

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		if reachableIDs[curr.id] {
			continue
		}
		reachableIDs[curr.id] = true

		if curr.depth < q.Depth {
			for _, neighbor := range adj[curr.id] {
				if !reachableIDs[neighbor] {
					queue = append(queue, queueItem{id: neighbor, depth: curr.depth + 1})
				}
			}
		}
	}

	return buildGraph(notes, links, reachableIDs), nil
}
