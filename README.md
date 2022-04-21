# Cuebe

Kubernetes release manager powered by [CUE](https://cuelang.org/).

> 🚧 Disclaimer:
> Cuebe is in beta version
> It is actively maintained but not yet ready to be used in production.

## Why?

Writting your Kubernetes manifests in CUE can drastically increase maintainability
and reliability of your deployment workflow.
Things that was previously reserved to source code
can now be used for the whole lifecycle of an application.
Things like strong typing, reducing boilerplate, unit tests, etc...

Cuebe helps you deploying such resources.
It is the glue between the configuration and the k8s cluster.
Akin what can kubectl do with YAML and JSON.

## How?

Cuebe builds a [Context](#context) and collects every Kubernetes manifests in it.
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

### Manifest

A Manifest is a specification of a Kubernetes object in CUE format.
A Manifest specifies the desired state of an object that Kubernetes will maintain when you apply it.
A CUE value is considered as a Manifest if it has both string properties `apiVersion` and `kind`.

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

Here both `namespace` and `nested.configmap` are considered as Manifest by Cuebe.
Whereas foo just serves during build phase.
Every Manifest must eventually resolves to a JSON serializable object (concrete only values).
Other values can be non-concrete, as Cuebe will not try to render them in a concrete format (YAML, JSON).

Since Cuebe can collect any manifests in your Build, it's up to you to define boundaries.
You can use it to deploy single Manifest, a bunch of unrelated Manifests or use the [Instance](#instance) concept.

### Instance

An instance is a group of [Manifests](#manifest) belonging to the same _application_.
It's a way to group multiple manifests under the same lifecycle.
If you're used to Helm, you can see a Cuebe Instance as a [Helm Release](https://helm.sh/docs/intro/using_helm/#three-big-concepts).

A Manifest needs the `"cuebe.loft-orbital.com/instance": <instance-name>` label (in `metadata.labels`) to be a member of the _<instance-name>_ Instance.
It's up to you to set this label on your Manifests.

When deploying an Instance for the first time, Cuebe will create a Instance object.
It's a cluster-scoped custom resource which will keep track of Instance members.
It means an Instance must be unique to a cluster.
You cannot have multiple Instances with the same name in a single cluster.

When deleting an Instance, even outside of Cuebe (e.g. with `kubectl`) it will automatically deletes subresources.
The way those resources are deleted can be managed with the `"cuebe.loft-orbital.com/deletion-policy"` annotation.
When this annotation is set to `abandon`, the object will not be actually deleted, but its link to the instance removed (we call that an orphan Manifest).
When this annotations is not set, the object will be normally deleted.

### Build

A Build is the action of building a [Context](#context) to [Manifests](#manifest), grouping them into [Instances](#instance) when required.
With `cuebe` cli you can _apply_ or _export_ a Build.

You can tweak the Build phase by using a special set of CUE attributes.
CUE introduced the concept of [attibutes](https://cuelang.org/docs/references/spec/#attributes)
to associate metadata information with value.
Cuebe leverage this mechanism to enhance your Build.
In addition to native attibutes, Cuebe introduces the following (exhaustive) list:

#### @ignore

The `@ignore` attribute marks a value to be ignored by Cuebe.
It means Cuebe will not dive in this value.
Cuebe will hence ignore any Manifest under this value.
This can speed up the build process, avoiding unnecessary recursions.

##### Syntax

```cue
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
it lacks some features.
Especially with secrets injection where it can become hard to maintain and scale, and does not fit well in a GitOps flow.
For that reason Cuebe introduces a way to inject values at runtime.

The `@inject` attribute allows surgical injection of external values in your Cuebe release.
We recommend using this attribute with parsimony, as `cue` itself will ignore it, making your Build hard to debug outside Cuebe.
One of our current use case is to inject sops encrypted values in the Build.
It allow us to keep a GitOps flow (no runtime config, everything commited) without leaking secrets.

Cuebe only supports local file injection as for now.

##### Syntax

```cue
@inject(type=<type>, src=<src> [,path=<path>])
```

- **type**: Injection type. Currerntly only supports `file`

- **src**: Injection source.
For file injection, the path has to be relative to the Build [Context](#context).
Supports cue, json or yaml plain or [sops-enccrypted](https://github.com/mozilla/sops) structured format,
or any text file format when injecting unstructured (c.f. path).

- **path**: [Optional] Path to extract the value from.
For file injection, when the path is not provided Cuebe treats the file as unstructured
and does a plain text injection.

##### Example

_injection.yaml_

```yaml
namespace:
  name: potato
```

_plaintext.md_

```md
# Best sauces

Ketchup Mayo
```

_main.cue_

```cue
namespace: {
  apiVersion: "v1"
  kind:       "Namespace"
  metadata: {
    name: string @inject(type=file, src=injection.yaml, path=$.namespace.name)
  }
}

configmap: {
	apiVersion: "v1"
	kind:       "ConfigMap"

	metadata: name: "sauces"

	data: {
		"README.md": string @inject(src=plaintext.md, type=file)
	}
}
```

### Context

A Context is basically a filesystem that Cuebe uses to Build manifests and instances.
Currently Cuebe only supports local contexts (single file or directory).
Archives (`tar.gz`) and remote Contexts (object storage, https endpoint, etc..) are in the pipe.

When sending multiple Contexts to Cuebe, they will be merged before build.
Think `rsync -a /ContextA/ /ContextB/`.

With Cuebe cli you can _pack_ a Context to upload it and reuse it during _apply_ or _export_ as soon as packs context (so-called cube) are supported.

## Examples

You will find some examples in the [example folder](https://github.com/loft-orbital/cuebe/tree/main/example).

## Roadmap

- [x] Better injection system
- [x] Release lifecycle management
- [ ] Remote Contexts
- [ ] Remote Injection
- [ ] More examples
