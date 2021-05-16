package homedir

import (
	"testing"

	"fmt"

	"github.com/prashantv/gostub"
)

func TestFillInHomeDir(t *testing.T) {
	stubs := gostub.New()
	defer stubs.Reset()

	inPathsNoPtr1 := []string{"asdf/234", "/dev/null", "~/.conf", "../../something", "~home/a2",
		"./~", "~/some/../../../thing"}
	wantPaths1 := []string{"asdf/234", "/dev/null", "/home/user/.conf", "../../something", "~home/a2",
		"./~", "/thing"}

	inPaths1 := make([]*string, len(inPathsNoPtr1))
	for i := range inPathsNoPtr1 {
		inPaths1[i] = &inPathsNoPtr1[i]
	}

	inPathsNoPtr2 := []string{"/home/unused", "~/unused"}

	tests := []struct {
		name       string
		homeDir    string
		homeDirErr bool
		paths      []*string
		wantPaths  []string
		wantErr    bool
	}{
		{
			name:       "success",
			homeDir:    "/home/user",
			homeDirErr: false,
			paths:      inPaths1,
			wantPaths:  wantPaths1,
			wantErr:    false,
		},
		{
			name:       "error",
			homeDir:    "/home/user",
			homeDirErr: true,
			paths:      []*string{&inPathsNoPtr2[0], &inPathsNoPtr2[1]},
			wantPaths:  nil,
			wantErr:    true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stubs.Stub(&osUserHomeDir, func() (string, error) {
				if test.homeDirErr {
					return "", fmt.Errorf("this is an error")
				}
				return test.homeDir, nil
			})

			err := FillInHomeDir(test.paths...)
			if (err != nil) != test.wantErr {
				t.Fatalf("FillInHomeDir() = err(%+v), wanted err(%v)", err, test.wantErr)
			}
			if err != nil {
				return
			}

			for i := range test.paths {
				// crashes if len(test.wantPaths) < len(test.paths), so don't do that
				if *test.paths[i] != test.wantPaths[i] {
					t.Errorf("FillInHomeDir()[%d] = %q, wanted %q", i, *test.paths[i], test.wantPaths[i])
				}
			}
		})
	}
}
