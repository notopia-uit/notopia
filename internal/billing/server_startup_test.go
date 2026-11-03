package billing

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

func TestServerRunDoesNotRunMigrations(t *testing.T) {
	f, err := parser.ParseFile(token.NewFileSet(), "server.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	ast.Inspect(f, func(n ast.Node) bool {
		if call, ok := n.(*ast.CallExpr); ok {
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok && (sel.Sel.Name == "RunMigrations" || sel.Sel.Name == "EnsureDefaults") {
				t.Error("server must not run migrations or seed defaults")
			}
		}
		return true
	})
}
