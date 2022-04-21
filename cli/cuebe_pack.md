## cuebe pack

Package a context.

### Synopsis


Package a context into a tar.gz archive.
This archive can be used later as a base context.

Package respects the .cuebeignore directives.


```
cuebe pack [flags]
```

### Examples

```

# Pack current directory
cuebe pack .

# Merge dir1/ and dir2/ and pack them
cuebe pack dir1/ dir2/

```

### Options

```
  -h, --help            help for pack
  -o, --output string   Output file. (default "cube.tar.gz")
```

### Options inherited from parent commands

```
      --timeout duration   Timeout, accpet any valid go Duration. (default 2m0s)
```

### SEE ALSO

* [cuebe](cli/cuebe.md)	 - Handle CUE kubernetes release

