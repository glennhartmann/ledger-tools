package fs

import (
	"testing"

	"github.com/spf13/afero"
)

func TestSetUpTestFilesMemFSOverlayForTesting(t *testing.T) {
	overlayFS := SetUpTestFilesMemFSOverlayForTesting()
	defer ResetForTesting()

	testHelper(t, overlayFS, true /* overlay */)
}

func TestSetUpMemFSOverlayForTesting(t *testing.T) {
	memFS := SetUpMemFSOverlayForTesting()
	defer ResetForTesting()

	testHelper(t, memFS, false /* overlay */)
}

func testHelper(t *testing.T, fs afero.Fs, overlay bool) {
	t.Helper()

	osFS := afero.NewOsFs()

	fstest := "fs_test.go"

	// can't access arbitrary files from disk
	test := func(innerFS afero.Fs) {
		if exists, err := afero.Exists(innerFS, fstest); err != nil {
			t.Errorf("afero.Exists(): %+v", err)
		} else if exists {
			t.Errorf("%q exists, but it shouldn't", fstest)
		}
	}
	test(fs)
	test(Get())

	// can create new files in memory which don't persist to disk
	test = func(innerFS afero.Fs) {
		file := "testfiles/newfile_fortesting"
		if err := afero.WriteFile(innerFS, file, []byte("1234"), 0644); err != nil {
			t.Errorf("afero.WriteFile(): %+v", err)
		}
		defer innerFS.Remove(file)

		if b, err := afero.ReadFile(innerFS, file); err != nil {
			t.Errorf("afero.ReadFile(): %+v", err)
		} else if string(b) != "1234" {
			t.Errorf("file contains %q, expected '1234'", string(b))
		}
		if exists, err := afero.Exists(osFS, file); err != nil {
			t.Errorf("afero.Exists(): %+v", err)
		} else if exists {
			t.Errorf("%q exists on osFS, but it shouldn't", file)
		}
	}
	test(fs)
	test(Get())

	asdf := "testfiles/asdf.txt"
	if overlay {
		// can read from testfiles/*
		test = func(innerFS afero.Fs) {
			if b, err := afero.ReadFile(innerFS, asdf); err != nil {
				t.Errorf("afero.ReadFile(): %+v", err)
			} else if string(b) != "asdf\n" {
				t.Errorf("file contains %q, expected 'asdf\\n'", string(b))
			}
		}
		test(fs)
		test(Get())

		// can "overwrite" a disk file in memory; changes don't persist to disk
		test = func(innerFS afero.Fs) {
			if err := afero.WriteFile(innerFS, asdf, []byte("1234"), 0644); err != nil {
				t.Errorf("afero.WriteFile(): %+v", err)
			}
			if b, err := afero.ReadFile(innerFS, asdf); err != nil {
				t.Errorf("afero.ReadFile(): %+v", err)
			} else if string(b) != "1234" {
				t.Errorf("file contains %q, expected '1234'", string(b))
			}
			if b, err := afero.ReadFile(osFS, asdf); err != nil {
				t.Errorf("afero.ReadFile(): %+v", err)
			} else if string(b) != "asdf\n" {
				t.Errorf("file contains %q, expected 'asdf\\n'", string(b))
			}
		}
		test(fs)
		test(Get())
	} else {
		// can't read from testfiles/*
		test = func(innerFS afero.Fs) {
			if exists, err := afero.Exists(innerFS, asdf); err != nil {
				t.Errorf("afero.Exists(): %+v", err)
			} else if exists {
				t.Errorf("%q exists, but it shouldn't", asdf)
			}
		}
		test(fs)
		test(Get())
	}

	// resetting lets us use disk again
	ResetForTesting()

	if b, err := afero.ReadFile(Get(), fstest); err != nil {
		t.Errorf("afero.ReadFile(): %+v", err)
	} else if len(b) == 0 {
		t.Errorf("file is empty, but shouldn't be")
	}
}
