package main

import (
	"fmt"

	"github.com/yannickkirschen/manifesto"
)

type MySpec struct {
	Message string `yaml:"message" json:"message"`
}

func MyListener(action manifesto.Action, manifest *manifesto.Manifest) error {
	if manifest.ApiVersion != "example.com/v1alpha1" || manifest.Kind != "MyManifest" {
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
	m := manifesto.ParseFile("example/my-manifest.yaml", &MySpec{}, &MySpec{})

	pool := manifesto.CreatePool()
	pool.Listen(MyListener)
	pool.Apply(m)

	m3 := pool.GetByKey(m.CreateKey())
	pool.Delete(m3.CreateKey())
}
