/*
Copyright © 2021 Loft Orbital

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package unifier

import (
	"fmt"

	"cuelang.org/go/cue"
	"cuelang.org/go/pkg/encoding/json"
	"cuelang.org/go/pkg/encoding/yaml"
)

// Unmarshaller can unmarshal raw bytes to cue.Value.
type Unmarshaller interface {
	Unmarshal(data []byte, ctx *cue.Context, options ...cue.BuildOption) (cue.Value, error)
}

// UnmarshallerFor returns the appropriate Unmarshaller for the extension or an error if the extension is not supported.
func UnmarshallerFor(ext string) (Unmarshaller, error) {
	switch ext {
	case ".cue":
		return &CUEUnmarshaller{}, nil
	case ".json":
		return &JSONUnmarshaller{}, nil
	case ".yaml", ".yml":
		return &YAMLUnmarshaller{}, nil
	default:
		return nil, fmt.Errorf("Unsupported extension %s", ext)
	}
}

// CUEUnmarshaller can unmarshal CUE content to cue.Value.
type CUEUnmarshaller struct{}

// Unmarshal unmarshals CUE-formatted data as a cue.Value.
func (c *CUEUnmarshaller) Unmarshal(data []byte, ctx *cue.Context, options ...cue.BuildOption) (cue.Value, error) {
	v := ctx.CompileBytes(data, options...)
	if v.Err() != nil {
		return v, fmt.Errorf("failed to unmarshal cue: %w", v.Err())
	}
	return v, nil
}

// JSONUnmarshaller can unmarshal JSON content to cue.Value.
type JSONUnmarshaller struct{}

// Unmarshal unmarshals JSON-formatted data as a cue.Value.
func (j *JSONUnmarshaller) Unmarshal(data []byte, ctx *cue.Context, options ...cue.BuildOption) (cue.Value, error) {
	exp, err := json.Unmarshal(data)
	if err != nil {
		return cue.Value{}, fmt.Errorf("failed to unmarshal json: %w", err)
	}
	v := ctx.BuildExpr(exp, options...)
	if v.Err() != nil {
		return v, fmt.Errorf("failed to unmarshal json: %w", v.Err())
	}
	return v, nil
}

// YAMLUnmarshaller can unmarshal YAML content to cue.Value.
type YAMLUnmarshaller struct{}

// Unmarshal unmarshals YAML-formatted data as a cue.Value.
func (y *YAMLUnmarshaller) Unmarshal(data []byte, ctx *cue.Context, options ...cue.BuildOption) (cue.Value, error) {
	exp, err := yaml.Unmarshal(data)
	if err != nil {
		return cue.Value{}, fmt.Errorf("failed to unmarshal yaml: %w", err)
	}
	v := ctx.BuildExpr(exp, options...)
	if v.Err() != nil {
		return v, fmt.Errorf("failed to unmarshal yaml: %w", v.Err())
	}
	return v, nil
}
