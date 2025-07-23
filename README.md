# FEST 

Fantastically Easy Static-site with Templ (FEST) is a minimalistic static-site generator that integrates nicely with Go & Templ.

## FEATURES

- Lightweight, only import Templ as dependencies.
- Use any Go libraries in is ecosystem and use it in your Go or Templ file.
- Routing features that is inspired router libraries, particularly [Chi](https://github.com/go-chi/chi).

## INSTALL

To get the library, simply run `go get`.
```sh
go get github.com/zilllaiss/fest
```

## USAGE

Basic example:
```go
package main

import (
	"context"
	"fest"
	"fest/examples/views"
	"fest/temfest"

	"github.com/a-h/templ"
)

// NOTE: remember to run templ generate :)

func main() {
	// nil will use the default config.
	// By default, FEST will use the Working Directory as source,
	// unless overrided by configs or FEST_SRC environment variable.
	g := fest.NewGenerator(context.Background(), "My site", nil)

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

	// render all components
	if err := g.Generate(); err != nil {
		panic(err)
    }
}
```

See examples for more.

## LICENSE

MIT
