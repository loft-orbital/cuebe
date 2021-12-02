## cuebe apply

Apply release to Kubernetes

### Synopsis

Apply CUE release to Kubernetes.

Apply uses server-side apply patch to apply the release.
For more information about server-side apply see:
  https://kubernetes.io/docs/reference/using-api/server-side-apply/


```
cuebe apply [flags]
```

### Examples

```

# Apply current directory with an encrypted file override
cuebe apply -i main.enc.yaml

# Extract Kubernetes context from CUE path
cuebe apply -c path.to.context

# Perform a dry-run (do not persist changes)
cuebe apply --dry-run

```

### Options

```
  -c, --context string           Kubernetes context, or a CUE path to extract it from.
      --dry-run                  Submit server-side request without persisting the resource.
  -e, --expression stringArray   Expressions to extract manifests from. Extract all manifests by default.
  -h, --help                     help for apply
  -i, --inject strings           Inject files into the release. Multiple format supported. Decrypt content with Mozilla sops if extension is .enc.*
  -p, --path string              Path to load CUE from. Default to current directory
  -t, --tag stringArray          Inject boolean or key=value tag.
```

### SEE ALSO

* [cuebe](cli/cuebe.md)	 - Handle CUE kubernetes release

