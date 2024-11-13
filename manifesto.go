package manifesto

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"reflect"

	"gopkg.in/yaml.v3"
)

var specTypes map[ResourceKey]reflect.Type = map[ResourceKey]reflect.Type{}
var statusTypes map[ResourceKey]reflect.Type = map[ResourceKey]reflect.Type{}

// RegisterTypes creates a new type association for a given API Version / Kind
// combination.
func RegisterType(apiVersion, kind string, spec, status any) {
	key := NewResourceKey(apiVersion, kind)
	specTypes[*key] = reflect.TypeOf(spec)
	statusTypes[*key] = reflect.TypeOf(status)
}

type internalManifest struct {
	ApiVersion string    `yaml:"apiVersion" json:"apiVersion"`
	Kind       string    `yaml:"kind" json:"kind"`
	Metadata   Metadata  `yaml:"metadata" json:"metadata"`
	Spec       yaml.Node `yaml:"spec" json:"spec"`
	Status     yaml.Node `yaml:"status" json:"status"`
	Errors     []error   `yaml:"errors,omitempty" json:"errors,omitempty"`
}

// CreateResourceKey created a new ResourceKey based on the ApiVersion and Kind.
func (manifest *internalManifest) CreateResourceKey() *ResourceKey {
	return &ResourceKey{manifest.ApiVersion, manifest.Kind}
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
	Errors []error `yaml:"errors,omitempty" json:"errors,omitempty"`
}

// CreateResourceKey created a new ResourceKey based on the ApiVersion and Kind.
func (manifest *Manifest) CreateResourceKey() *ResourceKey {
	return &ResourceKey{manifest.ApiVersion, manifest.Kind}
}

// CreateKey created a new ManifestKey based on the ApiVersion and Kind.
func (manifest *Manifest) CreateKey() *ManifestKey {
	return &ManifestKey{manifest.ApiVersion, manifest.Kind, manifest.Metadata.Name}
}

// Err adds an error to the list of errors.
func (manifest *Manifest) Err(err error) {
	manifest.Errors = append(manifest.Errors, err)
}

// Error adds an error message to the list of errors.
func (manifest *Manifest) Error(message string) {
	manifest.Errors = append(manifest.Errors, errors.New(message))
}

// Errorf adds an error message to the list of errors and formats the string
// withe the given arguments.
func (manifest *Manifest) Errorf(format string, a ...any) {
	manifest.Errors = append(manifest.Errors, fmt.Errorf(format, a...))
}

// ResourceKey identifies a resource type, regardless of its name.
type ResourceKey struct {
	ApiVersion string
	Kind       string
}

// NewResourceKey creates a new key based on the parameters.
func NewResourceKey(apiVersion, kind string) *ResourceKey {
	return &ResourceKey{apiVersion, kind}
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

// ParseFile reads a JSON/YAML file and returns the parsed Manifest.
func ParseFile(filename string, spec any, status any) *Manifest {
	content, err := os.ReadFile(filename)
	if err != nil {
		log.Fatal("Error when opening file: ", err)
	}

	return ParseBytes(content, spec, status)
}

// AutoParseFile reads a JSON/YAML file and returns the parsed Manifest.
// It detects the appropriate spec and status types, once they have been
// registered by calling RegisterType.
func AutoParseFile(filename string) *Manifest {
	content, err := os.ReadFile(filename)
	if err != nil {
		log.Fatal("Error when opening file: ", err)
	}

	return AutoParseBytes(content)
}

// ParseReader parses JSON/YAML data from an io.ReadCloser into a Manifest.
func ParseReader(r io.ReadCloser, spec any, status any) *Manifest {
	var manifest internalManifest
	err := yaml.NewDecoder(r).Decode(&manifest)

	if err != nil {
		log.Fatal("Error during unmarshal data: ", err)
	}

	return parseInternalManifest(&manifest, spec, status)
}

// AutoParseReader parses JSON/YAML data from an io.ReadCloser into a Manifest.
// It detects the appropriate spec and status types, once they have been
// registered by calling RegisterType.
func AutoParseReader(r io.ReadCloser) *Manifest {
	var manifest internalManifest
	err := yaml.NewDecoder(r).Decode(&manifest)

	if err != nil {
		log.Fatal("Error during unmarshal data: ", err)
	}

	return autoParseInternalManifest(&manifest)
}

// ParseString parses JSON/YAML data from a string into a Manifest.
func ParseString(s string, spec any, status any) *Manifest {
	return ParseBytes([]byte(s), spec, status)
}

// AutoParseString parses JSON/YAML data from a string into a Manifest.
// It detects the appropriate spec and status types, once they have been
// registered by calling RegisterType.
func AutoParseString(s string) *Manifest {
	return AutoParseBytes([]byte(s))
}

// ParseBytes parses JSON/YAML data from a byte slice into a Manifest.
func ParseBytes(b []byte, spec any, status any) *Manifest {
	var manifest internalManifest
	err := yaml.Unmarshal(b, &manifest)

	if err != nil {
		log.Fatal("Error during unmarshal string: ", err)
	}

	return parseInternalManifest(&manifest, spec, status)
}

// AutoParseBytes parses JSON/YAML data from a byte slice into a Manifest.
// It detects the appropriate spec and status types, once they have been
// registered by calling RegisterType.
func AutoParseBytes(b []byte) *Manifest {
	var manifest internalManifest
	err := yaml.Unmarshal(b, &manifest)

	if err != nil {
		log.Fatal("Error during unmarshal string: ", err)
	}

	return autoParseInternalManifest(&manifest)
}

func parseInternalManifest(internal *internalManifest, spec any, status any) *Manifest {
	manifest := &Manifest{
		ApiVersion: internal.ApiVersion,
		Kind:       internal.Kind,
		Metadata:   internal.Metadata,
		Errors:     internal.Errors,
	}

	err := parseSpecAndStatus(internal.Spec, internal.Status, spec, status)
	if err != nil {
		manifest.Err(err)
		return manifest
	}

	manifest.Spec = spec
	manifest.Status = status

	return manifest
}

func autoParseInternalManifest(internal *internalManifest) *Manifest {
	manifest := &Manifest{
		ApiVersion: internal.ApiVersion,
		Kind:       internal.Kind,
		Metadata:   internal.Metadata,
		Errors:     internal.Errors,
	}

	key := internal.CreateResourceKey()
	specType, ok := specTypes[*key]
	if !ok {
		manifest.Error("no type has been found to parse spec")
	}

	statusType, ok := statusTypes[*key]
	if !ok {
		manifest.Error("no type has been found to parse status")
	}

	spec := reflect.New(specType).Interface()
	status := reflect.New(statusType).Interface()

	err := parseSpecAndStatus(internal.Spec, internal.Status, spec, status)
	if err != nil {
		manifest.Err(err)
		return manifest
	}

	manifest.Spec = spec
	manifest.Status = status

	return manifest
}

func parseSpecAndStatus(specNode, statusNode yaml.Node, spec, status any) error {
	err := specNode.Decode(spec)
	if err != nil {
		return fmt.Errorf("error decoding spec into type %s: %s", reflect.TypeOf(specNode).Name(), err)
	}

	err = statusNode.Decode(status)
	if err != nil {
		return fmt.Errorf("error decoding status into type %s: %s", reflect.TypeOf(specNode).Name(), err)
	}

	return nil
}
