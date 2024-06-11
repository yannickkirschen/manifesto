# Manifesto

[![Lint commit message](https://github.com/yannickkirschen/manifesto/actions/workflows/commit-lint.yml/badge.svg)](https://github.com/yannickkirschen/manifesto/actions/workflows/commit-lint.yml)
[![Push](https://github.com/yannickkirschen/manifesto/actions/workflows/push.yml/badge.svg)](https://github.com/yannickkirschen/manifesto/actions/workflows/push.yml)
[![Release](https://github.com/yannickkirschen/manifesto/actions/workflows/release.yml/badge.svg)](https://github.com/yannickkirschen/manifesto/actions/workflows/release.yml)
[![GitHub release](https://img.shields.io/github/release/yannickkirschen/manifesto.svg)](https://github.com/yannickkirschen/manifesto/releases/)

Manifesto is a library that manages declarative APIs using the listener pattern,
just like Kubernetes with its resources and operators.

> [!WARNING]
> This project is currently under heavy development and not stable.

## Example usage

```yaml
# example/my-manifest.yaml

apiVersion: example.com/v1alpha1
kind: MyManifest

metadata:
  name: my-manifest-1

spec:
  message: hello, world
```

```go
import "github.com/yannickkirschen/manifesto"

type MySpec struct {
    Message string `yaml:"message" json:"message"`
}

func MyListener(pool *manifesto.Pool, action manifesto.Action, manifest manifesto.Manifest) {
    // Do something with manifesto
}

func main() {
    manifest := manifesto.ParseFile("example/my-manifest-1.yaml", &MySpec{}, &MySpec{})
    spec :=  manifest.Spec.(*MySpec) // Do something with spec

    pool := manifesto.CreatePool()
    pool.Listen(MyListener)

    pool.Apply(manifest)                     // Calls all listeners
    pool.ApplyPartial(MyListener, *manifest) // Calls all listeners, except the specified one
    pool.ApplySilent(*manifest)              // Does not call any listener at all

    key := manifesto.CreateKey()          // Based on apiVersion and kind
    theManifest, ok := pool.GetByKey(key) // Get the manifest and check existence

    pool.Delete(key) // Gets ignored if the key does not exist
}
```
