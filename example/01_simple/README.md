# Simple

This example describe a super simple configuration to be deployed on Kubernetes.
It contains a `Namespace` and a `ConfigMap` living in this namespace.
You don't need to deploy that in multiple step, cuebe takes care of deploying
the namespace first.

## Requirements

- Access to a Kubernetes cluster

## Running the example

To export YAML version of this release, run

```shell
cuebe export main.cue
```

To deploy this release in your current kubernetes context, run

```shell
cuebe apply main.cue
```
