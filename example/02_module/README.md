# module

This example leverage CUE modules to better organize the release.
It will still deploy a `Namespace` and a `ConfigMap`.

## Requirements

- Access to a Kubernetes cluster

## Running the example

To export YAML version of this release, run

```shell
cuebe export -e app
```

To deploy this release in your current kubernetes context, run

```shell
cuebe apply -e app
```
