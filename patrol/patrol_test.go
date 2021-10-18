package patrol_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type RepoTest struct {
	// the test folder that should be used for testing
	TestdataFolder string

	// name of the test
	Name string

	// description of the test, what are trying to assess?
	Description string

	// the git revision against which changes should be detected
	TestAgainstRevision string

	// the list of expected packages that should be flagged as changed between
	// HEAD and TestAgainstRevision
	ExpectedChangedPackages []string
}

func (test *RepoTest) Run(t *testing.T) {
	tmp, err := os.MkdirTemp("", "")
	require.NoError(t, err)

	fmt.Println(tmp)
	err = copy(filepath.Join("testdata", test.TestdataFolder), tmp)
}

type RepoTests []RepoTest

func (tests RepoTests) Run(t *testing.T) {
	for _, test := range tests {
		test.Run(t)
	}
}

func copy(source, destination string) error {
	var err error = filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		var relPath string = strings.Replace(path, source, "", 1)
		if relPath == "" {
			return nil
		}
		if info.IsDir() {
			return os.Mkdir(filepath.Join(destination, relPath), 0755)
		} else {
			data, err1 := ioutil.ReadFile(filepath.Join(source, relPath))
			if err1 != nil {
				return err1
			}
			return ioutil.WriteFile(filepath.Join(destination, relPath), data, 0777)
		}
	})
	return err
}
