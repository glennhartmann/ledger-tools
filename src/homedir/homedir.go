package homedir

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

var (
	// overridable for testing
	osUserHomeDir = os.UserHomeDir
)

func FillInHomeDir(paths ...*string) error {
	for _, path := range paths {
		*path = strings.TrimSpace(*path)
		if strings.HasPrefix(*path, "~/") {
			home, err := osUserHomeDir()
			if err != nil {
				return errors.Wrap(err, "osUserHomeDir()")
			}
			*path = filepath.Join(home, strings.TrimPrefix(*path, "~/"))
		}
	}
	return nil
}
