<p>
  <a href="https://pkg.go.dev/github.com/zilllaiss/fest">
    <img src="https://pkg.go.dev/badge/github.com/zilllaiss/fest" alt="Go Reference"/>
  </a>
</p>

# FEST 

Fantastically Easy Static-site with Templ (FEST) is a minimalistic static-site generator that integrates nicely with Go & [Templ](https://github.com/a-h/templ). It is aimed for Go developers to generate a static-website with existing Go libraries and available knowledge.

## FEATURES

- Write static-site with Go.
- Use any Go libraries in its ecosystem and use it in your Go or Templ files.
- Lightweight, only import Templ as dependency.
- Routing feature that is inspired by router libraries, particularly [Chi](https://github.com/go-chi/chi).

You can see the documentation/API references [here](https://pkg.go.dev/github.com/zilllaiss/fest).

## INSTALL

Initialize your Go project.
```sh
go mod init my-project
```

Since FEST uses Templ, you need to [install Templ](https://templ.guide/quick-start/installation) first. 
```sh
go install https://github.com/a-h/templ/tree/main/cmd/templ@latest
```

Then simply get the library by running `go get`.
```sh
go get github.com/zilllaiss/fest
```

## USAGE

Write your templ files like usual. In this example, I make `views/views.templ` folder and file for it.
```templ
package views 

templ Index() {
    <h1>Hello, World</h1>
}

templ About(name string) {
    <h1></h1>
    <p>My name is {name}. I'm a Go developer</p>
}
```

Import it to your Go files and use it with fest like a router!
```go
package main

import (
	"context"
	"my-project/views"

	"github.com/a-h/templ"
	"github.com/zilllaiss/fest"
	"github.com/zilllaiss/fest/temfest"
)

func main() {
	// nil will use the default config.
	// By default, FEST will use the Working Directory as source,
	// unless overrided by configs or FEST_SRC environment variable.
	g := fest.NewGenerator(context.Background(), "My site", nil)

	// copy assets to the default destionation path. You can specify it with
	g.CopyDir("assets", "")
	g.CopyFile("404.html", "")

	// add global style in the header
	g.Head.Add(
        // fest provides commonly used templates in temfest package
		temfest.ImportStyle("/assets/styles.css"),
		temfest.ImportStyle("/assets/nested/styles.css"),
	)

	// add homepage. This will use the site's name as Route title is not set
	g.AddRoute("/", views.Index())

    // set the page title. If not set, fest will use the site's name by default
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

Then run the binary to generate your website.
```sh 
# if you haven't run templ generate already
templ generate && go run main.go
```

By default it is available in `/dest` folder inside your project directory.

See [examples]("https://github.com/zilllaiss/templ") for more.

## LICENSE

MIT
