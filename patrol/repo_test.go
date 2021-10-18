package patrol_test

import "testing"

func TestRepo(t *testing.T) {
	tests := RepoTests{
		RepoTest{
			TestdataFolder: "internalchange",
			Name:           "change within module",
			Description: "A change to a package within the same module\n" +
				"should flag depending packages as changed",
			TestAgainstRevision: "HEAD~1",
			ExpectedChangedPackages: []string{
				"github.com/utilitywarehouse/internalchange/internal/bar",
				"github.com/utilitywarehouse/internalchange/pkg/foo",
				"github.com/utilitywarehouse/internalchange/pkg/cat",
			},
		},
		RepoTest{
			TestdataFolder: "modules",
			Name:           "change in go modules dependency",
			Description: "A change to a go modules dependency\n" +
				"should flag depending packages as changed",
			TestAgainstRevision: "HEAD~1",
			ExpectedChangedPackages: []string{
				"github.com/utilitywarehouse/modules",
			},
		},
		RepoTest{
			TestdataFolder: "vendoring",
			Name:           "change in vendored dependencies",
			Description: "A change to a vendored dependency\n" +
				"should flag depending packages as changed",
			TestAgainstRevision: "HEAD~1",
			ExpectedChangedPackages: []string{
				"github.com/utilitywarehouse/vendoring",
			},
		},
	}

	tests.Run(t)
}
