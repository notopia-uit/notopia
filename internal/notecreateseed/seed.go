package notecreateseed

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/google/uuid"
)

const (
	defaultNamespace = "6ba7b810-9dad-11d1-80b4-00c04fd430c8"
	workspaceIDStr   = "00000000-0000-4000-8000-000000000110"
	workspaceSlug    = "notopia"
	workspaceName    = "Notopia"
	rootFolderName   = ""
)

var (
	frontmatterRegex  = regexp.MustCompile(`(?s)^---\n.*?\n---\n`)
	wikiLinkRegex     = regexp.MustCompile(`\[\[([^\]]+?)\]\]`)
	markdownLinkRegex = regexp.MustCompile(`\[([^\]]+)\]\(([^\)]+)\)`)
	tagRegex          = regexp.MustCompile(`(^|\s)#([^\s#\)\]]+)`)
)

type Config struct {
	SourceDir string
	OutputSQL string
	Namespace uuid.UUID
}

type Folder struct {
	ID          uuid.UUID
	Name        string
	ParentID    *uuid.UUID
	WorkspaceID uuid.UUID
	Path        string
	Depth       int
}

type Note struct {
	ID       uuid.UUID
	Name     string
	FolderID uuid.UUID
	Tags     []string
	Size     uint64
	Path     string
}

type NoteLink struct {
	SourceID uuid.UUID
	TargetID uuid.UUID
}

func DefaultConfig() (Config, error) {
	namespace, err := uuid.Parse(defaultNamespace)
	if err != nil {
		return Config{}, err
	}
	return Config{
		SourceDir: "./submodule/trshpuppy-obsidian-notes",
		OutputSQL: "./internal/notecreateseed/seed.gen.sql",
		Namespace: namespace,
	}, nil
}

func Run(config Config) error {
	if config.SourceDir == "" || config.OutputSQL == "" {
		return errors.New("source dir and output sql are required")
	}
	workspaceID, err := uuid.Parse(workspaceIDStr)
	if err != nil {
		return fmt.Errorf("invalid workspace id: %w", err)
	}

	markdownFiles, err := collectMarkdownFiles(config.SourceDir)
	if err != nil {
		return err
	}

	noteIDByPath := map[string]uuid.UUID{}
	for _, filePath := range markdownFiles {
		relPath, err := filepath.Rel(config.SourceDir, filePath)
		if err != nil {
			return fmt.Errorf("failed to get relative path for %s: %w", filePath, err)
		}
		relPath = filepath.ToSlash(relPath)
		pathWithoutExt := strings.TrimSuffix(relPath, ".md")
		noteIDByPath[pathWithoutExt] = uuid.NewSHA1(config.Namespace, []byte(pathWithoutExt))
	}

	folders, folderIDByPath, err := buildFolders(markdownFiles, config.SourceDir, config.Namespace, workspaceID)
	if err != nil {
		return err
	}

	notes, noteLinks, err := buildNotes(markdownFiles, config.SourceDir, noteIDByPath, folderIDByPath)
	if err != nil {
		return err
	}

	sql := renderSQL(workspaceID, folders, notes, noteLinks)
	if err := os.WriteFile(config.OutputSQL, []byte(sql), 0o644); err != nil {
		return fmt.Errorf("failed to write seed sql: %w", err)
	}
	return nil
}

func collectMarkdownFiles(sourceDir string) ([]string, error) {
	var markdownFiles []string
	err := filepath.WalkDir(sourceDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if strings.HasPrefix(d.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(strings.ToLower(d.Name()), ".md") {
			markdownFiles = append(markdownFiles, path)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk source directory: %w", err)
	}
	sort.Strings(markdownFiles)
	return markdownFiles, nil
}

func buildFolders(markdownFiles []string, sourceDir string, namespace uuid.UUID, workspaceID uuid.UUID) ([]Folder, map[string]uuid.UUID, error) {
	folderIDByPath := map[string]uuid.UUID{}
	folderPaths := map[string]struct{}{"": {}}

	for _, filePath := range markdownFiles {
		relPath, err := filepath.Rel(sourceDir, filePath)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get relative path for %s: %w", filePath, err)
		}
		relPath = filepath.ToSlash(relPath)
		dir := filepath.Dir(relPath)
		if dir == "." {
			dir = ""
		}
		if dir != "" {
			parts := strings.Split(dir, "/")
			for i := 1; i <= len(parts); i++ {
				folderPaths[strings.Join(parts[:i], "/")] = struct{}{}
			}
		}
	}

	for path := range folderPaths {
		folderIDByPath[path] = uuid.NewSHA1(namespace, []byte(path))
	}

	folders := make([]Folder, 0, len(folderPaths))
	for path, id := range folderIDByPath {
		depth := 0
		if path != "" {
			depth = len(strings.Split(path, "/"))
		}
		name := rootFolderName
		var parentID *uuid.UUID
		if path != "" {
			parts := strings.Split(path, "/")
			name = parts[len(parts)-1]
			parentPath := strings.Join(parts[:len(parts)-1], "/")
			if parentPath == "." {
				parentPath = ""
			}
			pid := folderIDByPath[parentPath]
			parentID = &pid
		}
		folders = append(folders, Folder{
			ID:          id,
			Name:        name,
			ParentID:    parentID,
			WorkspaceID: workspaceID,
			Path:        path,
			Depth:       depth,
		})
	}

	sort.Slice(folders, func(i, j int) bool {
		if folders[i].Depth == folders[j].Depth {
			return folders[i].Path < folders[j].Path
		}
		return folders[i].Depth < folders[j].Depth
	})

	return folders, folderIDByPath, nil
}

func buildNotes(markdownFiles []string, sourceDir string, noteIDByPath map[string]uuid.UUID, folderIDByPath map[string]uuid.UUID) ([]Note, []NoteLink, error) {
	notes := make([]Note, 0, len(markdownFiles))
	linkSet := map[string]NoteLink{}

	for _, filePath := range markdownFiles {
		relPath, err := filepath.Rel(sourceDir, filePath)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get relative path for %s: %w", filePath, err)
		}
		relPath = filepath.ToSlash(relPath)
		pathWithoutExt := strings.TrimSuffix(relPath, ".md")
		noteID := noteIDByPath[pathWithoutExt]

		folderPath := filepath.Dir(relPath)
		if folderPath == "." {
			folderPath = ""
		}
		folderPath = filepath.ToSlash(folderPath)
		folderID := folderIDByPath[folderPath]

		content, err := os.ReadFile(filePath)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
		}

		noteName := strings.TrimSuffix(filepath.Base(relPath), ".md")
		tags := extractTags(string(content))
		size := contentSize(string(content))

		notes = append(notes, Note{
			ID:       noteID,
			Name:     noteName,
			FolderID: folderID,
			Tags:     tags,
			Size:     size,
			Path:     pathWithoutExt,
		})

		relDir := filepath.Dir(relPath)
		if relDir == "." {
			relDir = ""
		}
		targets := extractNoteLinks(string(content), noteIDByPath, relDir)
		for _, target := range targets {
			key := noteID.String() + ":" + target.String()
			linkSet[key] = NoteLink{SourceID: noteID, TargetID: target}
		}
	}

	sort.Slice(notes, func(i, j int) bool {
		return notes[i].Path < notes[j].Path
	})

	noteLinks := make([]NoteLink, 0, len(linkSet))
	for _, link := range linkSet {
		noteLinks = append(noteLinks, link)
	}
	sort.Slice(noteLinks, func(i, j int) bool {
		if noteLinks[i].SourceID == noteLinks[j].SourceID {
			return noteLinks[i].TargetID.String() < noteLinks[j].TargetID.String()
		}
		return noteLinks[i].SourceID.String() < noteLinks[j].SourceID.String()
	})

	return notes, noteLinks, nil
}

func extractTags(content string) []string {
	stripped := frontmatterRegex.ReplaceAllString(content, "")
	matches := tagRegex.FindAllStringSubmatch(stripped, -1)
	if len(matches) == 0 {
		return []string{}
	}
	tagSet := map[string]struct{}{}
	for _, match := range matches {
		if len(match) < 3 {
			continue
		}
		tag := strings.TrimSpace(match[2])
		if tag == "" {
			continue
		}
		tagSet[tag] = struct{}{}
	}
	tags := make([]string, 0, len(tagSet))
	for tag := range tagSet {
		tags = append(tags, tag)
	}
	sort.Strings(tags)
	return tags
}

func extractNoteLinks(content string, noteIDByPath map[string]uuid.UUID, sourceRelDir string) []uuid.UUID {
	stripped := frontmatterRegex.ReplaceAllString(content, "")
	targets := map[uuid.UUID]struct{}{}

	wikiMatches := wikiLinkRegex.FindAllStringSubmatch(stripped, -1)
	for _, match := range wikiMatches {
		if len(match) < 2 {
			continue
		}
		linkContent := match[1]
		parts := strings.SplitN(linkContent, "|", 2)
		linkPath := strings.TrimSpace(parts[0])
		if linkPath == "" {
			continue
		}
		linkPath = stripAnchor(linkPath)
		resolved := resolveLinkPath(linkPath)
		if id, ok := noteIDByPath[resolved]; ok {
			targets[id] = struct{}{}
		} else if sourceRelDir != "" {
			if id, ok := noteIDByPath[sourceRelDir+"/"+resolved]; ok {
				targets[id] = struct{}{}
			}
		}
	}

	markdownMatches := markdownLinkRegex.FindAllStringSubmatch(stripped, -1)
	for _, match := range markdownMatches {
		if len(match) < 3 {
			continue
		}
		linkURL := strings.TrimSpace(match[2])
		if linkURL == "" {
			continue
		}
		if !isRelativeFilePath(linkURL) {
			continue
		}
		linkPath := stripAnchor(linkURL)
		resolved := resolveLinkPath(linkPath)
		if id, ok := noteIDByPath[resolved]; ok {
			targets[id] = struct{}{}
		} else if sourceRelDir != "" {
			if id, ok := noteIDByPath[sourceRelDir+"/"+resolved]; ok {
				targets[id] = struct{}{}
			}
		}
	}

	result := make([]uuid.UUID, 0, len(targets))
	for id := range targets {
		result = append(result, id)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].String() < result[j].String() })
	return result
}

func contentSize(content string) uint64 {
	b, err := json.Marshal(content)
	if err != nil {
		return uint64(len(content))
	}
	return uint64(len(b))
}

func stripAnchor(path string) string {
	if before, _, ok := strings.Cut(path, "#"); ok {
		return before
	}
	return path
}

func isRelativeFilePath(urlPath string) bool {
	if strings.Contains(urlPath, "://") {
		return false
	}
	if strings.HasPrefix(urlPath, "#") {
		return false
	}
	trimmed := strings.TrimSuffix(urlPath, ".md")
	hasMdExt := len(urlPath) > len(trimmed) && urlPath[len(trimmed):] == ".md"
	startsWithRelative := strings.HasPrefix(urlPath, "../") || strings.HasPrefix(urlPath, "./") || strings.HasPrefix(urlPath, "/")
	return hasMdExt || startsWithRelative
}

func resolveLinkPath(linkPath string) string {
	linkPath = strings.TrimSpace(linkPath)
	linkPath = strings.TrimSuffix(linkPath, ".md")
	linkPath = strings.TrimPrefix(linkPath, "./")
	linkPath = strings.TrimPrefix(linkPath, "/")
	for strings.HasPrefix(linkPath, "../") {
		linkPath = strings.TrimPrefix(linkPath, "../")
	}
	linkPath = filepath.ToSlash(filepath.Clean(linkPath))
	if linkPath == "." {
		return ""
	}
	return linkPath
}

func renderSQL(workspaceID uuid.UUID, folders []Folder, notes []Note, links []NoteLink) string {
	var b strings.Builder
	b.WriteString("-- Generated by cmd/notecreateseed\n")
	b.WriteString("-- Workspace\n")
	fmt.Fprintf(&b, "INSERT INTO workspaces (id, slug, name) VALUES ('%s', '%s', '%s');\n\n", workspaceID, sqlString(workspaceSlug), sqlString(workspaceName))

	b.WriteString("-- Folders\n")
	if len(folders) > 0 {
		b.WriteString("INSERT INTO folders (id, name, icon, workspace_id, parent_id) VALUES\n")
		for i, folder := range folders {
			parent := "NULL"
			if folder.ParentID != nil {
				parent = fmt.Sprintf("'%s'", folder.ParentID.String())
			}
			fmt.Fprintf(&b, "  ('%s', '%s', NULL, '%s', %s)",
				folder.ID,
				sqlString(folder.Name),
				folder.WorkspaceID,
				parent)
			if i < len(folders)-1 {
				b.WriteString(",")
			}
			b.WriteString("\n")
		}
		b.WriteString(";\n\n")
	}

	b.WriteString("-- Notes\n")
	if len(notes) > 0 {
		b.WriteString("INSERT INTO notes (id, name, icon, folder_id, tags, size) VALUES\n")
		for i, note := range notes {
			tagsSQL := "ARRAY[]::text[]"
			if len(note.Tags) > 0 {
				tags := make([]string, 0, len(note.Tags))
				for _, tag := range note.Tags {
					tags = append(tags, fmt.Sprintf("'%s'", sqlString(tag)))
				}
				tagsSQL = fmt.Sprintf("ARRAY[%s]::text[]", strings.Join(tags, ", "))
			}
			fmt.Fprintf(&b, "  ('%s', '%s', NULL, '%s', %s, %d)",
				note.ID,
				sqlString(note.Name),
				note.FolderID,
				tagsSQL,
				note.Size)
			if i < len(notes)-1 {
				b.WriteString(",")
			}
			b.WriteString("\n")
		}
		b.WriteString(";\n\n")
	}

	b.WriteString("-- Note links\n")
	if len(links) > 0 {
		b.WriteString("INSERT INTO note_links (source_id, target_id) VALUES\n")
		for i, link := range links {
			fmt.Fprintf(&b, "  ('%s', '%s')",
				link.SourceID,
				link.TargetID)
			if i < len(links)-1 {
				b.WriteString(",")
			}
			b.WriteString("\n")
		}
		b.WriteString(";\n")
	}
	return b.String()
}

func sqlString(value string) string {
	return strings.ReplaceAll(value, "'", "''")
}
