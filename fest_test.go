package fest

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/zilllaiss/fest/internal/testfest"
	"github.com/zilllaiss/fest/temfest"
)

func TestGenerator(t *testing.T) {
	// should be recreated when calling Generate
	testPath := filepath.Join("tmp", "gen")

	_ = os.Remove("dist")
	_ = os.Remove(testPath)

	type testData struct {
		name    string
		wantErr bool
		config  *GeneratorConfig
		genPath string
		preRun  func() error
		postRun func() error
	}
	data := []testData{
		{
			name:    "minimum",
			config:  nil,
			genPath: "dist",
		},
		{
			name: "custom dest",
			config: &GeneratorConfig{
				Destination: filepath.Join(testPath, "customdest"),
			},
		},
		{
			name: "complete",
			config: &GeneratorConfig{
				Destination:    filepath.Join(testPath, "complete"),
				Source:         "example",
				SiteNameOption: SiteNameFront,
				Seperator:      ": ",
				BaseConfig: temfest.BaseConfig{
					Lang:       "id",
					CharSet:    "utf-8",
					NoViewport: true,
				},
			},
		},
		{
			name:   "envs",
			config: nil,
			preRun: func() error {
				err1 := os.Setenv("FEST_SRC", "tmp")
				err2 := os.Setenv("FEST_DEST", filepath.Join(testPath, "envs"))
				return cmp.Or(err1, err2)
			},
			postRun: func() error {
				err1 := os.Unsetenv("FEST_SRC")
				err2 := os.Unsetenv("FEST_DEST")
				return cmp.Or(err1, err2)
			},
		},
	}

	testFn := func(v testData) error {
		if v.preRun != nil {
			v.preRun()
		}
		if v.postRun != nil {
			defer v.postRun()
		}
		if v.config != nil && len(v.config.Destination) < 1 {
			if err := os.Setenv("FEST_DEST", testPath); err != nil {
				return err
			}
			defer os.Unsetenv("FEST_DEST")
		}
		g := NewGenerator(context.Background(), v.name, v.config)

		// path and comp is not important in this test
		g.AddRoute("/", testfest.Simple())

		if err := g.Generate(); err != nil {
			return err
		}

		var path, lang string
		charset := "UTF-8"

		if v.config != nil {
			if len(v.config.BaseConfig.CharSet) > 0 {
				charset = v.config.BaseConfig.CharSet
			}
			if len(v.config.BaseConfig.Lang) > 0 {
				lang = v.config.BaseConfig.Lang
			}
		} 

		if len(v.genPath) > 0 {
			path = v.genPath
		} else if v.config != nil && len(v.config.Destination) > 0 {
			path = v.config.Destination
		} else {
			path = testPath
		}

		p := filepath.Join(path, "index.html")
		f, err := os.Open(p)
		if err != nil {
			return err
		}
		defer f.Close()

		doc, err := goquery.NewDocumentFromReader(f)
		if err != nil {
			return err
		}

		var ok bool
		doc.Find("meta[name]").Each(func(i int, s *goquery.Selection) {
			val, ok := s.Attr("name")
			if ok || val == "viewport" {
				ok = true
				return
			}
		})
		if v.config != nil && v.config.BaseConfig.NoViewport && ok {
			return errors.New("meta[name=viewport] exists")
		}

		err = checkElAttr(nil, doc, `meta[charset]`, "charset", charset, true)
		err = checkElAttr(err, doc, `html`, "lang", lang, lang != "")
		if err != nil {
			return err
		}
		return nil
	}

	for _, v := range data {
		t.Run(v.name, func(t *testing.T) {
			if err := testFn(v); err != nil {
				t.Error(err)
			}
		})
	}
}

func checkElAttr(err error, doc *goquery.Document, selector, attr, val string, attrfound bool) error {
	if err != nil {
		return err
	}
	el := doc.Find(selector)
	if el.Length() < 1 {
		return fmt.Errorf("el %v does not exist", selector)
	}
	v, ok := el.Attr(attr)
	if (attrfound && !ok) || (!attrfound && ok) || val != v {
		return fmt.Errorf("%v: expected %t or val==%v, found %t, %v", selector, attrfound, val, ok, v)
	}
	return nil
}
