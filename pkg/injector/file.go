package injector

import "cuelang.org/go/cue"

type File struct {
	path   cue.Path
	result chan interface{}
}

func NewFile(src, srcPath, dstPath string) *File {
	r := make(chan interface{})
	// TODO fire gorotine to parse file
	return &File{path: cue.ParsePath(dstPath), result: r}
}

func (f *File) Inject(target cue.Value) cue.Value {
	r := <-f.result
	return target.FillPath(f.path, r)
}
