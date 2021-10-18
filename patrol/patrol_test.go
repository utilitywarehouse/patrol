package patrol_test

import (
	"fmt"
	"os"
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

type RepoTests []RepoTest

func (tests RepoTests) Test(t *testing.T) {
	tmp, err := os.MkdirTemp("", "")
	require.NoError(t, err)

	fmt.Println(tmp)
}
