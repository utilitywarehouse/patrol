package patrol_test

import "testing"

func TestRepo(t *testing.T) {
	tests := RepoTests{
		RepoTest{
			TestdataFolder: "internalchange",
			Name:           "change within module",
			Description: "A change to a package within the same module\n" +
				"should flag depending packages as changed",
			AllFiles: false,
		},
		RepoTest{
			TestdataFolder: "modules",
			Name:           "change in go modules dependency",
			Description: "A change to a go modules dependency\n" +
				"should flag depending packages as changed",
			AllFiles: false,
		},
		RepoTest{
			TestdataFolder: "prefixcollision",
			Name:           "change in go modules dependency whose path is prefixed by another dependency",
			Description: "A change to a go modules dependency should flag depending packages as changed,\n" +
				"even when another required module's path is a prefix of the changed one\n" +
				"(e.g. ordering-platform is a prefix of ordering-platform-contracts).",
			AllFiles: false,
		},
		RepoTest{
			TestdataFolder: "vendoring",
			Name:           "change in vendored dependencies",
			Description: "A change to a vendored dependency\n" +
				"should flag depending packages as changed",
			AllFiles: false,
		},
		RepoTest{
			TestdataFolder: "exportedtesting",
			Name:           "change in a package with packagename_test test package",
			Description: "A change to package x that is being tested " +
				"using x_test package should not result in a stack overflow :D",
			AllFiles: false,
		},
		RepoTest{
			TestdataFolder: "submodules",
			Name:           "change in go modules dependency sub package",
			Description: "A change to a go modules dependency\n" +
				"should flag depending packages as changed",
			AllFiles: false,
		},
		RepoTest{
			TestdataFolder: "alias",
			Name:           "change in go modules dependency that was aliased",
			Description: "A change to a go modules dependency\n" +
				"should flag depending packages as changed",
			AllFiles: false,
		},
		RepoTest{
			TestdataFolder: "assets",
			Name:           "change in files that are not go source files",
			Description: "A change to a file that is not a go source file\n" +
				"should flag a package as changed",
			AllFiles: true,
		},
		RepoTest{
			TestdataFolder: "nestedassets",
			Name:           "change in files that are not go source files, nested",
			Description: "A change to a file that is not a go source file\n" +
				"should flag a sub package as changed",
			AllFiles: true,
		},
	}

	tests.Run(t)
}
