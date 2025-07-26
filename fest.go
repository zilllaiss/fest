package fest

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/zilllaiss/fest/temfest"

	"github.com/a-h/templ"
)

var wd = func() string {
	w, _ := os.Getwd()
	return w
}()

// Site title options
type SiteTitleOption int

const (
	// Show the site title at the end with dash as seperator
	SiteTitleDashBack SiteTitleOption = iota

	// Show the site title at the beginning with dash as seperator
	SiteTitleDashFront

	// Show the site title at the back with a colon as seperator
	SiteTitleColonBack

	// Show the site title at the beginning with a colon as seperator
	SiteTitleColonFront

	// Don't show site title
	SiteTitleNone
)

type srcDst struct{ src, dst string }

type ctxKey string

const (
	ctxKeyTitle ctxKey = "routes' titles"
	ctxKeyError ctxKey = "error"
)

// Get the current router's title. This will panic if this is used outside templ.
func GetTitle(ctx context.Context) string { return ctx.Value(ctxKeyTitle).(string) }

// The main type containing configuration for static files generation.
type Generator struct {
	Head Head
	Body Body

	src      string
	dest     string
	siteName string
	lang     string

	siteTitleOption SiteTitleOption

	ctx    context.Context
	noBase bool
	routes []*Route
	root   *os.Root

	files, dirs []srcDst
}

// Generator configurations.
type GeneratorConfig struct {
	// Don't use the built-in base template (temfest.Base)
	NoBase bool

	// Directory that will be used as root for finding necessary files.
	Source string

	// Directory where the generated files will be.
	Destination string

	// HTML language. Does not do anything if Base sets to none.
	Lang string

	// Styling for the site title.
	SiteTitleOption SiteTitleOption
}

// use nil to use default configs
func NewGenerator(ctx context.Context, siteName string, config *GeneratorConfig) *Generator {
	g := &Generator{}

	if config == nil {
		config = &GeneratorConfig{Destination: "dist"}
	}
	if len(config.Lang) < 1 {
		config.Lang = "en"
	}

	festDest, ok := os.LookupEnv("FEST_DEST")
	g.dest = ternary(ok, festDest, config.Destination)

	festSrc, ok := os.LookupEnv("FEST_SRC")
	g.src = ternary(ok, festSrc, config.Source)

	g.ctx = ctx
	g.siteName = siteName
	g.lang = config.Lang
	g.noBase = config.NoBase
	g.siteTitleOption = config.SiteTitleOption

	return g
}

// Add a single route with the specified path that will generate a file from the component
// relative from the Generator destionation.
func (g *Generator) AddRoute(path string, comp templ.Component) *Route {
	r := g.inspect(path, comp)

	g.routes = append(g.routes, r)
	return r
}

// Add a single route with the specified path that
// will generate a file from the component
// relative from the Generator destionation.
// All errors returned from this function will be handled
// in g.Generate method.
func (g *Generator) AddRouteFunc(
	path string, fn func(context.Context) (templ.Component, error),
) *Route {
	r := g.inspect(path, nil)

	t, err := fn(g.ctx)
	if err != nil {
		err = fmt.Errorf("error while rendering route. err: \"%w\", path: \"%v\" ", err, path)
		g.addError(path, err)
		return nil
	}
	r.comp = t

	g.routes = append(g.routes, r)
	return r
}

// Return the generator context
func (g *Generator) Context() context.Context { return g.ctx }

// Copy src that is a file to the dst inside generated location.
// Relative from Generator Source
func (g *Generator) CopyFile(src, dst string) {
	path := ternary(len(g.src) > 0, g.src, wd)
	g.files = append(g.files, srcDst{
		src: filepath.Join(path, src),
		dst: filepath.Join(dst, filepath.Base(src)),
	})
}

// Copy src that is a directory to dst inside generated location.
// Will also copy the directory instead of the content-only.
// Relative from Generator Source
func (g *Generator) CopyDir(src, dst string) {
	path := ternary(len(g.src) > 0, g.src, wd)
	g.dirs = append(g.dirs, srcDst{
		src: filepath.Join(path, src),
		dst: filepath.Join(dst, filepath.Base(src)),
	})
}

// Generate all the components added to g.
func (g *Generator) Generate() error {
	if err := os.MkdirAll(g.dest, 0744); err != nil {
		return fmt.Errorf("error making dir: %w", err)
	}

	root, err := os.OpenRoot(g.dest)
	if err != nil {
		return fmt.Errorf("error opening root: %w", err)
	}
	g.root = root

	defer g.root.Close()

	val := g.ctx.Value(ctxKeyError)
	if val != nil {
		errs := val.(map[string]error)
		var complete error
		route := RouteError{complete: complete}
		for k, err := range errs {
			complete = fmt.Errorf("%v: %w; ", k, err)
		}
		return fmt.Errorf("%w (%v)", route, complete.Error())
	}

	for _, v := range g.dirs {
		if err := copyDir(v.src, filepath.Join(g.dest, v.dst)); err != nil {
			return fmt.Errorf("error copying \"%v\" to \"%v\": %w", v.src, v.dst, err)
		}
	}

	for _, v := range g.files {
		if err := copyFile(v.src, filepath.Join(g.dest, v.dst)); err != nil {
			return fmt.Errorf("error copying \"%v\" to \"%v\" %w", v.src, v.dst, err)
		}
	}

	for _, r := range g.routes {
		var filename string
		var dir string
		var path string

		if r.isHTMLFile {
			// filename = filepath.Base(dir)
			dir = filepath.Dir(r.path)
			path = r.path
		} else {
			filename = "index.html"
			dir = r.path
			path = filepath.Join(r.path, filename)
		}

		// why no MkdirAll for root?
		if err := os.MkdirAll(filepath.Join(g.dest, dir), 0744); err != nil {
			return fmt.Errorf("error while making parent directory %v: %w", g.dest, err)
		}

		f, err := g.root.Create(path)
		if err != nil {
			return fmt.Errorf("error while creating path: %w", err)
		}
		defer f.Close()

		var title string

		if r.noTitle {
			// let the title be empty
		} else if len(r.title) > 0 {
			switch g.siteTitleOption {
			case SiteTitleDashBack:
				title = r.title + " - " + g.siteName
			case SiteTitleDashFront:
				title = g.siteName + " - " + r.title
			case SiteTitleColonBack:
				title = r.title + ": " + g.siteName
			case SiteTitleColonFront:
				title = g.siteName + ": " + r.title
			case SiteTitleNone:
				title = g.siteName
			default:
				return errors.New("unrecognized SiteName enum")
			}
		} else {
			title = g.siteName
		}

		// override the base
		if r.base != nil {
			r.comp = temfest.Override(r.base, r.comp)
		} else {
			if !g.noBase {
				// prepend the main component
				body := append([]templ.Component{r.comp}, g.Body...)
				r.comp = temfest.Base(title, g.lang, g.Head, body)
			}
		}

		newCtx := context.WithValue(g.ctx, ctxKeyTitle, r.title)

		if err := r.comp.Render(newCtx, f); err != nil {
			return fmt.Errorf("error while rendering: %w", err)
		}
	}
	return nil
}

func (g *Generator) inspect(path string, comp templ.Component) *Route {
	if path[0] == '/' {
		path = string(path[1:])
	}

	r := &Route{path: path, comp: comp}

	if strings.HasSuffix(path, "html") {
		r.isHTMLFile = true
	}

	return r
}

func (g *Generator) addError(path string, err error) {
	val := g.ctx.Value(ctxKeyError)
	if val == nil {
		val = map[string]error{}
	}
	errs := val.(map[string]error)
	errs[path] = err
	g.ctx = context.WithValue(g.ctx, ctxKeyError, errs)
}

type Route struct {
	noTitle bool

	path  string
	comp  templ.Component
	title string

	isHTMLFile bool

	// Only non-nil when overrided
	base templ.Component
}

// Set the Route title. This will use the site's name
// when title is not set, unless SiteName is  set to none.
func (r *Route) SetTitle(title string) *Route {
	r.title = title
	return r
}

// Override the base component. Note that it must have the implemented templ { children... }
func (r *Route) OverrideBase(comp templ.Component) *Route {
	r.base = comp
	return r
}

// Disable the built-in title.
func (r *Route) NoTitle() *Route {
	r.noTitle = true
	return r
}

type RouteError struct {
	complete error
}

func (r RouteError) Unwrap() error { return r.complete }

func (r RouteError) Error() string { return "route error" }

// Components that will be rendered in the html <head> tag.
type Head []templ.Component

// Append templ components to <head> tag.
func (h *Head) Add(comp ...templ.Component) { *h = append(*h, comp...) }

// Components that will be rendered in the html <body> tag. This is intended for use with
// embeed tags such as <script>, <style>, etc. Use AddRoute to add the main component.
type Body []templ.Component

// Append templ components to Body
func (b *Body) Add(comp ...templ.Component) { *b = append(*b, comp...) }
