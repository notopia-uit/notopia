package authorization_test

import (
	"testing"

	"github.com/casbin/casbin/v3"
)

func GetLocalEnforcer() (*casbin.Enforcer, error) {
	e, err := casbin.NewEnforcer("model.conf", "policy_test.csv")
	if err != nil {
		return nil, err
	}
	return e, nil
}

func TestModels(t *testing.T) {
	e, err := GetLocalEnforcer()
	if err != nil {
		t.Fatalf("Failed to create enforcer: %v", err)
	}

	tests := []struct {
		name  string
		sub   string
		dom   string
		obj   string
		act   string
		allow bool
	}{
		{"user:111 (owner) can edit workspace:111", "user:111", "workspace:111", "workspace", "edit", true},
		{"user:111 (owner) can delete note in workspace:111", "user:111", "workspace:111", "note", "delete", true},
		{"user:112 (editor) can write note in workspace:111", "user:112", "workspace:111", "note", "write", true},
		{"user:112 (editor) CANNOT edit workspace:111", "user:112", "workspace:111", "workspace", "edit", false},
		{"user:110 (viewer) can read note in workspace:111", "user:110", "workspace:111", "note", "read", true},
		{"user:110 (viewer) CANNOT write note in workspace:111", "user:110", "workspace:111", "note", "write", false},
		{"user:112 (owner) can delete workspace:112", "user:112", "workspace:112", "workspace", "delete", true},
		{"user:111 (editor) can read workspace:112", "user:111", "workspace:112", "workspace", "read", true},
		{"user:111 (editor) CANNOT delete workspace:112", "user:111", "workspace:112", "workspace", "delete", false},
		{"user:110 (no role) CANNOT read note in workspace:112", "user:110", "workspace:112", "note", "read", false},
		{"user:110 (owner) can edit workspace:110", "user:110", "workspace:110", "workspace", "edit", true},
		{"user:112 (no role) CANNOT read note in workspace:110", "user:112", "workspace:110", "note", "read", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ok, ex, err := e.EnforceEx(tc.sub, tc.dom, tc.obj, tc.act)
			if err != nil {
				t.Fatalf("Enforcer threw an error: %v", err)
			}

			if ok != tc.allow {
				t.Errorf("Expected allow=%v, got %v. Details: %v", tc.allow, ok, ex)
			}
		})
	}
}
