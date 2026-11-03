package seed

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

func TestSeedDoesNotRunMigrations(t *testing.T) {
	f, err := parser.ParseFile(token.NewFileSet(), "seed.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	ast.Inspect(f, func(n ast.Node) bool {
		if call, ok := n.(*ast.CallExpr); ok {
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok && sel.Sel.Name == "RunMigrations" {
				t.Error("seed must not run migrations")
			}
		}
		return true
	})
}
