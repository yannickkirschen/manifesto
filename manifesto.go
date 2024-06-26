package manifesto

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"

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
	Errors     []string  `yaml:"errors,omitempty" json:"errors,omitempty"`
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

	// Errors contains errors that occurred while handling the manifest.
	Errors []string `yaml:"errors,omitempty" json:"errors,omitempty"`
}

// CreateKey created a new ManifestKey based on the ApiVersion and Kind.
func (manifest *Manifest) CreateKey() *ManifestKey {
	return &ManifestKey{manifest.ApiVersion, manifest.Kind, manifest.Metadata.Name}
}

// Error adds an error message to the list of errors.
func (manifest *Manifest) Error(message string) {
	manifest.Errors = append(manifest.Errors, message)
}

// ManifestKey is a primary key for manifests.
type ManifestKey struct {
	ApiVersion string
	Kind       string
	Name       string
}

// NewManifestKey creates a new key based on the parameters.
func NewManifestKey(apiVersion string, kind string, name string) *ManifestKey {
	return &ManifestKey{apiVersion, kind, name}
}

// Metadata contains all additional information on a manifest.
type Metadata struct {
	// Name is the name of the manifest. Within a Kind, the name must be unique.
	Name string `yaml:"name" json:"name"`

	// Labels is a key-value dictionary for storing any additional data.
	Labels map[string]string `yaml:"labels" json:"labels"`
}

// Listener is a function that is called when a manifest has been changed.
type Listener func(*Pool, Action, Manifest)

// Pool holds all manifests and listeners.
type Pool struct {
	manifests map[ManifestKey]Manifest
	listeners []Listener
	wg        sync.WaitGroup
}

// CreatePool creates an empty Pool.
func CreatePool() *Pool {
	return &Pool{
		manifests: make(map[ManifestKey]Manifest),
		listeners: make([]Listener, 0),
	}
}

// Listen add a listener to the pool.
func (pool *Pool) Listen(listener Listener) {
	pool.listeners = append(pool.listeners, listener)
}

// Apply adds or updates the manifest to or in the pool and calls all listeners.
// The manifest is transferred as value, not as reference. By doing so, we
// prevent race conditions.
func (pool *Pool) Apply(manifest Manifest) {
	key := manifest.CreateKey()

	if _, ok := pool.manifests[*key]; ok {
		pool.manifests[*key] = manifest
		for _, listener := range pool.listeners {
			pool.apply(listener, Updated, &manifest)
		}
	} else {
		pool.manifests[*key] = manifest
		for _, listener := range pool.listeners {
			pool.apply(listener, Created, &manifest)
		}
	}
}

// ApplyPartial adds or updates the manifest to or in the pool and calls all
// listeners except the specified one. This is meant to be used when a listener
// changes a manifest and should not be called again for that change (that could
// result in an endless loop). The manifest is transferred as value, not as
// reference. By doing so, we prevent race conditions.
func (pool *Pool) ApplyPartial(except Listener, manifest Manifest) {
	key := manifest.CreateKey()
	exceptName := fmt.Sprintf("%v", except)

	if _, ok := pool.manifests[*key]; ok {
		pool.manifests[*key] = manifest
		for _, listener := range pool.listeners {
			if fmt.Sprintf("%v", listener) != exceptName {
				pool.apply(listener, Updated, &manifest)
			}
		}
	} else {
		pool.manifests[*key] = manifest
		for _, listener := range pool.listeners {
			pool.apply(listener, Created, &manifest)
		}
	}
}

// ApplySilent adds or updates the manifest to or in the pool WITHOUT calling
// the listeners. This function is especially useful when using Manifesto just
// as a simple database without listeners.
func (pool *Pool) ApplySilent(manifest Manifest) {
	pool.manifests[*manifest.CreateKey()] = manifest
}

func (pool *Pool) apply(listener Listener, action Action, manifest *Manifest) {
	pool.wg.Add(1)
	go func(listener Listener) {
		defer pool.wg.Done()
		listener(pool, action, *manifest)
	}(listener)
}

// Delete deletes a manifest from the pool.
func (pool *Pool) Delete(key *ManifestKey) {
	if manifest, ok := pool.manifests[*key]; ok {
		delete(pool.manifests, *key)
		for _, listener := range pool.listeners {
			pool.wg.Add(1)
			go func(listener Listener) {
				defer pool.wg.Done()
				listener(pool, Deleted, manifest)
			}(listener)
		}
	}
}

// GetByKey searches for a manifest and returns it.
func (pool *Pool) GetByKey(key *ManifestKey) (Manifest, bool) {
	manifest, ok := pool.manifests[*key]
	return manifest, ok
}

// Find goes through all existing manifests and filters for a testing function.
func (pool *Pool) Find(test func(Manifest) bool) []Manifest {
	manifests := make([]Manifest, 0)
	for _, manifest := range pool.manifests {
		if test(manifest) {
			manifests = append(manifests, manifest)
		}
	}
	return manifests
}

// Waits till all listeners have completed their work.
func (pool *Pool) Wait() {
	pool.wg.Wait()
}

// ParseFile reads a JSON/YAML file and returns the parsed Manifest.
func ParseFile(filename string, spec any, status any) *Manifest {
	content, err := os.ReadFile(filename)
	if err != nil {
		log.Fatal("Error when opening file: ", err)
	}

	return ParseBytes(content, spec, status)
}

// ParseReader parses JSON/YAML data from an io.ReadCloser into a Manifest.
func ParseReader(r io.ReadCloser, spec any, status any) *Manifest {
	var manifest manifest
	err := yaml.NewDecoder(r).Decode(&manifest)

	if err != nil {
		log.Fatal("Error during unmarshal data: ", err)
	}

	return parseManifest(&manifest, spec, status)
}

// ParseString parses JSON/YAML data from a string into a Manifest.
func ParseString(s string, spec any, status any) *Manifest {
	return ParseBytes([]byte(s), spec, status)
}

// ParseBytes parses JSON/YAML data from a byte slice into a Manifest.
func ParseBytes(b []byte, spec any, status any) *Manifest {
	var manifest manifest
	err := yaml.Unmarshal(b, &manifest)

	if err != nil {
		log.Fatal("Error during unmarshal string: ", err)
	}

	return parseManifest(&manifest, spec, status)
}

func parseManifest(manifest *manifest, spec any, status any) *Manifest {
	err := manifest.Spec.Decode(spec)
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
		Errors:     manifest.Errors,
	}
}
