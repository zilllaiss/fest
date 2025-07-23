package basic

import (
	"context"

	"github.com/zilllaiss/fest"
	"github.com/zilllaiss/fest/examples/views"
	"github.com/zilllaiss/fest/temfest"

	"github.com/a-h/templ"
)

// NOTE: remember to run templ generate

func CreateRoutes(config *fest.GeneratorConfig) *fest.Generator {
	// nil will use the default config.
	// By default, FEST will use the Working Directory as source,
	// unless overrided by configs or FEST_SRC environment variable.
	g := fest.NewGenerator(context.Background(), "My site", config)

	// copy assets
	g.CopyDir("examples/basic/assets", "")
	g.CopyFile("examples/basic/404.html", "")

	// add global style in the header
	g.Head.Add(
		temfest.ImportStyle("/assets/styles.css"),
		temfest.ImportStyle("/assets/nested/styles.css"),
	)

	// add homepage. This will use the site's name as Route title is not set
	g.AddRoute("/", views.Index())

	// we always need this page
	g.AddRoute("/about", views.AboutUs()).
		SetTitle("About Us")

	// you can use any templ component like templ.Raw
	g.AddRoute("/posts/first", templ.Raw("<h1>Hello</h1>")).
		SetTitle("First Post")

	return g
}
