package examples

import (
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/zilllaiss/fest"
	"github.com/zilllaiss/fest/examples/basic"
)

var rootPath string

func TestMain(m *testing.M) {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	currentPath := wd
	var found bool

	for range 3 {
		cur := filepath.Join(currentPath, "go.mod")

		gomod, err := os.Open(cur)
		if err != nil {
			currentPath = filepath.Join(currentPath, "..")
			continue
		}
		defer gomod.Close()
		rootPath = currentPath
		found = true
	}
	if !found {
		panic("go.mod not found")
	}

	if code := m.Run(); code != 0 {
		log.Fatalln("test failed")
		return
	}
}

func TestExamples(t *testing.T) {
	basicPath := "../dist-basic"

	g := basic.CreateRoutes(&fest.GeneratorConfig{
		Destination: basicPath,
		Source:      rootPath,
	})

	defer os.RemoveAll(basicPath)

	if err := g.Generate(); err != nil {
		t.Fatal(err)
	}
}
