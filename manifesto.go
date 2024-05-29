package manifesto

import (
	"errors"
	"log"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Action tells what happened to a manifest when it is passed to listeners.
type Action int8

const (
	Created Action = iota
	Updated
	Deleted
)

type manifest struct {
	ApiVersion string    `yaml:"apiVersion" json:"apiVersion"`
	Kind       string    `yaml:"kind" json:"kind"`
	Metadata   Metadata  `yaml:"metadata" json:"metadata"`
	Spec       yaml.Node `yaml:"spec" json:"spec"`
	Status     yaml.Node `yaml:"status" json:"status"`
}

// Manifest is the root entity.
type Manifest struct {
	// ApiVersion defines the version of the API to determine the schema of the
	// Spec and Status field. Usually it is something like foo.example.com/v1.
	ApiVersion string `yaml:"apiVersion" json:"apiVersion"`

	// Kind is the kind of object represented here. Together with the ApiVersion
	// it defines the schema of the Spec and Status.
	Kind string `yaml:"kind" json:"kind"`

	// Metadata holds all metadata information.
	Metadata Metadata `yaml:"metadata" json:"metadata"`

	// Spec holds the actual data. Developers must provide their own struct to
	// be used as Spec.
	Spec any `yaml:"spec" json:"spec"`

	// Status holds status information. Developers must provide their own struct
	// to be used as Status.
	Status any `yaml:"status" json:"status"`
}

// CreateKey created a new ManifestKey based on the ApiVersion and Kind.
func (manifest *Manifest) CreateKey() ManifestKey {
	return ManifestKey{manifest.ApiVersion, manifest.Kind, manifest.Metadata.Name}
}

// ManifestKey is a primary key for manifests.
type ManifestKey struct {
	ApiVersion string
	Kind       string
	Name       string
}

// Metadata contains all additional information on a manifest.
type Metadata struct {
	// Name is the name of the manifest. Within a Kind, the name must be unique.
	Name string `yaml:"name" json:"name"`

	// Labels is a key-value dictionary for storing any additional data.
	Labels map[string]string
}

// Listener is a function that is called when a manifest has been changed.
type Listener func(Action, *Manifest)

// Pool holds all manifests and listeners.
type Pool struct {
	manifests map[ManifestKey]*Manifest
	listeners []Listener
}

// CreatePool creates an empty Pool.
func CreatePool() *Pool {
	return &Pool{make(map[ManifestKey]*Manifest), make([]Listener, 0)}
}

// Listen add a listener to the pool.
func (pool *Pool) Listen(listener Listener) {
	pool.listeners = append(pool.listeners, listener)
}

// Apply add or updates the manifest to or in the pool and calls all listeners.
func (pool *Pool) Apply(manifest *Manifest) {
	key := manifest.CreateKey()

	if _, ok := pool.manifests[key]; ok {
		for _, listener := range pool.listeners {
			pool.manifests[key] = manifest
			listener(Updated, manifest)
		}
	} else {
		for _, listener := range pool.listeners {
			pool.manifests[key] = manifest
			listener(Created, manifest)
		}
	}
}

// Delete deletes a manifest from the pool.
func (pool *Pool) Delete(key ManifestKey) {
	if manifest, ok := pool.manifests[key]; ok {
		delete(pool.manifests, key)
		for _, listener := range pool.listeners {
			listener(Deleted, manifest)
		}
	}
}

// GetByKey searches for a manifest and returns it.
func (pool *Pool) GetByKey(key ManifestKey) *Manifest {
	return pool.manifests[key]
}

// ParseFile reads a JSON/YAML file and returns the parsed Manifest.
func ParseFile(filename string, spec any, status any) *Manifest {
	content, err := os.ReadFile(filename)
	if err != nil {
		log.Fatal("Error when opening file: ", err)
	}

	var manifest manifest
	if strings.HasSuffix(filename, ".json") || strings.HasSuffix(filename, ".yaml") || strings.HasSuffix(filename, ".yml") {
		err = yaml.Unmarshal(content, &manifest)
	} else {
		err = errors.New("unknown file type (must be .json, .yaml, .yml): " + filename)
	}

	if err != nil {
		log.Fatal("Error during unmarshal file: ", err)
	}

	err = manifest.Spec.Decode(spec)
	if err != nil {
		log.Fatal("Error during unmarshal spec: ", err)
	}

	err = manifest.Status.Decode(status)
	if err != nil {
		log.Fatal("Error during unmarshal status: ", err)
	}

	return &Manifest{
		ApiVersion: manifest.ApiVersion,
		Kind:       manifest.Kind,
		Metadata:   manifest.Metadata,
		Spec:       spec,
		Status:     status,
	}
}
