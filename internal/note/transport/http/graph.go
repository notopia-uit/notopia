package http

import (
	"context"
	"errors"

	"github.com/notopia-uit/notopia/pkg/api/note"
)

func (h *StrictHandler) GetGraph(
	ctx context.Context,
	request note.GetGraphRequestObject,
) (note.GetGraphResponseObject, error) {
	return nil, errors.New("not implemented")
}
