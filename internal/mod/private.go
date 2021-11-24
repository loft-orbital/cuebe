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
package mod

import (
	"os"
	"path"
	"strings"
)

// CuePrivatePattern extracts the pattern to find private module from CUEPRIVATE environment variable.
var CuePrivatePattern string = os.Getenv("CUEPRIVATE")

// MatchPrefixPatterns reports whether any path prefix of target matches one of
// the glob patterns (as defined by path.Match) in the comma-separated globs
// list. This implements the algorithm used when matching a module path to the
// GOPRIVATE environment variable, as described by 'go help module-private'.
//
// It ignores any empty or malformed patterns in the list.
func MatchPrefixPatterns(globs, target string) bool {
	for globs != "" {
		// Extract next non-empty glob in comma-separated list.
		var glob string
		if i := strings.Index(globs, ","); i >= 0 {
			glob, globs = globs[:i], globs[i+1:]
		} else {
			glob, globs = globs, ""
		}
		if glob == "" {
			continue
		}

		// A glob with N+1 path elements (N slashes) needs to be matched
		// against the first N+1 path elements of target,
		// which end just before the N+1'th slash.
		n := strings.Count(glob, "/")
		prefix := target
		// Walk target, counting slashes, truncating at the N+1'th slash.
		for i := 0; i < len(target); i++ {
			if target[i] == '/' {
				if n == 0 {
					prefix = target[:i]
					break
				}
				n--
			}
		}
		if n > 0 {
			// Not enough prefix elements.
			continue
		}
		matched, _ := path.Match(glob, prefix)
		if matched {
			return true
		}
	}
	return false
}

// IsPrivate returns true if mod is a private module.
func IsPrivate(mod string) bool {
	return MatchPrefixPatterns(CuePrivatePattern, mod)
}

// CredentialsFor gets the user and password credentials for host.
// First looking at <HOST_DOMAIN>_TOKEN and <HOST_DOMAIN>_USER env var
// and fallbacking to netrc if they are not set.
func CredentialsFor(host string) (usr, pwd string) {
	// Try with environment variable
	env := strings.ToUpper(strings.ReplaceAll(host, ".", "_"))
	pwd = os.Getenv(env + "_TOKEN")
	usr = os.Getenv(env + "_USER")

	// Fallback to netrc
	if pwd == "" {
		netrc, err := NetrcCredentials(host)
		if err != nil {
			return
		}
		usr = netrc.Login
		pwd = netrc.Password
	}

	return
}
