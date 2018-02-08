package config

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	expected := &Config{
		Dependencies: []Dependency{
			Dependency{
				name:    "MySQL",
				exeName: "mysql",
				path:    "/usr/bin/mysql",
				source:  "",
			},
			Dependency{
				name:    "PT Query Digest",
				exeName: "pt-query-digest",
				path:    "/usr/bin/pt-query-digest",
				source:  "",
			},
		},
	}

	actual, _ := New()

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("New should be %v, but got %v", expected, actual)
	}
}

func TestPrintStatus(t *testing.T) {
	h := Health{
		Conn:   "Open",
		Port:   "3306",
		Socket: "var/lib/xyz",
		Errors: nil,
	}

	expected := fmt.Sprintf("  %+v\n\n", h)

	var actualBuf bytes.Buffer
	h.PrintStatus(&actualBuf)

	// remove "&"" from buffer address string
	actual := strings.Replace(actualBuf.String(), "&", "", 1)

	if actual != expected {
		t.Errorf("PrintStatus of %v should be\n%s but got\n%s", h, expected, actual)
	}
}

func TestAddDependency(t *testing.T) {
	expected := &Config{
		Dependencies: []Dependency{
			Dependency{
				name:    "MySQL",
				exeName: "mysql",
				path:    "/usr/bin/mysql",
				source:  "",
			},
		},
	}

	actual := &Config{}

	actual.addDependency("MySQL", "mysql", "/usr/bin/mysql", "")

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("addDependency should be %v, but got %v", expected, actual)
	}

}

func TestVerifyDependencies(t *testing.T) {
	tt := []struct {
		name        string
		cfg         Config
		expectedCfg *Config
		expectedErr error
	}{
		{name: "Empty Configuration",
			cfg:         Config{},
			expectedCfg: nil,
			expectedErr: fmt.Errorf("cannot verify dependencies on an empty Config"),
		},
		{name: "Missing 1 of 1 Dependencies",
			cfg: Config{
				Dependencies: []Dependency{
					Dependency{
						name:    "Foo",
						exeName: "foo",
						path:    "/usr/bin/foo",
						source:  "",
					},
				},
			},
			expectedCfg: nil,
			expectedErr: fmt.Errorf("missing required dependency stat /usr/bin/foo: no such file or directory"),
		},
		{name: "Missing 1 of Multiple Dependencies",
			cfg: Config{
				Dependencies: []Dependency{
					Dependency{
						name:    "MySql",
						exeName: "mysql",
						path:    "/usr/bin/mysql",
						source:  "",
					},
					Dependency{
						name:    "Foo",
						exeName: "foo",
						path:    "/usr/bin/foo",
						source:  "",
					},
					Dependency{
						name:    "MySql",
						exeName: "mysql",
						path:    "/usr/bin/mysql",
						source:  "",
					},
				},
			},
			expectedCfg: nil,
			expectedErr: fmt.Errorf("missing required dependency stat /usr/bin/foo: no such file or directory"),
		},
		{name: "No Missing Dependencies",
			cfg: Config{
				Dependencies: []Dependency{
					Dependency{
						name:    "MySql",
						exeName: "mysql",
						path:    "/usr/bin/mysql",
						source:  "",
					},
				},
			},
			expectedCfg: &Config{
				Dependencies: []Dependency{
					Dependency{
						name:    "MySql",
						exeName: "mysql",
						path:    "/usr/bin/mysql",
						source:  "",
					},
				},
			},
			expectedErr: nil,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			actual, actualErr := tc.cfg.verifyDependencies()

			if !reflect.DeepEqual(actual, tc.expectedCfg) && (actualErr.Error() != tc.expectedErr.Error()) {
				t.Errorf("verifyDependencies of %s should be %+v, but got %+v", tc.name, tc.expectedCfg, actual)
			}

			if actualErr != nil && (tc.expectedErr.Error() != actualErr.Error()) {
				t.Errorf("verifyDependencies of %s should throw error \"%+v\", but got \"%v\"", tc.name, tc.expectedErr.Error(), actualErr.Error())
			}
		})
	}
}
