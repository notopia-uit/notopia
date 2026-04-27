package domain

import (
	"time"

	"github.com/notopia-uit/notopia/internal/note/errs"
)

type TrashedBy uint8

const (
	TrashedByUnspecified TrashedBy = iota
	TrashedByPurpose
	TrashedByParent
)

func (t TrashedBy) IsValid() bool {
	switch t {
	case TrashedByUnspecified, TrashedByPurpose, TrashedByParent:
		return true
	default:
		return false
	}
}

func (t TrashedBy) String() string {
	switch t {
	case TrashedByUnspecified:
		return "unspecified"
	case TrashedByPurpose:
		return "purpose"
	case TrashedByParent:
		return "parent"
	default:
		return "unknown"
	}
}

// NOTE: if need, we can have ParseTrashedBy, but for now mostly we map from outside into

type Trashed struct {
	by TrashedBy
	at time.Time
}

func NewTrashed(by TrashedBy, at time.Time) (Trashed, error) {
	if !by.IsValid() {
		return Trashed{}, errs.NewTrashedInvalid(by.String(), at)
	}
	return Trashed{
		by: by,
		at: at,
	}, nil
}

func NewUntrashed() Trashed {
	return Trashed{
		by: TrashedByUnspecified,
		at: time.Time{},
	}
}

func (t Trashed) By() TrashedBy { return t.by }

func (t Trashed) At() time.Time { return t.at }

func (t Trashed) IsTrashed() bool {
	return t.by != TrashedByUnspecified && !t.at.IsZero()
}
