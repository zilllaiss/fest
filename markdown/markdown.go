package markdown

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"go.abhg.dev/goldmark/frontmatter"
	"go.abhg.dev/goldmark/toc"
)

// The type used to process markdown files. It is simple wrapper 
// of goldmark.Markdown.
type MarkdownProcessor struct {
	// The maximum depts of header that will be parsed. Default 3
	TOCMaxDepth int

	md goldmark.Markdown
}

// New Markdown used to parse files.
func NewMarkdown(extensions ...goldmark.Extender) *MarkdownProcessor {
	exts := []goldmark.Extender{&frontmatter.Extender{}}
	exts = append(exts, extensions...)

	m := &MarkdownProcessor{md: goldmark.New(goldmark.WithExtensions(exts...))}

	m.md.Parser().AddOptions(parser.WithAutoHeadingID())
	return m
}

// Parse a markdown file in the specified path.
func (m *MarkdownProcessor) ParseFile(path string) (MarkdownData, error) {
	if m.TOCMaxDepth == 0 {
		m.TOCMaxDepth = 3
	}
	md, err := m.parseFile(path)
	if err != nil {
		return MarkdownData{}, err
	}
	return *md, nil
}

// Parse all markdown files in the specified path.
func (m *MarkdownProcessor) ParseFiles(path string) ([]*MarkdownData, error) {
	if m.TOCMaxDepth == 0 {
		m.TOCMaxDepth = 3
	}

	mdDir, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	f := []*MarkdownData{}

	for _, content := range mdDir {
		if ext := filepath.Ext(content.Name()); content.IsDir() || ext != ".md" {
			continue
		}
		pathWithFilename := filepath.Join(path, content.Name())

		festMd, err := m.parseFile(pathWithFilename)
		if err != nil {
			return nil, err
		}

		f = append(f, festMd)
	}

	return f, nil
}

// Return the original parser. Useful when you want to do something else with it.
func (m *MarkdownProcessor) Goldmark() goldmark.Markdown { return m.md }

func (m *MarkdownProcessor) parseFile(pathWithFilename string) (*MarkdownData, error) {
	ctx := parser.NewContext()
	filename := filepath.Base(pathWithFilename)

	if ext := filepath.Ext(filename); ext != ".md" {
		return nil, errors.New("not an md file")
	}
	md, err := os.ReadFile(pathWithFilename)
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

// The type containing all data parsed from the markdown file.
type MarkdownData struct {
	// Filename excluding the .md extension.
	Slug string

	// Main content excluding the frontmatter.
	Content string

	// Table of Content parsed from headers.
	TOC string

	fm *frontmatter.Data
}

// Decode the frontmatter data. This will panic if data is not a pointer type. 
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
