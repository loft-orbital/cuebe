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
	"io/ioutil"
	"path"
	"strings"

	"go.mozilla.org/sops/v3/decrypt"
)

// ReadFile reads the file named by filename and returns the contents.
// If the filename ends by .enc.*, it will be decrypted (Mozilla sops format).
func ReadFile(filename string) ([]byte, error) {
	ext := path.Ext(filename)
	if strings.HasSuffix(filename, ".enc"+ext) {
		return decrypt.File(filename, ext)
	}
	return ioutil.ReadFile(filename)
}
