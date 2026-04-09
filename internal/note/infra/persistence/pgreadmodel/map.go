package pgreadmodel

import (
	"fmt"

	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/internal/note/errs"
	"github.com/notopia-uit/notopia/internal/note/infra/persistence/pgsqlc"
)

func toAppTrashedBy(trashedBy *string) (app.TrashedBy, error) {
	if trashedBy == nil {
		return app.TrashedByUnspecified, errs.NewPersistenceInternal("trashedBy is nil", nil)
	}
	switch *trashedBy {
	case string(pgsqlc.TrashedByPurpose):
		return app.TrashedByPurpose, nil
	case string(pgsqlc.TrashedByParent):
		return app.TrashedByParent, nil
	default:
		return app.TrashedByUnspecified, errs.NewPersistenceInternal(fmt.Sprintf("unknown trashedBy value: %s", *trashedBy), nil)
	}
}
