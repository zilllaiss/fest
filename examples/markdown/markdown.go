package markdown

import (
	"context"
	"time"

	"github.com/a-h/templ"
	"github.com/zilllaiss/fest"
	"github.com/zilllaiss/fest/examples/views"
	"github.com/zilllaiss/fest/markdown"
)

func UseMarkdown(config *fest.GeneratorConfig) error {
	m := markdown.NewMarkdown()

	md, err := m.ParseFile("examples/markdown/something.md")
	if err != nil {
		return err
	}

	var p Post
	if err := md.GetFrontmatter(&p); err != nil {
		return err
	}

	mds, err := m.ParseFiles("examples/markdown/posts")
	if err != nil {
		return err
	}

	g := fest.NewGenerator(context.Background(), "My site", config)

	g.AddRoute("/post/"+md.Slug, postComp(p.Title, md.Slug, md.Content))

	fest.NewRoutesT("/post/{s}", mds).
		SetTitle("{s}").AddToGenerator(g, postsFn)

	if err := g.Generate(); err != nil {
		return err
	}
	return nil
}

type Post struct {
	Title     string    `yaml:"title"`
	Published time.Time `yaml:"published"`
	Tags      []string  `yaml:"tags"`
}

func postsFn(
	ctx context.Context,
	rp *fest.RouteParam[*markdown.MarkdownData],
) (templ.Component, error) {
	md := rp.GetItem()
	rp.SetSlug(md.Slug)

	var post Post
	if err := md.GetFrontmatter(&post); err != nil {
		return nil, err
	}

	return postComp(post.Title, md.Slug, md.Content), nil
}

func postComp(title, slug, content string) templ.Component {
	contentComp := templ.Raw(content)
	return templ.Join(views.Article(title, slug), contentComp)
}
