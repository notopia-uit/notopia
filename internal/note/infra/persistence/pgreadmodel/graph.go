package pgreadmodel

import (
	"math"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/internal/note/infra/persistence/pgsqlc"
)

func calculateGraphWeight(size, minSize, maxSize int32) float32 {
	var w float32
	if maxSize == minSize {
		w = 1
	} else {
		w = float32(size-minSize) / float32(maxSize-minSize)
	}
	res := 0.5 + w*0.5
	return float32(int((res*10)+0.5)) / 10.0
}

func buildGraph(notes []*pgsqlc.Note, links []*pgsqlc.NoteLink, reachableIDs map[string]bool) app.Graph {
	var minSize int32 = math.MaxInt32
	var maxSize int32 = -1
	reachableNotesMap := make(map[uuid.UUID]*pgsqlc.Note)

	for _, n := range notes {
		if reachableIDs[n.ID.String()] {
			reachableNotesMap[n.ID] = n
			if n.Size < minSize {
				minSize = n.Size
			}
			if n.Size > maxSize {
				maxSize = n.Size
			}
		}
	}

	var graphNodes []app.GraphNode
	var graphLinks []app.GraphLink
	tagsAdded := make(map[string]bool)

	for _, n := range reachableNotesMap {
		graphNodes = append(graphNodes, app.GraphNode{
			ID:     n.ID.String(),
			Name:   n.Name,
			Type:   app.GraphNodeTypeNote,
			Weight: calculateGraphWeight(n.Size, minSize, maxSize),
		})

		for _, tag := range n.Tags {
			tagID := "#" + tag

			if reachableIDs[tagID] {
				if !tagsAdded[tagID] {
					graphNodes = append(graphNodes, app.GraphNode{
						ID:     tagID,
						Name:   tag,
						Type:   app.GraphNodeTypeTag,
						Weight: 0,
					})
					tagsAdded[tagID] = true
				}

				graphLinks = append(graphLinks, app.GraphLink{
					Source: n.ID.String(),
					Target: tagID,
				})
			}
		}
	}

	for _, l := range links {
		if reachableIDs[l.SourceID.String()] && reachableIDs[l.TargetID.String()] {
			graphLinks = append(graphLinks, app.GraphLink{
				Source: l.SourceID.String(),
				Target: l.TargetID.String(),
			})
		}
	}

	return app.Graph{
		Nodes: graphNodes,
		Links: graphLinks,
	}
}
