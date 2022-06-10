## cuebe delete

Delete all instances found in Build.

### Synopsis


Delete all instances found in provided context from the k8s cluster.

It first group manifests found in the context by instance.
Then it deletes those instances.
Cuebe delete respects the deletion policy annotation "instance.cuebe.loftorbital.com/deletion-policy".
		

```
cuebe delete [flags]
```

### Examples

```

# Delete all instances in the current dir
cuebe delete .

# Same but doing a dry-run
cuebe delete --dry-run .

# Delete using Kubernetes context from <Build>.path.to.context
cuebe apply -c .release.context .

# Delete using one of your available kubectl config context
cuebe apply -c colima .

```

### Options

```
  -c, --cluster string           Kube config context. If starting with a . (dot), it will be extracted from the Build at this CUE path.
      --dry-run                  Submit server-side request without persisting the resource.
  -e, --expression stringArray   Expressions to extract manifests from. Default to root.
  -f, --force                    Force apply.
  -h, --help                     help for delete
  -m, --manager string           Field manager. Override at your own risk. (default "cuebe")
  -t, --tag stringArray          Inject boolean or key=value tag.
```

### Options inherited from parent commands

```
      --timeout duration   Timeout, accpet any valid go Duration. (default 2m0s)
```

### SEE ALSO

* [cuebe](cli/cuebe.md)	 - Handle CUE kubernetes release

