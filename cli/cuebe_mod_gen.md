## cuebe mod gen

genenerates CUE definitions from Go modules.

### Synopsis

Collects all cuegetgo.go files (including those in cue.mod/pkg/**) and
generates CUE definitions for all imported packages.

To add a new package to generate definitions for, include it in the import directive of your cuegetgo.go file.
Use a blank identifier to import the package solely for its side-effects.

~~~go
package cuegetgo

import (
  _ "k8s.io/api/apps/v1"
)
~~~



```
cuebe mod gen [flags]
```

### Options

```
  -h, --help   help for gen
```

### SEE ALSO

* [cuebe mod](cli/cuebe_mod.md)	 - manage CUE modules.

