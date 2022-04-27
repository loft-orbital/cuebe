## cuebe install

Install Cuebe to K8s cluster.

### Synopsis


Install Cuebe custom resource definitions to the k8s cluster.
		

```
cuebe install [flags]
```

### Examples

```

# Install to current config context.
cuebe install

# Same but targetting my-cluster.
cuebe install -c my-cluster

```

### Options

```
  -c, --cluster string   Kube config context.
      --dry-run          Submit server-side request without persisting the resource.
  -f, --force            Force apply.
  -h, --help             help for install
  -m, --manager string   Field manager. Override at your own risk. (default "cuebe")
```

### Options inherited from parent commands

```
      --timeout duration   Timeout, accpet any valid go Duration. (default 2m0s)
```

### SEE ALSO

* [cuebe](cli/cuebe.md)	 - Handle CUE kubernetes release

