package main

import (
	"flag"
	"log"

	"github.com/notopia-uit/notopia/internal/notecreateseed"
)

func main() {
	sourceDir := flag.String("source", "./submodule/trshpuppy-obsidian-notes", "Obsidian vault directory")
	outputSQL := flag.String("output", "./internal/notecreateseed/seed.gen.sql", "Output seed sql path")
	flag.Parse()

	config, err := notecreateseed.DefaultConfig()
	if err != nil {
		log.Fatalf("failed to init config: %v", err)
	}
	config.SourceDir = *sourceDir
	config.OutputSQL = *outputSQL

	if err := notecreateseed.Run(config); err != nil {
		log.Fatalf("failed to generate seed: %v", err)
	}
}
