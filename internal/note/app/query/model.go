package query

import (
	"time"

	"github.com/google/uuid"
)

type Note struct {
	Id                 uuid.UUID
	Name               string
	Icon               *string
	Tags               []string
	FolderId           uuid.UUID
	BacklinksCount     int
	OutgoingLinksCount int
	UpdatedAt          time.Time
}
