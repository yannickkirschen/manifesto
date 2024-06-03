package main

import (
	"fmt"
	"log"

	"github.com/yannickkirschen/manifesto"
)

type MySpec struct {
	Message string `yaml:"message" json:"message"`
}

func MyListener(action manifesto.Action, manifest *manifesto.Manifest) error {
	if manifest.ApiVersion != "example.com/v1alpha1" || manifest.Kind != "MyManifest" {
		log.Printf("Unknown API Version and kind: %s/%s", manifest.ApiVersion, manifest.Kind)
		return nil
	}

	spec := manifest.Spec.(*MySpec)

	switch action {
	case manifesto.Created:
		fmt.Println("Created:", spec.Message)
	case manifesto.Updated:
		fmt.Println("Updated:", spec.Message)
	case manifesto.Deleted:
		fmt.Println("Deleted:", spec.Message)
	}

	return nil
}

func main() {
	m1 := manifesto.ParseFile("example/my-manifest-1.yaml", &MySpec{}, &MySpec{})

	pool := manifesto.CreatePool()
	pool.Listen(MyListener)
	pool.Apply(m1)

	m3, _ := pool.GetByKey(m1.CreateKey())
	m3.Error("Houston, we have a problem!")
	pool.Delete(m3.CreateKey())
}
