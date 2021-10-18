# Patrol

*Patrol* is a utility to help you understand what packages within a Go module
changed between commits. It was created to be used within CI pipelines to only
build what needs to be built, but maybe someone else can find a cool use for
this :). Patrol currently detects changes in:

- packages within the module itself
- `go.mod` dependency
- vendored dependencies

To understand all (potential) changes, Patrol traverses the whole dependencies
graph which means that if you have a structure that looks like this

```
external-package@v1.0.0 -> yourModule/foo -> yourModule/bar
```

and you update `external-package` to version `v1.0.1` Patrol will report
`yourModule/foo` (depending on `external-package`) and `yourModule/bar`
(depending on `yourModule/foo`) as changed.

## Getting started

### Install as binary
Patrol can be installed as any other Go binary, you just need to run

```
go install github.com/utilitywarehouse/patrol
```

and you can use it like this
``` patrol -from={commit hash} .  ```

This is an example run against [My Services
monorepo](https://github.com/utilitywarehouse/my-services-mono):

```
$ patrol -from=0a359e246ba3c7c76b0ad0e1d734ae103455b7a9 .

github.com/utilitywarehouse/my-services-mono/services/broadband-services-api/cmd/broadband-services-api
github.com/utilitywarehouse/my-services-mono/pkg/broadband
github.com/utilitywarehouse/my-services-mono/services/energy-services-projector/cmd/energy-services-projector
github.com/utilitywarehouse/my-services-mono/services/energy-services-projector/internal/handler
```

Patrol does nothing more than reporting what packages (or other packages they
depend on) changed in between commits. If for example your goal is to understand
what Docker images you should build as part of your CI run, and you know your
executables live under `services/`, you could by filtering your results with
[`ripgrep`](https://github.com/BurntSushi/ripgrep):

```
$ patrol -from=0a359e246ba3c7c76b0ad0e1d734ae103455b7a9 . | rg services

github.com/utilitywarehouse/my-services-mono/services/broadband-services-api/cmd/broadband-services-api
github.com/utilitywarehouse/my-services-mono/services/energy-services-projector/cmd/energy-services-projector
github.com/utilitywarehouse/my-services-mono/services/energy-services-projector/internal/handler
```

### Use as a Go library
If you want to integrate Patrol into your scripts, and your scripts are written
in Go (maybe using something like [mage](https://magefile.org/)) you can easily do so:

```golang
package main

import (
	"fmt"

	"github.com/utilitywarehouse/patrol/patrol"
)

func main() {

	repo, err := patrol.NewRepo("path/to/your/repo")
	if err != nil {
		panic(err)
	}

	revision := "a0e002f951f56d53d552f9427b3331b11ea66e92"

	changes, err := repo.ChangesFrom(revision)
	if err != nil {
		panic(err)
	}

	for _, c := range changes {
		fmt.Println(c)
	}
}
```

## Contributing
So did Patrol blow up on you or you finally saw an actual stack overflow? Graphs
do that sometimes. Sorry if that happened, but if you found you want to improve
or fix I'd suggest starting by writing a test.

Tests in this are fairly peculiar, but we need other repositories to test a tool
like Patrol. You can take a look at any of the case defined in
[`testdata`](patrol/testdata) but here's a checklist that might help:

- add a new folder within `patrol/testdata` with a name expressing what you're
  trying to test
- add as many folders as you'd like to `patrol/testdata/{your-test}/commits`.
  Each folder represents an actual commit, and they will be applied on top of
  each other as real commits when tests are run.
- add a new test case in [patrol/repo\_test.go](patrol/repo_test.go)
