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
)

// Routes generates multiple routes. User must call AddToGenerator
// in order to add the routes to the Generator.
type Routes[T any] struct {
	path    string
	title   string
	data    []T
	useData bool
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
func (r *Routes[T]) AddToGenerator(g *Generator, fn RouteFunc[T]) {
	if !strings.Contains(r.path, "{s}") {
		g.addError(r.path, errors.New("slug not found"))
		return
	}
	for i, d := range r.data {
		slug := ternary(r.useData, fmt.Sprintf("%v", r.data[i]), strconv.Itoa(i+1))
		rp := &RouteParam[T]{item: d, slug: slug}

		comp, err := fn(g.ctx, rp)
		if err != nil {
			g.addError(r.path, err)
			return
		}
		slug = rp.slug

		ext := filepath.Ext(slug)
		if len(ext) > 0 {
			slug = strings.TrimSuffix(slug, ext)
		}

		t := ternary(len(rp.title) > 0, rp.title, r.title)

		path := strings.ReplaceAll(r.path, "{s}", slug)
		title := strings.ReplaceAll(t, "{s}", slug)

		g.AddRoute(path, comp).SetTitle(title)
	}
}

// SetTitle sets each item route's title where `{s}` will be replaced with slug.
func (r *Routes[T]) SetTitle(title string) *Routes[T] {
	r.title = title
	return r
}

type RouteFunc[T any] = func(context.Context, *RouteParam[T]) (templ.Component, error)

// RouteParam is the current route parameter.
type RouteParam[T any] struct {
	item  T
	slug  string
	title string
}

// GetSlug gets the currently set slug.
func (rp *RouteParam[T]) GetSlug() string { return rp.slug }

// SetSlug sets the current slug.
func (rp *RouteParam[T]) SetSlug(slug string) { rp.slug = slug }

// GetItem gets the current data item.
func (rp *RouteParam[T]) GetItem() T { return rp.item }

// SetTitle sets the current route title. It overrides the title sets for Routes.
func (rp *RouteParam[T]) SetTitle(title string) { rp.title = title }
