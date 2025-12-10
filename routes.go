package fest

import (
	"context"
	"errors"
	"fmt"
	"math"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/a-h/templ"
	"github.com/zilllaiss/fest/temfest"
)

// Routes generates multiple routes. User must call AddToGenerator
// in order to add the routes to the Generator.
type Routes[T any] struct {
	HeadBody   HeadBody

	path  string
	title *string
	data  []T

	// configs
	useData bool

	baseConfig *temfest.BaseConfig
}

// NewRoutes creates routes from the data with specified size.
// By default, the slug will be the same as string item
func NewRoutes(path string, slugs []string) *Routes[string] {
	return &Routes[string]{path: path, data: slugs, useData: true}
}

// NewRoutesT creates routes from the data with specified size.
// By default, the slug will be 1-based index of the item.
func NewRoutesT[T any](path string, data []T) *Routes[T] {
	return &Routes[T]{path: path, data: data}
}

// Pagination contains the information of the current page route.
type Pagination[T any] struct {
	// The current page number
	Current int

	// The total number of pages, or the last page.
	Total int

	// The chuck of data sliced with the set size from original data.
	Chunk []T
}

// NewPaginatedRoutes creates routes that will be paginated from data with specified size.
func NewPaginatedRoutes[T any](path string, data []T, size int) *Routes[*Pagination[T]] {
	it := slices.Chunk(data, size)
	r := &Routes[*Pagination[T]]{path: path}
	total := int(math.Ceil(float64(len(data)) / float64(size)))

	page := 1
	for chunk := range it {
		p := &Pagination[T]{Current: page, Total: total, Chunk: chunk}
		r.data = append(r.data, p)
		page++
	}

	return r
}

// AddToGenerator adds routes to the set Generator.
func (rs *Routes[T]) AddToGenerator(g *Generator, fn RouteFunc[T]) {
	if !strings.Contains(rs.path, "{s}") {
		g.addError(rs.path, errors.New("slug not found"))
		return
	}
	for i, d := range rs.data {
		slug := ternary(rs.useData, fmt.Sprintf("%v", rs.data[i]), strconv.Itoa(i+1))
		rp := &RouteParam[T]{item: d, slug: slug}

		comp, err := fn(g.ctx, rp)
		if err != nil {
			g.addError(rs.path, err)
			return
		}
		slug = rp.slug
		if rs.title == nil {
			rs.title = ptr("{s}")
		}

		ext := filepath.Ext(slug)
		if len(ext) > 0 {
			slug = strings.TrimSuffix(slug, ext)
		}

		t := ternary(len(rp.title) > 0, rp.title, *rs.title)

		path := strings.ReplaceAll(rs.path, "{s}", slug)
		title := strings.ReplaceAll(t, "{s}", slug)

		if rs.baseConfig == nil {
			rs.baseConfig = &temfest.BaseConfig{}
		}
		if rp.baseConfig != nil {
			inheritChildValues(rs.baseConfig, rp.baseConfig)
			rs.HeadBody.Head(rp.HeadBody.head...)
			rs.HeadBody.Body(rp.HeadBody.body...)
		}
		r := g.AddRoute(path, comp).SetTitle(title).BaseConfig(*rs.baseConfig)

		r.HeadBody.Head(rs.HeadBody.head...)
		r.HeadBody.Body(rs.HeadBody.body...)
	}
}

// SetTitle sets each item route's title where `{s}` will be replaced with slug.
// If the title is empty, then only the site's name used. The default is "{s}"
func (rs *Routes[T]) SetTitle(title string) *Routes[T] {
	rs.title = &title
	return rs
}

// BaseConfig sets the temfest.Base config for the routes.
// Unset/empty field will be no-op.
func (rs *Routes[T]) BaseConfig(conf temfest.BaseConfig) *Routes[T] {
	rs.baseConfig = &conf
	return rs
}

type RouteFunc[T any] = func(context.Context, *RouteParam[T]) (templ.Component, error)

// RouteParam is the current route parameter.
type RouteParam[T any] struct {
	item  T
	slug  string
	title string

	baseConfig *temfest.BaseConfig
	HeadBody   HeadBody
}

// BaseConfig sets the temfest.Base config for the current route.
// Unset/empty field will be no-op.
func (rp *RouteParam[T]) BaseConfig(conf temfest.BaseConfig) { rp.baseConfig = &conf }

// GetSlug gets the currently set slug.
func (rp *RouteParam[T]) GetSlug() string { return rp.slug }

// SetSlug sets the current slug.
func (rp *RouteParam[T]) SetSlug(slug string) { rp.slug = slug }

// GetItem gets the current data item.
func (rp *RouteParam[T]) GetItem() T { return rp.item }

// SetTitle sets the current route title. It overrides the title sets for Routes.
func (rp *RouteParam[T]) SetTitle(title string) { rp.title = title }
