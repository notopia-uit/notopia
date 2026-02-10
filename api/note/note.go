package note

import (
	"embed"
	"net/url"

	"github.com/getkin/kin-openapi/openapi3"
)

//go:embed dist/openapi.yaml
var spec embed.FS

func GetOpenAPI(loader *openapi3.Loader) (*openapi3.T, error) {
	if loader == nil {
		loader = openapi3.NewLoader()
	}
	loader.IsExternalRefsAllowed = true
	loader.ReadFromURIFunc = func(loader *openapi3.Loader, uri *url.URL) ([]byte, error) {
		return spec.ReadFile(uri.Path)
	}
	return loader.LoadFromFile("dist/openapi.yaml")
}
