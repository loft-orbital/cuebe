## cuebe apply

Apply context to k8s cluster.

### Synopsis

Apply context to k8s cluster.

Apply uses server-side apply patch to apply the context.
For more information about server-side apply see:
  https://kubernetes.io/docs/reference/using-api/server-side-apply/

It applies every manifests found in the provided context,
grouping them by instance if necessary.


```
cuebe apply [flags]
```

### Examples

```

# Apply current directory with an encrypted file override
cuebe apply . main.enc.yaml

# Extract Kubernetes context from <Build>.path.to.context
cuebe apply -c .release.context .

# Apply using one of your available kubectl config context
cuebe apply -c colima .

# Perform a dry-run (do not persist changes)
cuebe apply --dry-run .

```

### Options

```
  -c, --cluster string           Kube config context. If starting with a . (dot), it will be extracted from the Build at this CUE path.
      --dry-run                  Submit server-side request without persisting the resource.
  -e, --expression stringArray   Expressions to extract manifests from. Default to root.
  -f, --force                    Force apply.
  -h, --help                     help for apply
  -m, --manager string           Field manager. Override at your own risk. (default "cuebe")
  -t, --tag stringArray          Inject boolean or key=value tag.
```

### Options inherited from parent commands

```
      --timeout duration   Timeout, accpet any valid go Duration. (default 2m0s)
```

### SEE ALSO

* [cuebe](cli/cuebe.md)	 - Handle CUE kubernetes release

