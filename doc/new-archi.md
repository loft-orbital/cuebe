# Cuebe new architecture

## Manifest

The manifest is a specification of a Kubernetes API object.
It is often represented as a JSON or YAML file.
With `cuebe` a manifest can be written in CUE.
This CUE manifest has to be concrete in order to be deployed or exported.

## Instance

An instance is a group of one or many Manifests, together forming an application.
A Manifest is identified as part of an instance with a label `cuebe.loft-orbital.com/instance`.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: myconfig
  labels:
    cuebe.loft-orbital.com/instance: foo
```

Is part of the _foo_ instance.
As it is a bit cumbersome to retrieve every resources given a label,
`cuebe` keeps a reference of every Manifest belonging to an instance in a special _ConfigMap_.
`cuebe` create and update this ConfigMap in the `cuebe-system` namespace.
The data field of this ConfigMap contains a reference to every instance's Manifest by using `<apiVersion>/<kind>/<metadata.name>` as key
and a sha1 hash of the Manifest as a value.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-db
  namespace: cuebe-system
data:
  v1/ConfigMap/db-config: c38f650d1eb1c15ed9c2a0534186d01e4248fe75
  apps/v1/StatefulSet/postgres: 3b9275790eae37d2c06dd1c853a3cd4885698ae3
  v1/Service/postgres: 9848b49dda7c5a84e6460e844bd5d9a878c1a8ed
```

This instance _my-db_ is composed of a ConfigMap, a Service and a StatefulSet.

## Context

A context is an aggregation of files (local or remote) of any type.
A context is passed to cuebe during a Build to export or apply Manifests.
`cuebe` keeps the folder structure of Context passed to a Build.

## Package

You can package a Context to use it later as a base for your Build.
Be sure not to include any sensible data in your Context before packaging it.
`cuebe` will rely on a `.cuebeignore` file to help you keep your packages clean.

## Build

A Build is the action of using a Context to create Manifests.
You can then export them as standard k8s YAML manifests, or directly deploy them to one or multiple k8s clusters.
When applying Manifests that are part of an instance, `cuebe` will take care of creating / updating the instance ConfigMap in `cuebe-system` namespace.
