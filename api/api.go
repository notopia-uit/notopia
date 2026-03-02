package api

import (
	"embed"
	"net/url"

	"github.com/getkin/kin-openapi/openapi3"
)

//go:embed bundled/*
var specs embed.FS

type Spec string

var (
	NoteSpec     Spec = "note.json"
	DocumentSpec Spec = "document.json"
	OpenapiSpec  Spec = "openapi.json"
)

func GetSpec(loader *openapi3.Loader, spec Spec) (*openapi3.T, error) {
	if loader == nil {
		loader = openapi3.NewLoader()
	}
	loader.IsExternalRefsAllowed = true
	loader.ReadFromURIFunc = func(loader *openapi3.Loader, uri *url.URL) ([]byte, error) {
		return specs.ReadFile(uri.Path)
	}
	return loader.LoadFromFile("bundled/" + string(spec))
}
