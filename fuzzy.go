package process

import (
	"strings"
)

func fuzzy(match func(string) bool) func(*Process) bool {
	return func(pv *Process) bool {
		if match(pv.Name) ||
		match(pv.Cwd) ||
		match(pv.State) ||
		match(pv.Cmdline) ||
		match(pv.Username) ||
		match(pv.Executable) ||
		match(strings.Join(pv.Args, " ")) {
			return true
		}

		return false
	}
}

