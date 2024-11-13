package manifesto

import (
	"io"
	"strings"
	"testing"
)

const myManifest = `apiVersion: example.com/v1alpha1
kind: MyManifest
metadata:
  name: my-manifest-1
spec:
  message: hello, world
`

func wantMyManifest(manifest *Manifest) bool {
	spec := manifest.Spec.(*MySpec)
	return manifest.Metadata.Name == "my-manifest-1" && spec.Message == "hello, world"
}

func checkMyManifest(t *testing.T, manifest *Manifest, from string) {
	if !wantMyManifest(manifest) {
		t.Fatalf("Unable to parse manifest from %s", from)
	}
}

type MySpec struct {
	Message string `yaml:"message" json:"message"`
}

func TestParseFile(t *testing.T) {
	manifest := ParseFile("examples/my-manifest-1.yaml", &MySpec{}, &MySpec{})
	checkMyManifest(t, manifest, "file")
}

func TestParseReader(t *testing.T) {
	r := io.NopCloser(strings.NewReader(myManifest))
	manifest := ParseReader(r, &MySpec{}, &MySpec{})
	checkMyManifest(t, manifest, "reader")
}

func TestParseString(t *testing.T) {
	manifest := ParseString(myManifest, &MySpec{}, &MySpec{})
	checkMyManifest(t, manifest, "string")
}

func TestParseBytes(t *testing.T) {
	b := []byte(myManifest)
	manifest := ParseBytes(b, &MySpec{}, &MySpec{})
	checkMyManifest(t, manifest, "bytes")
}

func TestRegisterTypes(t *testing.T) {
	RegisterType("example.com/v1alpha1", "MyManifest", MySpec{}, MySpec{})
	manifest := ParseFile("examples/my-manifest-1.yaml", &MySpec{}, &MySpec{})
	checkMyManifest(t, manifest, "register types")
}
