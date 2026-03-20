package domain

import "time"

type TrashedBy string

var (
	TrashedByUnspecified TrashedBy = "unspecified"
	TrashedByPurpose     TrashedBy = "purpose"
	TrashedByParent      TrashedBy = "parent"
)

func (t TrashedBy) String() string {
	return string(t)
}

type Trashed struct {
	by TrashedBy
	at time.Time
}

func NewTrashed(by TrashedBy, at time.Time) *Trashed {
	return &Trashed{
		by: by,
		at: at,
	}
}
