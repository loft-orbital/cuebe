# cuebe

Kubernetes release manager powered by [CUE](https://cuelang.org/).

> 🚧 Disclaimer:
> This is an alpha version of cuebe.
> cuebe is not yet ready for production environments.

## Why?

Writting your Kubernetes manifests in CUE can drastically increase maintainability
and reliability of your deployment workflow.
Things that was previously reserved to source code
can now be used for the whole lifecycle of an application.
Things like strong typing, reducing boilerplate, unit tests, etc...

cuebe helps you deploying such resources.
It is the glue between the configuration and the k8s cluster.
Akin what can kubectl do with YAML and JSON.

## How?

cuebe loads a CUE instance and collects every Kubernetes manifests in it.
It then uses [server-side apply](https://kubernetes.io/docs/reference/using-api/server-side-apply/)
to deploy those resources.

## Install

### Release builds

Download the [latest release](https://github.com/loft-orbital/cuebe/releases/latest) from GitHub.

### Install from Source

You will need [Go 1.17](https://go.dev/doc/install) or later installed.

```shell
go install github.com/loft-orbital/cuebe/cmd/cuebe@latest
```

This will install `cuebe` in your `$GOBIN` folder.

## Concepts

### Release

A release is a collection of K8s manifests that together form a bundle of resources to deploy
on a cluster. It takes form of CUE instance(s) containing one or more K8s manifests.
A CUE value is considered a K8s manifests if it has both string properties `apiVersion` and `kind`.

```cue
foo: "bar"
namespace: {
  apiVersion: "v1"
  kind:       "Namespace"
}
nested: configmap: {
  apiVersion: "v1"
  kind:       "ConfigMap"
}
```

Here both `namespace` and `nested.configmap` are considered as K8s manifest by cuebe.
Whereas foo just serves during the build.
Every manifests in a release must eventually resolves to concrete value.
Other values can be non-concrete, as `cuebe` will not try to render them in a concrete format (YAML, JSON).

Since `cuebe` can collect any manifests in your release, it's up to you to define boundaries.
You can see a release as a single K8s manifest,
as a whole application (akin a [Helm](https://helm.sh) chart or a [Kustomization](https://kustomize.io/)),
or even as a superset of multiple application (like [Helmfile](https://github.com/roboll/helmfile) or [Helmsman](https://github.com/Praqma/helmsman)).
To help you start we designed some example on how we use `cuebe` ourself.
If you find a new way of using `cuebe` please share it if possible.
We're still learning and building a best practices guide.

### Attributes

CUE introduced the concept of [attibutes](https://cuelang.org/docs/references/spec/#attributes)
to associate metadata information with value.
Cuebe leverage this mechanism to enhance your release.
In addition to native attibutes, `cuebe` introduces the following (exhaustive) list:

#### @ignore

The `@ignore` attributes mark a value to be ignored by cuebe.
It means cuebe will not dive in this value.
cuebe will hence ignore every manifest under this value.
This can speed up the build process, avoiding unnecessary process.

##### Syntax

```
@ignore()
```

##### Example

```cue
foo: "bar"
namespace: {
  apiVersion: "v1"
  kind:       "Namespace"
}
nested: {
  // This manifest will be ignored
  configmap: {
    apiVersion: "v1"
    kind:       "ConfigMap"
  }
} @ignore()
```

### @inject

Although CUE itslef offers a way to [inject value at runtime](https://cuetorials.com/patterns/inject/),
we found it lacking some features.
Especially with secrets injection, it can become hard to maintain and scale, and does not fit well in a GitOps flow.
For that reason we introduced a way to inject values at runtime.

The `@inject` attribute allows surgical injection of external values in your Cuebe release.
We recommend using this attribute with parsimony, as `cue` itself will ignore it.
One of our current usecase is to inject sops encrypted value in our release.
It allow us to keep a GitOps flow (no runtime config, everything commited) without leaking secrets.

Cuebe only supports local file injection as for now.

##### Syntax

```cue
@inject(type=<type>, src=<src> [,path=<path>])
```

**type**: Injection type. Currerntly only supports `file`
**src**: Injection source.
For file injection, the relative path to the file to inject. Supports cue, json or yaml plain or [sops-enccrypted](https://github.com/mozilla/sops) files.
**path**: [Optional] Path to extract the value from. Default to root.

##### Example

```txtar
-- injection.yaml --
namespace:
  name: potato
-- main.cue --
namespace: {
  apiVersion: "v1"
  kind:       "Namespace"
  metadata: {
    name: string @inject(type=file, src=injection.yaml, path=$.namespace.name)
  }
}
```

## Examples

You will find some examples in the [example folder](https://github.com/loft-orbital/cuebe/tree/main/example).

## Roadmap

- [ ] Better injection system
- [ ] Release lifecycle management
