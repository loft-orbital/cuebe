# cuebe

Kubernetes release manager powered by [CUE](https://cuelang.org/).

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

**parameters**: none

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

### Injection

Although CUE itslef offers a way to [inject value at runtime](https://cuetorials.com/patterns/inject/),
we found it lacking some features.
Especially with secrets injection, it can become hard to maintain and scale, and does not fit well in a GitOps flow.
For that reason we introduced a way to inject values at runtime.

Right now this system is simplistic, it allows to inject one or more files at root level.
Those files can be plain text or encrypted with [sops](https://github.com/mozilla/sops).

In a near future we will probably introduce a new attribute to tackle this feature in
a more smarter, more featureful way.

## Examples

You will find some examples in the [example folder](https://github.com/loft-orbital/cuebe/tree/main/example).

## Roadmap

- [ ] Better injection system
- [ ] Release lifecycle management
