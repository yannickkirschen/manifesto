package manifesto

import (
	"io"
	"strings"
	"testing"
)

type MySpec struct {
	Message string `yaml:"message" json:"message"`
}

func TestParseFile(t *testing.T) {
	manifest := ParseFile("example/my-manifest.yaml", &MySpec{}, &MySpec{})
	spec := manifest.Spec.(*MySpec)

	want := manifest.Metadata.Name == "my-manifest" && spec.Message == "hello, world"

	if !want {
		t.Fatalf("Unable to parse manifest from file")
	}
}

func TestParseReader(t *testing.T) {
	r := io.NopCloser(strings.NewReader(`apiVersion: example.com/v1alpha1
kind: MyManifest
metadata:
  name: my-manifest
spec:
  message: hello, world
    `))
	manifest := ParseReader(r, &MySpec{}, &MySpec{})

	spec := manifest.Spec.(*MySpec)

	want := manifest.Metadata.Name == "my-manifest" && spec.Message == "hello, world"

	if !want {
		t.Fatalf("Unable to parse manifest from reader")
	}
}

func TestParseString(t *testing.T) {
	s := `apiVersion: example.com/v1alpha1
kind: MyManifest
metadata:
  name: my-manifest
spec:
  message: hello, world
    `
	manifest := ParseString(s, &MySpec{}, &MySpec{})

	spec := manifest.Spec.(*MySpec)

	want := manifest.Metadata.Name == "my-manifest" && spec.Message == "hello, world"

	if !want {
		t.Fatalf("Unable to parse manifest from string")
	}
}

func TestParseBytes(t *testing.T) {
	b := []byte(`apiVersion: example.com/v1alpha1
kind: MyManifest
metadata:
  name: my-manifest
spec:
  message: hello, world
    `)
	manifest := ParseBytes(b, &MySpec{}, &MySpec{})

	spec := manifest.Spec.(*MySpec)

	want := manifest.Metadata.Name == "my-manifest" && spec.Message == "hello, world"

	if !want {
		t.Fatalf("Unable to parse manifest from bytes")
	}
}
