package pgrepo

import (
	"time"

	"github.com/notopia-uit/notopia/internal/note/domain"
	"github.com/notopia-uit/notopia/internal/note/infra/persistence/pgsqlc"
)

func fromDomainTrashedBy(trashedBy domain.TrashedBy) (*string, bool) {
	switch trashedBy {
	case domain.TrashedByPurpose:
		return new(string(pgsqlc.TrashedByPurpose)), true
	case domain.TrashedByParent:
		return new(string(pgsqlc.TrashedByParent)), true
	case domain.TrashedByUnspecified:
		return nil, false
	default:
		return nil, false
	}
}

func toDomainTrashedBy(trashedBy string) domain.TrashedBy {
	switch trashedBy {
	case string(pgsqlc.TrashedByPurpose):
		return domain.TrashedByPurpose
	case string(pgsqlc.TrashedByParent):
		return domain.TrashedByParent
	default:
		return domain.TrashedByUnspecified
	}
}

func toDomainTrashed(by *string, at *time.Time) domain.Trashed {
	if by == nil || at == nil {
		return domain.NewUntrashed()
	}
	return domain.NewTrashed(toDomainTrashedBy(*by), *at)
}
