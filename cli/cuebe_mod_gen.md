## cuebe mod gen

genenerates CUE definitions from Go modules.

### Synopsis

Collects all cuegetgo.go files (including those in cue.mod/pkg/**) and
generates CUE definitions for all imported packages.

To add a new package to generate definitions for, include it in the godef list of your cue.mod/module.cue file.
You can fix version by appending @version/

~~~cue
module: "github.com/company/module"

godef: [
  "k8s.io/api/apps/v1",
  "k8s.io/api/batch/v1@v0.23.3",
  ]
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

