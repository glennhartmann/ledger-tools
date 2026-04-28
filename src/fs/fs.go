// Package fs is a wrapper around the `afero` package to make testing easy.
package fs

import (
	"regexp"

	"github.com/spf13/afero"
)

var (
	defaultFS = afero.NewOsFs()
	fs        = defaultFS
)

// Get returns the current afero fs.
func Get() afero.Fs {
	return fs
}

// SetUpTestFilesMemFSOverlayForTesting sets up the file system as an overlay,
// where the read-only base is a regexp-filtered OsFS that only contains
// testfiles/*, and the writable overlay is a MemMapFs. That means input files
// from testfiles/* can be read in the test, but anything written will happen
// in-memory, and you can be confident in testing "file does not exist"
// scenarios with anything outside of testfiles/*. Remember to call
// `ResetForTesting()` after the test.
func SetUpTestFilesMemFSOverlayForTesting() afero.Fs {
	rxFS := afero.NewRegexpFs(defaultFS, regexp.MustCompile(`testfiles/.*`))
	overlayFS := afero.NewCopyOnWriteFs(rxFS, afero.NewMemMapFs())
	fs = overlayFS
	return fs
}

// SetUpMemFSOverlayForTesting sets up a simple in-memory FS for testing. No
// files from disk are accessible. Remember to call `ResetForTesting()` after
// the test.
func SetUpMemFSOverlayForTesting() afero.Fs {
	memFS := afero.NewMemMapFs()
	fs = memFS
	return fs
}

// ResetForTesting resets the file system to its default state. This should be
// called at the end of tests that call either of the "SetUp*ForTesting"
// methods.
func ResetForTesting() {
	fs = defaultFS
}
