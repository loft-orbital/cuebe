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
	"io/fs"
	"os"
	"path"
	"strings"

	"go.mozilla.org/sops/v3/decrypt"
)

// ReadFile reads the file named by filename and returns the contents.
// If the filename ends by .enc.*, it will be decrypted (Mozilla sops format).
func ReadFile(filename string, fsys fs.FS) ([]byte, error) {
	if fsys == nil {
		fsys = os.DirFS("")
	}

	data, err := fs.ReadFile(fsys, filename)
	if err != nil {
		return nil, fmt.Errorf("could not read file: %w", err)
	}

	// encrypted file
	ext := path.Ext(filename)
	if strings.HasSuffix(filename, ".enc"+ext) {
		data, err = decrypt.Data(data, ext[1:])
		if err != nil {
			return nil, fmt.Errorf("could not decrypt data: %w", err)
		}
	}

	return data, nil
}
