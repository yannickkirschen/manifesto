# Manifesto

[![Lint commit message](https://github.com/yannickkirschen/manifesto/actions/workflows/commit-lint.yml/badge.svg)](https://github.com/yannickkirschen/manifesto/actions/workflows/commit-lint.yml)
[![Push](https://github.com/yannickkirschen/manifesto/actions/workflows/push.yml/badge.svg)](https://github.com/yannickkirschen/manifesto/actions/workflows/push.yml)
[![Release](https://github.com/yannickkirschen/manifesto/actions/workflows/release.yml/badge.svg)](https://github.com/yannickkirschen/manifesto/actions/workflows/release.yml)
[![GitHub release](https://img.shields.io/github/release/yannickkirschen/manifesto.svg)](https://github.com/yannickkirschen/manifesto/releases/)

Manifesto is a library that manages declarative APIs using the listener pattern,
just like Kubernetes with its resources and operators.

> [!WARNING]
> This project is currently under heavy development and not stable.

## Loading Manifests

Let's define an explanatory manifest in `examples/my-manifest-1.yaml`:

```yaml
apiVersion: example.com/v1alpha1
kind: MyManifest

metadata:
  name: my-manifest-1

spec:
  message: hello, world
```

We want to parse this manifest in a structure we have defined as:

```go
type MySpec struct {
    Message string `yaml:"message" json:"message"`
}
```

There are two different ways of parsing, as described as follows.

### Parsing with static type declaration

When parsing, we define what type we expect for the spec and status field:

```go
manifest := manifesto.ParseFile("examples/my-manifest-1.yaml", &MySpec{}, &MySpec{})
// Do something with manifest
```

### Parsing with dynamic type declaration

Before parsing, we register a type for the spec and status field that should be
applied for a API version / Kind combination:

```go
manifesto.RegisterType("example.com/v1alpha1", "MyManifest", MySpec{}, MySpec{})
manifest := manifesto.AutoParseFile("examples/my-manifest-1.yaml")
// Do something with manifest
```

It is a good pattern to define type associations in the `init` function of a
package.

## Using a Manifest Pool

```go
pool := manifesto.CreatePool()
pool.Listen(MyListener)

pool.Apply(manifest)                     // Calls all listeners
pool.ApplyPartial(MyListener, *manifest) // Calls all listeners, except the specified one
pool.ApplySilent(*manifest)              // Does not call any listener at all

key := manifesto.CreateKey()          // Based on apiVersion and kind
theManifest, ok := pool.GetByKey(key) // Get the manifest and check existence

pool.Delete(key) // Gets ignored if the key does not exist
```
