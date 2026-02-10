package note

import (
	"github.com/notopia-uit/notopia/pkg/common/metadata"
	"github.com/notopia-uit/notopia/pkg/note/controller"
)

type Server = controller.Server

const (
	// TODO: Let something track this version?
	ServiceName    metadata.ServiceName    = "notopia-note"
	ServiceVersion metadata.ServiceVersion = "v0.0.0"
)
