package manifesto

import (
	"strings"
	"testing"
)

func TestFind(t *testing.T) {
	m1 := ParseFile("examples/my-manifest-1.yaml", &MySpec{}, &MySpec{})
	m2 := ParseFile("examples/my-manifest-2.yaml", &MySpec{}, &MySpec{})

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
