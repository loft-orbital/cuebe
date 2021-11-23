package modfile

import (
	"golang.org/x/mod/module"
)

const CueModFile = "cue.mod/module.cue"

type File struct {
	Module  string           `json:"module"`
	Require []module.Version `json:"require,omitempty"`
}
