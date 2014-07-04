package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"code.google.com/p/go.net/html"
)

var (
	formatAsDot = flag.Bool("dot", false, "Print out in dot format")
)

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <filename>\n", os.Args[0])
		flag.PrintDefaults()
	}
}

// DependencyWriter is the interface for how to write out dependencies.
//
// Start() should be called first, followed by `n` numbers of calls to
// Write(). Finally, End() should be called last.
type DependencyWriter interface {
	Start()
	Write(name, dep string)
	End()
}

// simpleDependencyWriter writes out dependencies as a list of pairs.
type simpleDependencyWriter struct {
	w io.Writer
}

func (w *simpleDependencyWriter) Start()                 {}
func (w *simpleDependencyWriter) Write(name, dep string) { fmt.Fprintln(w.w, name, dep) }
func (w *simpleDependencyWriter) End()                   {}

// dotDependencyWriter writes out dependencies in a dot-file format.
type dotDependencyWriter struct {
	w io.Writer
}

func (w *dotDependencyWriter) Start() {
	fmt.Fprintln(w.w, "digraph dependencies {")
}
func (w *dotDependencyWriter) Write(name, dep string) {
	fmt.Fprintf(w.w, "  \"%s\" -> \"%s\";\n", name, dep)
}
func (w *dotDependencyWriter) End() {
	fmt.Fprintln(w.w, "}")
}

// WriteDeps writes out all of the dependencies to the dependency writer for
// the supplied filename.
func WriteDeps(w DependencyWriter, name string) error {
	seen := make(map[string]struct{})
	w.Start()
	if err := writeDeps(w, name, seen); err != nil {
		return err
	}
	w.End()
	return nil
}

func writeDeps(w DependencyWriter, name string, seen map[string]struct{}) error {
	if _, ok := seen[name]; ok {
		return nil
	}
	seen[name] = struct{}{}
	deps, err := dependencies(name)
	if err != nil {
		return err
	}
	for _, dep := range deps {
		w.Write(name, dep)
	}
	for _, dep := range deps {
		if err := writeDeps(w, dep, seen); err != nil {
			return err
		}
	}
	return nil
}

func dependencies(name string) ([]string, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	includes := make([]string, 0)
	z := html.NewTokenizer(f)
	for {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			if z.Err() == io.EOF {
				return includes, nil
			}
			return nil, z.Err()
		case html.StartTagToken, html.SelfClosingTagToken:
			tagname, hasAttr := z.TagName()
			if bytes.Equal(tagname, []byte("link")) {
				var k, v []byte
				var isImport bool
				var href string
				for hasAttr {
					k, v, hasAttr = z.TagAttr()
					if bytes.Equal(k, []byte("rel")) && bytes.Equal(v, []byte("import")) {
						isImport = true
					} else if bytes.Equal(k, []byte("href")) {
						href = string(v)
					}
				}
				// TODO(pope): Add error handling if we have an import but no path.
				if isImport && href != "" {
					u, err := url.Parse(href)
					if err != nil {
						return nil, err
					}
					// Ignore absolute paths, as those don't get vulcanized.
					if u.IsAbs() || strings.HasPrefix(href, "/") {
						break
					}
					includes = append(includes, filepath.Join(path.Dir(name), href))
				}
			}
		}
	}
}

func main() {
	flag.Parse()
	if len(flag.Args()) != 1 {
		flag.Usage()
		os.Exit(1)
	}
	name := flag.Arg(0)

	var w DependencyWriter
	if *formatAsDot {
		w = &dotDependencyWriter{os.Stdout}
	} else {
		w = &simpleDependencyWriter{os.Stdout}
	}
	if err := WriteDeps(w, name); err != nil {
		log.Fatal(err)
	}
}
