package main

import (
	"fmt"

	"github.com/yannickkirschen/manifesto"
)

type MySpec struct {
	Message string `yaml:"message" json:"message"`
}

func init() {
	manifesto.RegisterType("example.com/v1alpha1", "MyManifest", MySpec{}, MySpec{})
}

func main() {
	m1 := manifesto.AutoParseFile("examples/my-manifest-1.yaml")
	fmt.Println(m1.Spec.(*MySpec).Message)

	m2 := manifesto.AutoParseFile("examples/my-manifest-2.yaml")
	fmt.Println(m2.Spec.(*MySpec).Message)
}
