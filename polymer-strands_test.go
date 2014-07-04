package main

import (
	"bytes"
	"testing"
)

func TestWriteDeps(t *testing.T) {
	var buf bytes.Buffer
	w := &simpleDependencyWriter{&buf}
	err := WriteDeps(w, "testdata/index.html")
	if err != nil {
		t.Fatal(err)
	}
	expected := `testdata/index.html testdata/hello.html
testdata/index.html testdata/heading.html
testdata/hello.html testdata/heading.html
`
	if expected != buf.String() {
		t.Errorf("Got: \"%v\", want: \"%v\"", buf.String(), expected)
	}
}

func TestWriteDepsDot(t *testing.T) {
	var buf bytes.Buffer
	w := &dotDependencyWriter{&buf}
	err := WriteDeps(w, "testdata/index.html")
	if err != nil {
		t.Fatal(err)
	}
	expected := `digraph dependencies {
  "testdata/index.html" -> "testdata/hello.html";
  "testdata/index.html" -> "testdata/heading.html";
  "testdata/hello.html" -> "testdata/heading.html";
}
`
	if expected != buf.String() {
		t.Errorf("Got: \"%v\", want: \"%v\"", buf.String(), expected)
	}
}
