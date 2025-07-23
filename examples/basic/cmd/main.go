package main

import "github.com/zilllaiss/fest/examples/basic"

func main() {
	// use the default config
	g := basic.CreateRoutes(nil)

	// render all components
	if err := g.Generate(); err != nil {
		panic(err)
	}
}
