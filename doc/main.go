package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/go-git/go-billy/v5/osfs"
	"github.com/loft-orbital/cuebe/cmd/cuebe/cmd"
	"github.com/loft-orbital/cuebe/doc/cli"
	"github.com/loft-orbital/cuebe/internal/mod"
)

func main() {
	rootpath := "build/doc"
	staticpath := "doc/static"

	os.RemoveAll(rootpath)
	rfs := osfs.New(rootpath)

	// Copy static files
	fmt.Print("Copying static files...")
	static := osfs.New(staticpath)
	if err := mod.BillyCopy(rfs, static); err != nil {
		log.Fatal(fmt.Errorf(" \u274C: %w", err))
	}
	fmt.Println(" \u2705")

	// Home page
	fmt.Print("Creating homepage...")
	hp, err := rfs.Create("README.md")
	if err := mod.BillyCopy(rfs, static); err != nil {
		log.Fatal(fmt.Errorf(" \u274C: %w", err))
	}
	defer hp.Close()
	rm, err := os.Open("README.md")
	if err := mod.BillyCopy(rfs, static); err != nil {
		log.Fatal(fmt.Errorf(" \u274C: %w", err))
	}
	defer rm.Close()
	if _, err := io.Copy(hp, rm); err != nil {
		log.Fatal(fmt.Errorf(" \u274C: %w", err))
	}
	fmt.Println(" \u2705")

	// generate cli doc
	fmt.Print("Generating cuebe cli documentation...")
	cliFS, err := rfs.Chroot("cli")
	if err != nil {
		log.Fatal(fmt.Errorf(" \u274C: %w", err))
	}
	if err := cli.Build(cmd.RootCmd, cliFS, rfs.Root()); err != nil {
		log.Fatal(fmt.Errorf(" \u274C: %w", err))
	}
	fmt.Println(" \u2705")
}
