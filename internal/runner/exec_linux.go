//go:build linux

package runner

import (
	"errors"
	"os"
	"os/exec"
	"syscall"
)

// Replace replaces the prox process with the target command. It only returns
// when the target cannot be located or executed.
func Replace(arguments []string, environment []string) (int, error) {
	path, err := exec.LookPath(arguments[0])
	if err != nil {
		return 127, err
	}
	if err := syscall.Exec(path, arguments, environment); err != nil {
		if errors.Is(err, os.ErrPermission) {
			return 126, err
		}
		return 125, err
	}
	return 0, nil
}
