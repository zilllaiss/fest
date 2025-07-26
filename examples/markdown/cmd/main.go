package main

import (
	"errors"
	"log"

	"github.com/zilllaiss/fest"
	"github.com/zilllaiss/fest/examples/markdown"
)

func main() {
	if err := markdown.UseMarkdown(nil); err != nil {
		re := &fest.RouteError{}
		if errors.As(err, re) {
			log.Println(re.Unwrap())

		}
		log.Fatalln(err)
	}
}
