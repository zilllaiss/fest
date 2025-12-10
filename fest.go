// Package fest is the main fest package
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

// SiteNameOption specifies the title options
type SiteNameOption int

const (
	// Show the site title at the end
	SiteNameBack SiteNameOption = iota

	// Show the site title at the beginning
	SiteNameFront

	// Don't show site name in the title
	SiteNameNone
)

type srcDst struct{ src, dst string }

type ctxKey string

const (
	ctxKeyTitle ctxKey = "routes' titles"
	ctxKeyError ctxKey = "error"
)

// GetTitle gets the current router's title.
func GetTitle(ctx context.Context) string { return ctx.Value(ctxKeyTitle).(string) }

// Generator contains configuration for static files generation.
type Generator struct {
	HeadBody HeadBody

	src      string
	dest     string
	siteName string

	baseConfig temfest.BaseConfig

	siteTitleOption SiteNameOption
	seperator       string

	ctx    context.Context
	noBase bool
	routes []*Route
	root   *os.Root

	files, dirs []srcDst
}

// GeneratorConfig is configurations for Generator.
type GeneratorConfig struct {
	// Don't use the built-in base template that is temfest.Base.
	// Note that utilities that modify temfest.Base will be no-op,
	// thus it is recommended to modify Head, Body, and BaseConfig instead.
	NoBase bool

	// Directory that will be used as root directory for finding necessary files.
	// By default it's "."
	Source string

	// Directory where static files will be generated. By default it's "./dist".
	Destination string

	// SiteNameOption is the position in the title where the site's
	// should be rendered. By default it's SiteNameBack.
	SiteNameOption SiteNameOption

	// Seperator is the string used to seperate page title with site's name.
	// By default it's " - ".
	Seperator string

	// BaseConfig is temfest.Base config.
	BaseConfig temfest.BaseConfig
}

// NewGenerator creates a new generator. Use nil to use default configs
func NewGenerator(ctx context.Context, siteName string, config *GeneratorConfig) *Generator {
	g := &Generator{}

	if config == nil {
		config = &GeneratorConfig{Destination: "dist"}
	}
	if len(config.Seperator) < 1 {
		config.Seperator = " - "
	}

	festDest, ok := os.LookupEnv("FEST_DEST")
	g.dest = ternary(ok, festDest, config.Destination)

	festSrc, ok := os.LookupEnv("FEST_SRC")
	g.src = ternary(ok, festSrc, config.Source)

	g.ctx = ctx
	g.siteName = siteName
	g.noBase = config.NoBase
	g.siteTitleOption = config.SiteNameOption
	g.seperator = config.Seperator
	g.baseConfig = config.BaseConfig

	return g
}

// AddRoute adds a single route with the specified path that will generate a file from the component
// relative from the Generator destination.
func (g *Generator) AddRoute(path string, comp templ.Component) *Route {
	r := g.inspect(path, comp)

	g.routes = append(g.routes, r)
	return r
}

// AddRouteFunc adds a single route with the specified path that
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

// Context returns the generator context.
func (g *Generator) Context() context.Context { return g.ctx }

// CopyFile copies src that is a file to the dst inside generated location.
// Relative from Source of g.
func (g *Generator) CopyFile(src, dst string) {
	path := ternary(len(g.src) > 0, g.src, wd)
	g.files = append(g.files, srcDst{
		src: filepath.Join(path, src),
		dst: filepath.Join(dst, filepath.Base(src)),
	})
}

// CopyDir copies src that is a directory to dst inside generated location.
// Will also copy the directory instead of the content-only.
// Relative from Source set of g.
func (g *Generator) CopyDir(src, dst string) {
	path := ternary(len(g.src) > 0, g.src, wd)
	g.dirs = append(g.dirs, srcDst{
		src: filepath.Join(path, src),
		dst: filepath.Join(dst, filepath.Base(src)),
	})
}

// Generate generates all the components added to g.
func (g *Generator) Generate() error {
	if err := os.MkdirAll(g.dest, 0o744); err != nil {
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
		if err := os.MkdirAll(filepath.Join(g.dest, dir), 0o744); err != nil {
			return fmt.Errorf("error while making parent directory %v: %w", g.dest, err)
		}

		f, err := g.root.Create(path)
		if err != nil {
			return fmt.Errorf("error while creating path: %w", err)
		}
		defer f.Close()

		var title, tc string

		if r.title != nil {
			rt := *r.title
			tc = *r.title
			switch g.siteTitleOption {
			case SiteNameBack:
				title = rt + g.seperator + g.siteName
			case SiteNameFront:
				title = g.siteName + g.seperator + rt
			case SiteNameNone:
				title = rt
			default:
				return errors.New("unrecognized SiteNameOption enum")
			}
		} else {
			title = g.siteName
		}

		// override the base
		if r.base != nil {
			r.comp = temfest.Nest(r.base, r.comp)
		} else if !g.noBase {
			if r.baseConfig == nil {
				r.baseConfig = &temfest.BaseConfig{}
			}

			cp := ptr(g.baseConfig)
			inheritChildValues(cp, r.baseConfig)
			head := append(g.HeadBody.head, r.HeadBody.head...)
			body := append(g.HeadBody.body, r.HeadBody.body...)

			r.comp = temfest.Base(title, r.comp, head, body, cp)
		}

		newCtx := context.WithValue(g.ctx, ctxKeyTitle, tc)

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

// Route contains data necessary to generate a route.
type Route struct {
	path  string
	comp  templ.Component
	title *string

	isHTMLFile bool

	// Only non-nil when overrided
	base templ.Component

	baseConfig *temfest.BaseConfig
	HeadBody   HeadBody
}

// SetTitle sets the Route title. By default,
// the route will use site's name.
func (r *Route) SetTitle(title string) *Route {
	r.title = &title
	return r
}

func (r *Route) BaseConfig(conf temfest.BaseConfig) *Route {
	r.baseConfig = &conf
	return r
}

// OverrideBase overrides the base component. Note that it must have the implemented templ { children... }
func (r *Route) OverrideBase(comp templ.Component) *Route {
	r.base = comp
	return r
}

type RouteError struct {
	complete error
}

func (r RouteError) Unwrap() error { return r.complete }

func (r RouteError) Error() string { return "route error" }

// HeadBody represents `<head>` and `<body>` tags.
// It is mainly used to append.
type HeadBody struct{ head, body []templ.Component }

func (hb *HeadBody) Head(comps ...templ.Component) { hb.head = append(hb.head, comps...) }
func (hb *HeadBody) Body(comps ...templ.Component) { hb.body = append(hb.body, comps...) }
