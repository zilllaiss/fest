package examples

import (
	"os"
	"testing"

	"github.com/zilllaiss/fest"
	"github.com/zilllaiss/fest/examples/basic"
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
