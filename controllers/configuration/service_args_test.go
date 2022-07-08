package configuration

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func ptr(s string) *string {
	return &s
}

func TestUpdateServiceArguments(t *testing.T) {
	initialArguments := `--key=value
--other=other-value
--with-space value2
`
	for _, tc := range []struct {
		name              string
		update            map[string]*string
		expectedArguments []string
		expectedMissing   []string
		expectedUpdated   bool
	}{
		{
			name:   "simple-update",
			update: map[string]*string{"key": ptr("new-value")},
			expectedArguments: []string{
				"--key=new-value",
				"--other=other-value",
				"--with-space value2",
			},
			expectedUpdated: true,
		},
		{
			name:   "update-many-delete-one",
			update: map[string]*string{"--key": ptr("new-value"), "--other": ptr("other-new-value"), "with-space": nil},
			expectedArguments: []string{
				"--key=new-value",
				"--other=other-new-value",
			},
			expectedMissing: []string{
				"--with-space",
			},
			expectedUpdated: true,
		},
		{
			name:   "update-many-single-list",
			update: map[string]*string{"--key": ptr("new-value"), "--other": ptr("other-new-value")},
			expectedArguments: []string{
				"--key=new-value",
				"--other=other-new-value",
			},
			expectedUpdated: true,
		},
		{
			name: "no-updates",
			expectedArguments: []string{
				"--key=value",
				"--other=other-value",
				"--with-space value2",
			},
		},
		{
			name:   "new-opt",
			update: map[string]*string{"--new-opt": ptr("opt-value")},
			expectedArguments: []string{
				"--key=value",
				"--other=other-value",
				"--with-space value2",
				"--new-opt=opt-value",
			},
			expectedUpdated: true,
		},
		{
			name:   "delete-non-existent",
			update: map[string]*string{"--new-opt": nil, "new-opt-2": nil},
			expectedArguments: []string{
				"--key=value",
				"--other=other-value",
				"--with-space value2",
			},
			expectedUpdated: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			file := fmt.Sprintf("testdata/%s", tc.name)
			if err := os.WriteFile(file, []byte(initialArguments), 0660); err != nil {
				t.Fatalf("Expected no error setting up arguments file but received %q", err)
			}
			defer os.Remove(file)
			updated, err := updateServiceArguments(file, tc.update)
			if err != nil {
				t.Fatalf("Expected no error updating arguments file but received %q", err)
			}
			if updated != tc.expectedUpdated {
				t.Fatalf("Expected updated to be %v but it was %v instead", tc.expectedUpdated, updated)
			}
			b, err := os.ReadFile(file)
			if err != nil {
				t.Fatalf("Expected no error reading arguments file but received %q", err)
			}
			newArguments := string(b)
			for _, expected := range tc.expectedArguments {
				if !strings.Contains(newArguments, expected+"\n") {
					t.Fatalf("Expected new arguments to contain %q but they did not", expected)
				}
			}
			for _, notExpected := range tc.expectedMissing {
				if strings.Contains(newArguments, notExpected) {
					t.Fatalf("Expected new arguments to not contain %q but they did", notExpected)
				}
			}
		})
	}
}
