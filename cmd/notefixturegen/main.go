package main

import (
	"flag"
	"log"

	"github.com/notopia-uit/notopia/internal/notefixturegen"
)

func main() {
	sourceDir := flag.String("source", "./submodule/trshpuppy-obsidian-notes", "Obsidian vault directory")
	outputSQL := flag.String("output", "./internal/notefixturegen/fixtures.gen.sql", "Output fixtures sql path")
	flag.Parse()

	config, err := notefixturegen.DefaultConfig()
	if err != nil {
		log.Fatalf("failed to init config: %v", err)
	}
	config.SourceDir = *sourceDir
	config.OutputSQL = *outputSQL

	if err := notefixturegen.Run(config); err != nil {
		log.Fatalf("failed to generate fixtures: %v", err)
	}
}
