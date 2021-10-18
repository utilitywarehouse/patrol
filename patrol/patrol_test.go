package patrol_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/utilitywarehouse/patrol/patrol"
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
	// create tmp dir for the test
	tmp, err := os.MkdirTemp("", "")
	require.NoError(t, err)

	fmt.Println(tmp)

	// init repo
	repo, err := git.PlainInit(tmp, false)
	require.NoError(t, err)

	// loop over all the commits for this test case
	commitsDir := filepath.Join("testdata", test.TestdataFolder, "commits")
	versions, err := os.ReadDir(commitsDir)
	require.NoError(t, err)

	for i, v := range versions {
		// copy all files from a "commit"
		err = copy(filepath.Join(commitsDir, v.Name()), tmp)
		require.NoError(t, err)

		worktree, err := repo.Worktree()
		require.NoError(t, err)

		err = worktree.AddGlob(".")
		require.NoError(t, err)

		// make a new commit
		_, err = worktree.Commit(fmt.Sprintf("commit #%v", i+1), &git.CommitOptions{
			Author: &object.Signature{
				Name:  "patrol test",
				Email: "patrol@test.me",
				When:  time.Now(),
			},
		})
		require.NoError(t, err)
	}

	r, err := patrol.NewRepo(tmp)
	require.NoError(t, err)

	changes, err := r.ChangesFrom(test.TestAgainstRevision)
	require.NoError(t, err)
	assert.Equal(t, test.ExpectedChangedPackages, changes)
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
			err := os.Mkdir(filepath.Join(destination, relPath), 0755)
			if err != nil {
				if errors.Is(err, os.ErrExist) {
					return nil
				}
				return err
			}
		} else {
			data, err1 := ioutil.ReadFile(filepath.Join(source, relPath))
			if err1 != nil {
				return err1
			}
			return ioutil.WriteFile(filepath.Join(destination, relPath), data, 0777)
		}

		return nil
	})
	return err
}
