package main

import (
	"github.com/zilllaiss/fest/examples/routes"
	"log"
)

func main() {
	g := routes.CreateRoute(nil)
	if err := g.Generate(); err != nil {
		log.Fatalln(err)
	}
}
