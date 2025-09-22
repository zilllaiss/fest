// Package markdown provides utilities used for parsing mardown files.
package markdown

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"go.abhg.dev/goldmark/frontmatter"
	"go.abhg.dev/goldmark/toc"
)

type MarkdownParser interface {
	// ParseFile parses the set markdown file.
	ParseFile(path string) (*MarkdownData, error)

	// ParseFiles parses all markdown files in a path.
	ParseFiles(path string) ([]*MarkdownData, error)
}

// NewMarkdown initializes a new markdown parser with default configurations.
func NewMarkdown(options ...goldmark.Option) MarkdownParser {
	opts := []goldmark.Option{goldmark.WithExtensions(&frontmatter.Extender{})}
	opts = append(opts, options...)

	m := &MarkdownProcessor{md: goldmark.New(opts...)}

	m.md.Parser().AddOptions(parser.WithAutoHeadingID())
	return m
}

// MarkdownProcessor is the default FEST markdown parser. It is simple wrapper
// of goldmark.Markdown.
type MarkdownProcessor struct {
	// The maximum depts of header that will be parsed. Default 3
	TOCMaxDepth int

	md goldmark.Markdown
}

// ParseFile parses a markdown file in path.
func (m *MarkdownProcessor) ParseFile(path string) (*MarkdownData, error) {
	if m.TOCMaxDepth == 0 {
		m.TOCMaxDepth = 3
	}
	ctx := parser.NewContext()
	filename := filepath.Base(path)

	if ext := filepath.Ext(filename); ext != ".md" {
		return nil, fmt.Errorf("not an md file: %v", filename)
	}

	md, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var mdBuf bytes.Buffer
	var tocBuf bytes.Buffer

	doc := m.md.Parser().Parse(text.NewReader(md), parser.WithContext(ctx))

	tree, err := toc.Inspect(doc, md, toc.MaxDepth(m.TOCMaxDepth), toc.Compact(true))
	if err != nil {
		return nil, err
	}

	list := toc.RenderList(tree)
	if list != nil {
		if err := m.md.Renderer().Render(&tocBuf, md, list); err != nil {
			return nil, err
		}
	}

	if err := m.md.Renderer().Render(&mdBuf, md, doc); err != nil {
		return nil, err
	}

	var festMd MarkdownData
	ptrFestMd := &festMd

	metaData := frontmatter.Get(ctx)
	ptrFestMd.fm = metaData

	ptrFestMd.Slug = strings.TrimSuffix(filename, ".md")
	ptrFestMd.Content = mdBuf.String()
	ptrFestMd.TOC = tocBuf.String()

	return ptrFestMd, nil
}

// ParseFiles parse all markdown files in path. Internally, this calls
// ScanForMarkdown, ParseFile, and skip any nil MarkdownData.
func (m *MarkdownProcessor) ParseFiles(path string) ([]*MarkdownData, error) {
	f := []*MarkdownData{}

	if err := ScanForMarkdown(path, func(p string) error {
		festMd, err := m.ParseFile(p)
		if err != nil {
			return err
		}
		if festMd != nil {
			f = append(f, festMd)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return f, nil
}

// Unwrap returns the original goldmark.Markdown.
func (m *MarkdownProcessor) Unwrap() goldmark.Markdown { return m.md }

// MarkdownData contains all data parsed from the markdown file.
type MarkdownData struct {
	// Filename excluding the .md extension.
	Slug string

	// Main content excluding the frontmatter.
	Content string

	// Table of Content parsed from headers.
	TOC string

	fm *frontmatter.Data
}

// GetFrontmatter decodes the frontmatter data. This will panic if data is not a pointer type.
//
// Be sure to have the frontmatter seperators (i.e. - or +) to be equal in amount
// for both sides, and don't have leading spaces, as it is one of the most
// commons cause for the frontmatter not properly processed.
func (md *MarkdownData) GetFrontmatter(data any) error {
	if md.fm == nil {
		return errors.New("frontmatter not found")
	}
	if err := md.fm.Decode(data); err != nil {
		return err
	}
	return nil
}

// ScanForMarkdown scan a path for markdown files and pass the path and the filename
// to fn as p. If there is an error, it will be os.ReadDir error or an error from
// fn
func ScanForMarkdown(path string, fn func(p string) error) error {
	msg := "error scanning for markdown"
	mdDir, err := os.ReadDir(path)
	if err != nil {
		return fmt.Errorf("%v: %w", msg, err)
	}

	for _, content := range mdDir {
		if ext := filepath.Ext(content.Name()); content.IsDir() || ext != ".md" {
			continue
		}
		p := filepath.Join(path, content.Name())

		if err := fn(p); err != nil {
			return fmt.Errorf("%v: %w", msg, err)
		}
	}
	return nil
}
