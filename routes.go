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

type Routes[T any] struct {
	path    string
	title   string
	data    []T
	useData bool
}

func NewRoutes(path string, slugs []string) *Routes[string] {
	return &Routes[string]{path: path, data: slugs, useData: true}
}

func NewRoutesT[T any](path string, data []T) *Routes[T] {
	return &Routes[T]{path: path, data: data}
}

type Pagination[T any] struct {
	Current, Total int
	Chunk          []T
}

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

func (r *Routes[T]) AddToGenerator(g *Generator, fn RoutesFunc[T]) {
	if !strings.Contains(r.path, "{s}") {
		g.addError(r.path, errors.New("slug not found"))
		return
	}
	for i, d := range r.data {
		slug := ternary(r.useData, fmt.Sprintf("%v", r.data[i]), strconv.Itoa(i+1))
		rp := &RoutesParam[T]{item: d, slug: slug}

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

		path := strings.ReplaceAll(r.path, "{s}", slug)
		title := strings.ReplaceAll(r.title, "{s}", slug)

		g.AddRoute(path, comp).SetTitle(title)
	}
}

func (r *Routes[T]) SetTitle(title string) *Routes[T] {
	r.title = title
	return r
}

type RoutesFunc[T any] = func(context.Context, *RoutesParam[T]) (templ.Component, error)

type RoutesParam[T any] struct {
	item T
	slug string
}

func (rp *RoutesParam[T]) GetSlug() string     { return rp.slug }
func (rp *RoutesParam[T]) SetSlug(slug string) { rp.slug = slug }
func (rp *RoutesParam[T]) GetItem() T          { return rp.item }

