package policy

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParsePolicyCSV(t *testing.T) {
	m, err := parsePolicyCSV("# comment\n\np, owner, workspace, read\n  p, viewer, workspace, read  \ng2, note, workspace_item\n" +
		"g, user:1, owner, workspace:w1\n")
	require.NoError(t, err)

	p, err := m.GetPolicy("p", "p")
	require.NoError(t, err)
	assert.Equal(t, [][]string{{"owner", "workspace", "read"}, {"viewer", "workspace", "read"}}, p)

	g, err := m.GetPolicy("g", "g")
	require.NoError(t, err)
	assert.Equal(t, [][]string{{"user:1", "owner", "workspace:w1"}}, g)

	g2, err := m.GetPolicy("g", "g2")
	require.NoError(t, err)
	assert.Equal(t, [][]string{{"note", "workspace_item"}}, g2)
}

func TestParsePolicyCSVEmbeddedFiles(t *testing.T) {
	m, err := parsePolicyCSV(CSV)
	require.NoError(t, err)
	p, err := m.GetPolicy("p", "p")
	require.NoError(t, err)
	assert.Len(t, p, 16)
	g2, err := m.GetPolicy("g", "g2")
	require.NoError(t, err)
	assert.Len(t, g2, 2)

	_, err = parsePolicyCSV(FixturesCSV)
	require.NoError(t, err)
}

func TestParsePolicyCSVRejectsUnknownType(t *testing.T) {
	_, err := parsePolicyCSV("g3, a, b\n")
	assert.Error(t, err)
}
