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
	manifest := ParseFile("example/my-manifest-1.yaml", &MySpec{}, &MySpec{})
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

func TestFind(t *testing.T) {
	m1 := ParseFile("example/my-manifest-1.yaml", &MySpec{}, &MySpec{})
	m2 := ParseFile("example/my-manifest-2.yaml", &MySpec{}, &MySpec{})

	pool := CreatePool()
	pool.Apply(*m1)
	pool.Apply(*m2)

	manifests := pool.Find(
		func(m Manifest) bool {
			spec := m.Spec.(*MySpec)
			return m.ApiVersion == "example.com/v1alpha1" && m.Kind == "MyManifest" && strings.Contains(spec.Message, "world")
		})

	if len(manifests) != 2 {
		t.Fatalf("Unable to find manifests: found %d manifests", len(manifests))
	}
}

func TestDelete(t *testing.T) {
	manifest := ParseString(myManifest, &MySpec{}, &MySpec{})
	checkMyManifest(t, manifest, "string")

	pool := CreatePool()
	pool.Apply(*manifest)

	key := manifest.CreateKey()
	_, ok := pool.GetByKey(key)
	if !ok {
		t.Fatalf("Manifest does not exist after insertion")
	}

	pool.Delete(key)
	_, ok = pool.GetByKey(key)
	if ok {
		t.Fatalf("Manifest does exist after deletion")
	}
}

func TestReferences(t *testing.T) {
	m1 := ParseString(myManifest, &MySpec{}, &MySpec{})
	checkMyManifest(t, m1, "string")

	pool := CreatePool()
	pool.Apply(*m1)

	m2, _ := pool.GetByKey(m1.CreateKey())
	m1.Metadata.Name = "new-name"

	if m2.Metadata.Name != "my-manifest-1" {
		t.Fatalf("Manifest name changed")
	}
}
