package patrol_test

import "testing"

func TestRepo(t *testing.T) {
	tests := RepoTests{
		RepoTest{
			TestdataFolder: "internalchange",
			Name:           "change within module",
			Description: "A change to a package within the same module\n" +
				"should flag depending packages as changed",
			TestAgainstRevision:     "HEAD~1",
			ExpectedChangedPackages: []string{},
		},
	}

	tests.Run(t)
}
