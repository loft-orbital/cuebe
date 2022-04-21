## cuebe export

Export manifests as YAML.

### Synopsis


Export CUE release as a kubectl-compatible multi document YAML manifest.
If --output is set, manifests will be written here, one file by instances.
		

```
cuebe export [flags]
```

### Examples

```

# Export current directory with an encrypted file override
cuebe export -i main.enc.yaml

```

### Options

```
  -e, --expression strings   Expressions to extract manifests from. Default to root.
  -h, --help                 help for export
  -t, --tag stringArray      Inject boolean or key=value tag.
```

### Options inherited from parent commands

```
      --timeout duration   Timeout, accpet any valid go Duration. (default 2m0s)
```

### SEE ALSO

* [cuebe](cli/cuebe.md)	 - Handle CUE kubernetes release

