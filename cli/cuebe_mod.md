## cuebe mod

manage CUE modules.

### Synopsis

cuebe mod provides access to operations on CUE modules.
Most mod subcommands use a custom implementation of the MVS algorithm.
This is a temporary solution until [github.com/cue-lang#85](https://github.com/cue-lang/cue/issues/851) is addressed.

#### Adding requirements

Simply add a new entry to the 'require' key in cue.mod/module.cue file:

~~~cue
require: [
  { path: "github.com/tomato/ketchup", version: "v1.0.3" },
]
~~~

#### Private modules

When dealing with private modules, cuebe offers two solutions:
- leveraging credentials within your ~/.netrc file
- exporting **HOST_ADDRESS_TOKEN** and **HOST_ADDRESS_USER** environment variable (e.g. **GITHUB_COM_TOKEN** for github.com)


### Options

```
  -h, --help   help for mod
```

### Options inherited from parent commands

```
      --timeout duration   Timeout, accpet any valid go Duration. (default 2m0s)
```

### SEE ALSO

* [cuebe](cli/cuebe.md)	 - Handle CUE kubernetes release
* [cuebe mod gen](cli/cuebe_mod_gen.md)	 - genenerates CUE definitions from Go modules.
* [cuebe mod vendor](cli/cuebe_mod_vendor.md)	 - vendors requirements in cue.mod/pkg

