package main

import (
	"fmt"
	"log"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/load"
)

type TypeMeta struct {
	Kind       string `json:"kind"`
	ApiVersion string `json:"apiVersion"`
}

func main() {
	// We need a cue.Context, the New'd return is ready to use
	ctx := cuecontext.New()

	kobj := ctx.EncodeType(TypeMeta{})
	if kobj.Err() != nil {
		log.Fatal("Failed to encode type meta", kobj.Err())
	}
	fmt.Println("kobj", kobj)

	// The entrypoints are the same as the files you'd specify at the command line
	entrypoints := []string{}

	// Load Cue files into Cue build.Instances slice
	// the second arg is a configuration object, we'll see this later
	bis := load.Instances(entrypoints, nil)

	// Loop over the instances, checking for errors and printing
	for _, bi := range bis {
		fmt.Println("Loop", bi.Deps)
		// check for errors on the instance
		// these are typically parsing errors
		if bi.Err != nil {
			fmt.Println("Error during load:", bi.Err)
			continue
		}

		// Use cue.Context to turn build.Instance to cue.Instance
		value := ctx.BuildInstance(bi)
		if value.Err() != nil {
			fmt.Println("Error during build:", value.Err())
			continue
		}

		value = value.LookupPath(cue.ParsePath("r.resources"))

		value.Walk(func(v cue.Value) bool {
			fmt.Println("step")
			if kobj.Subsumes(v) {
				fmt.Println("K8s object found:", v)
				fmt.Println(v.Path())
				return false
			}
			return true
		}, nil)

		// Validate the value
		err := value.Validate()
		if err != nil {
			fmt.Println("Error during validate:", err)
			continue
		}
	}

}
