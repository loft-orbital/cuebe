// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mod

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

type netrcLine struct {
	machine  string
	login    string
	password string
}

type NetrcMachine struct {
	Login    string
	Password string
}

var (
	netrcOnce sync.Once
	netrc     map[string]NetrcMachine
	netrcErr  error
)

func NetrcMachines() map[string]NetrcMachine {
	netrcOnce.Do(readNetrc)

	return netrc
}

func NetrcCredentials(machine string) (NetrcMachine, error) {
	machines := NetrcMachines()

	if cred, ok := machines[machine]; ok {
		return cred, nil
	}

	return NetrcMachine{}, fmt.Errorf("No credentials found for %s", machine)
}

func parseNetrc(data string) map[string]NetrcMachine {
	// See https://www.gnu.org/software/inetutils/manual/html_node/The-_002enetrc-file.html
	// for documentation on the .netrc format.
	nrc := make(map[string]NetrcMachine)
	var l netrcLine
	inMacro := false
	for _, line := range strings.Split(data, "\n") {
		if inMacro {
			if line == "" {
				inMacro = false
			}
			continue
		}

		f := strings.Fields(line)
		i := 0
		for ; i < len(f)-1; i += 2 {
			// Reset at each "machine" token.
			// “The auto-login process searches the .netrc file for a machine token
			// that matches […]. Once a match is made, the subsequent .netrc tokens
			// are processed, stopping when the end of file is reached or another
			// machine or a default token is encountered.”
			switch f[i] {
			case "machine":
				l = netrcLine{machine: f[i+1]}
			case "default":
				break
			case "login":
				l.login = f[i+1]
			case "password":
				l.password = f[i+1]
			case "macdef":
				// “A macro is defined with the specified name; its contents begin with
				// the next .netrc line and continue until a null line (consecutive
				// new-line characters) is encountered.”
				inMacro = true
			}

			if l.machine != "" && l.login != "" && l.password != "" {
				nrc[l.machine] = NetrcMachine{l.login, l.password}
				l = netrcLine{}
			}
		}

		if i < len(f) && f[i] == "default" {
			// “There can be only one default token, and it must be after all machine tokens.”
			break
		}
	}

	return nrc
}

func netrcPath() (string, error) {
	if env := os.Getenv("NETRC"); env != "" {
		return env, nil
	}
	dir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	base := ".netrc"
	if runtime.GOOS == "windows" {
		base = "_netrc"
	}
	return filepath.Join(dir, base), nil
}

func readNetrc() {
	path, err := netrcPath()
	if err != nil {
		netrcErr = err
		return
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			netrcErr = err
		}
		return
	}

	netrc = parseNetrc(string(data))
}
