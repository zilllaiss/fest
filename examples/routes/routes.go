package routes

import (
	"context"
	"fmt"
	"strconv"

	"github.com/zilllaiss/fest"
	"github.com/zilllaiss/fest/examples/views"

	"github.com/a-h/templ"
)

func CreateRoute(config *fest.GeneratorConfig) *fest.Generator {
	g := fest.NewGenerator(context.Background(), "Routes", config)

	// title can be added with slug
	fest.NewRoutes("/post/{s}", simpleSlugs).
		SetTitle("{s} post").AddToGenerator(g, simple)

	data := getCustomData()

	// Use custom data. This time, use default title i.e. site's name only
	fest.NewRoutesT("/blogs/{s}", data).
		AddToGenerator(g, custom)

	// Pagination. Since fest use 1-based index for slug by default,
	// we don't need anything else for the page number
	fest.NewPaginatedRoutes("/page/{s}", data, 3).
		SetTitle("page {s}").AddToGenerator(g, pageFn)

	return g
}

var simpleSlugs = []string{
	"first", "second", "third",
	"fourth", "fifth", "sixth",
	"seventh", "eight", "ninth",
}

func simple(ctx context.Context, rp *fest.RoutesParam[string]) (templ.Component, error) {
	// For this example, we sill use a templ.Raw for simplicity.
	// You can get the slug using GetSlug()
	return templ.Raw(fmt.Sprintf("<h1>%v</h1>", rp.GetSlug())), nil
}

type something struct {
	someData string
	number   int
}

func getCustomData() []*something {
	s := []*something{}
	for i := range 9 {
		num := i + 1
		s = append(s, &something{number: num, someData: simpleSlugs[i]})
	}
	return s
}

func custom(ctx context.Context, rp *fest.RoutesParam[*something]) (templ.Component, error) {
	// Get the individual item catched from data.
	item := rp.GetItem()
	// Set the slug. By default, NewRoutesT will use the 1-based indices
	// of the data slice as slugs
	rp.SetSlug(item.someData)

	return views.Article(strconv.Itoa(item.number), item.someData), nil
}

func pageFn(
	ctx context.Context,
	rp *fest.RoutesParam[*fest.Pagination[*something]],
) (templ.Component, error) {
	page := rp.GetItem()
	raw := fmt.Sprintf("<h1>page %d of %d</h1>", page.Current, page.Total)
	group := []templ.Component{templ.Raw(raw)}

	for _, item := range page.Chunk {
		group = append(group, views.Article(strconv.Itoa(item.number), item.someData))
	}

	return templ.Join(group...), nil
}
