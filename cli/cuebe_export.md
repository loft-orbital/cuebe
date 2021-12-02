## cuebe export

Export manifests as YAML.

### Synopsis

Export CUE release as a kubectl-compatible multi document YAML manifest.

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
  -e, --expression stringArray   Expressions to extract manifests from. Extract all manifests by default.
  -h, --help                     help for export
  -i, --inject strings           Raw YAML files to inject. Can be encrypted with sops.
  -p, --path string              Path to load CUE from. Default to current directory
  -t, --tag stringArray          Inject boolean or key=value tag.
```

### SEE ALSO

* [cuebe](cli/cuebe.md)	 - Handle CUE kubernetes release

