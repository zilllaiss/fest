package examples

import (
	"fest"
	"fest/examples/basic"
	"os"
	"testing"
)

func TestExamples(t *testing.T) {
	basicPath := "../dist-basic"

	g := basic.CreateRoutes(&fest.GeneratorConfig{
		Destination: basicPath,
	})

	defer os.RemoveAll(basicPath)

	if err := g.Generate(); err != nil {
		t.Fatal(err)
	}
}
