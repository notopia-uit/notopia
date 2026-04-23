package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

const (
	// Default UUID namespace - using a fixed namespace UUID for consistent output
	defaultNamespace = "6ba7b810-9dad-11d1-80b4-00c04fd430c8" // UUID v4 constant
)

var (
	// Regex patterns
	frontmatterRegex  = regexp.MustCompile(`(?s)^---\n.*?\n---\n`)
	wikiLinkRegex     = regexp.MustCompile(`\[\[([^\]]+?)\]\]`)
	markdownLinkRegex = regexp.MustCompile(`\[([^\]]+)\]\(([^\)]+)\)`)
	// Match tags as standalone #word patterns (word boundaries)
	tagRegex = regexp.MustCompile(`(^|\s)#([^\s#\)\]]+)`)
	// Regex to detect if a link is a relative file path (not a URL or anchor)
	relativePathRegex = regexp.MustCompile(`^(\.{1,2}/|[^/:]+).*\.md($|#)`)
)

type Config struct {
	SourceDir     string
	OutputDir     string
	MaxGoroutines int
	NamespaceUUID uuid.UUID
}

type FileMapping struct {
	SourcePath   string
	RelativePath string
	UUID         uuid.UUID
	Content      string
}

func main() {
	// Parse command line arguments
	maxGoroutines := flag.Int("goroutines", 50, "Maximum number of concurrent goroutines")
	namespaceStr := flag.String("namespace", defaultNamespace, "UUID namespace for generating file UUIDs")
	flag.Parse()

	// Parse namespace UUID
	namespaceUUID, err := uuid.Parse(*namespaceStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid namespace UUID: %v\n", err)
		os.Exit(1)
	}

	config := Config{
		SourceDir:     "../../../submodule/trshpuppy-obsidian-notes",
		OutputDir:     "./seed-data",
		MaxGoroutines: *maxGoroutines,
		NamespaceUUID: namespaceUUID,
	}

	if err := run(config); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Transformation completed successfully!")
}

func run(config Config) error {
	// Clean and create output directory
	if err := os.RemoveAll(config.OutputDir); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to clean output directory: %w", err)
	}
	if err := os.MkdirAll(config.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Collect all markdown files
	var markdownFiles []string
	err := filepath.WalkDir(config.SourceDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(strings.ToLower(d.Name()), ".md") {
			markdownFiles = append(markdownFiles, path)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to walk source directory: %w", err)
	}

	fmt.Printf("Found %d markdown files\n", len(markdownFiles))

	// First pass: build file mapping (path -> UUID)
	fileMapping := make(map[string]uuid.UUID)
	for _, filePath := range markdownFiles {
		relPath, err := filepath.Rel(config.SourceDir, filePath)
		if err != nil {
			return fmt.Errorf("failed to get relative path for %s: %w", filePath, err)
		}
		// Strip .md extension and normalize
		pathWithoutExt := strings.TrimSuffix(relPath, ".md")
		fileUUID := uuid.NewSHA1(config.NamespaceUUID, []byte(pathWithoutExt))
		fileMapping[pathWithoutExt] = fileUUID
	}

	// Second pass: process files in parallel
	var g errgroup.Group
	g.SetLimit(config.MaxGoroutines)

	for _, filePath := range markdownFiles {
		filePath := filePath // capture loop variable
		g.Go(func() error {
			return processFile(filePath, config, fileMapping)
		})
	}

	if err := g.Wait(); err != nil {
		return fmt.Errorf("failed to process files: %w", err)
	}

	return nil
}

func processFile(filePath string, config Config, fileMapping map[string]uuid.UUID) error {
	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	// Get relative path and UUID
	relPath, err := filepath.Rel(config.SourceDir, filePath)
	if err != nil {
		return fmt.Errorf("failed to get relative path for %s: %w", filePath, err)
	}
	pathWithoutExt := strings.TrimSuffix(relPath, ".md")
	fileUUID := fileMapping[pathWithoutExt]

	// Transform content
	transformed := transformContent(string(content), config, fileMapping)

	// Write output file
	outputFileName := fmt.Sprintf("%s.md", fileUUID.String())
	outputPath := filepath.Join(config.OutputDir, outputFileName)
	if err := os.WriteFile(outputPath, []byte(transformed), 0644); err != nil {
		return fmt.Errorf("failed to write output file %s: %w", outputPath, err)
	}

	return nil
}

func transformContent(content string, config Config, fileMapping map[string]uuid.UUID) string {
	// Step 1: Strip frontmatter
	content = frontmatterRegex.ReplaceAllString(content, "")

	// Step 2: Replace markdown links with relative file paths
	content = markdownLinkRegex.ReplaceAllStringFunc(content, func(match string) string {
		// Extract link text and URL using capturing groups
		parts := markdownLinkRegex.FindStringSubmatch(match)
		if len(parts) < 3 {
			return match
		}
		linkText := parts[1]
		linkURL := parts[2]

		// Check if this is a relative file path (not a URL or anchor-only)
		if isRelativeFilePath(linkURL) {
			// Extract the file path (remove anchor if present)
			filePath := linkURL
			if idx := strings.Index(filePath, "#"); idx != -1 {
				filePath = filePath[:idx]
			}

			// Resolve the absolute path relative to vault root
			resolvedPath := resolveLinkPath(filePath)

			// Look up UUID for this path
			linkUUID, found := fileMapping[resolvedPath]
			if !found {
				// If not found, generate UUID from the path anyway
				linkUUID = uuid.NewSHA1(config.NamespaceUUID, []byte(resolvedPath))
			}

			// Return the JSX replacement (condensed to one line)
			return fmt.Sprintf(`<a href="@%s" data-notopia-ref="%s">%s</a>`, linkUUID.String(), linkUUID.String(), linkText)
		}

		// Not a relative file path, keep original
		return match
	})

	// Step 3: Protect remaining markdown links from tag processing by temporarily replacing them
	markdownLinks := make(map[string]string)
	linkCounter := 0
	content = markdownLinkRegex.ReplaceAllStringFunc(content, func(match string) string {
		placeholder := fmt.Sprintf("__MDLINK_%d__", linkCounter)
		markdownLinks[placeholder] = match
		linkCounter++
		return placeholder
	})

	// Step 4: Replace wiki-style links
	content = wikiLinkRegex.ReplaceAllStringFunc(content, func(match string) string {
		// Extract the link content (without [[ and ]])
		linkContent := match[2 : len(match)-2]

		// Split by | to handle aliases [[file|alias]]
		parts := strings.SplitN(linkContent, "|", 2)
		linkPath := parts[0]

		// Resolve the absolute path relative to vault root
		resolvedPath := resolveLinkPath(linkPath)

		// Look up UUID for this path
		linkUUID, found := fileMapping[resolvedPath]
		if !found {
			// If not found, generate UUID from the path anyway
			linkUUID = uuid.NewSHA1(config.NamespaceUUID, []byte(resolvedPath))
		}

		// Return the JSX replacement (condensed to one line)
		return fmt.Sprintf(`<a href="@%s" data-notopia-ref="%s">@%s</a>`, linkUUID.String(), linkUUID.String(), linkUUID.String())
	})

	// Step 5: Replace tags (now markdown links are protected)
	content = tagRegex.ReplaceAllStringFunc(content, func(match string) string {
		// The regex captures whitespace/start and the tag
		hashIdx := strings.Index(match, "#")
		if hashIdx == -1 {
			return match
		}

		prefix := match[:hashIdx]
		tag := match[hashIdx+1:]

		// Return the prefix plus JSX replacement (condensed to one line)
		return prefix + fmt.Sprintf(`<a href="#%s" data-notopia-tag="%s">#%s</a>`, tag, tag, tag)
	})

	// Step 6: Restore remaining markdown links
	for placeholder, original := range markdownLinks {
		content = strings.ReplaceAll(content, placeholder, original)
	}

	return content
}

func isRelativeFilePath(urlPath string) bool {
	// Check if the path looks like a relative file path
	// It should NOT be:
	// - A URL (contains ://)
	// - Just an anchor (starts with #)
	
	if strings.Contains(urlPath, "://") {
		return false // It's a URL like http://, https://, etc.
	}
	
	if strings.HasPrefix(urlPath, "#") {
		return false // It's just an anchor
	}
	
	// Check if it looks like a file path (has .md or starts with ../, ./, or just a name)
	trimmed := strings.TrimSuffix(urlPath, ".md")
	hasMdExt := len(urlPath) > len(trimmed) && urlPath[len(trimmed):] == ".md"
	
	startsWithRelative := strings.HasPrefix(urlPath, "../") || strings.HasPrefix(urlPath, "./") || strings.HasPrefix(urlPath, "/")
	
	// If it has .md extension or starts with relative path, it's likely a file
	return hasMdExt || startsWithRelative
}

func resolveLinkPath(linkPath string) string {
	// Normalize the path
	linkPath = strings.TrimSpace(linkPath)

	// Remove .md extension if present
	linkPath = strings.TrimSuffix(linkPath, ".md")

	// Handle different path formats
	// Remove leading ./ or / or ../
	linkPath = strings.TrimPrefix(linkPath, "./")
	linkPath = strings.TrimPrefix(linkPath, "/")

	// For ../ paths, we need to resolve them relative to the file
	// But since we're working with vault-relative paths, we normalize
	for strings.HasPrefix(linkPath, "../") {
		linkPath = strings.TrimPrefix(linkPath, "../")
	}

	// Clean the path
	linkPath = filepath.Clean(linkPath)

	return linkPath
}
