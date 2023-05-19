package patrol

import (
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"golang.org/x/mod/modfile"
)

type Repo struct {
	path   string
	Module *modfile.File

	// map of packages, with the package name as key (e.g.:
	// github.com/uw-labs/patrol/patrol)
	Packages map[string]*Package
}

type Package struct {
	Name         string
	PartOfModule bool
	Dependants   []*Package
	Changed      bool
}

// NewRepo constructs a Repo from path, which needs to contain a go.mod file.
// It builds a map of all packages found in that repo and the dependencies
// between them.
func NewRepo(path string) (*Repo, error) {
	repo := &Repo{
		path:     path,
		Packages: map[string]*Package{},
	}

	// Parse go.mod
	b, err := os.ReadFile(filepath.Join(path, "go.mod"))
	if err != nil {
		return nil, err
	}

	mod, err := modfile.Parse(filepath.Join(path, "go.mod"), b, nil)
	if err != nil {
		return nil, err
	}

	repo.Module = mod

	// Find all go packages starting from path
	err = filepath.Walk(path, func(p string, f os.FileInfo, err error) error {
		if f.IsDir() && !directoryShouldBeIgnored(p) {
			fset := token.NewFileSet()

			// We're interested in each package imports at this point
			pkgs, err := parser.ParseDir(fset, p, nil, parser.ImportsOnly)
			if err != nil {
				return err
			}

			for _, pkg := range pkgs {
				if len(pkg.Files) == 0 {
					continue
				}

				var imports []string
				for _, file := range pkg.Files {
					// Don't map test packages
					if !strings.HasSuffix(file.Name.Name, "_test") {
						for _, imp := range file.Imports {
							imports = append(imports, strings.ReplaceAll(imp.Path.Value, `"`, ""))
						}
					}
				}
				repo.addPackage(strings.TrimPrefix(p, path+"/"), imports)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return repo, nil
}

// ChangesFrom returns a list of all packages within the repository (excluding
// packages in vendor/) that changed since the given revision. A package will
// be flagged as change if any file within the package itself changed or if any
// packages it imports (whether local, vendored or external modules) changed
// since the given revision.
func (r *Repo) ChangesFrom(revision string) ([]string, error) {
	err := r.detectInternalChangesFrom(revision)
	if err != nil {
		return nil, err
	}

	err = r.detectGoModulesChanges(revision)
	if err != nil {
		return nil, err
	}

	var changedOwnedPackages []string
	for _, pkg := range r.Packages {
		if pkg.PartOfModule && pkg.Changed {
			changedOwnedPackages = append(changedOwnedPackages, pkg.Name)
		}
	}

	return changedOwnedPackages, nil
}

// addPackage adds the package found at path to the repo, and also adds it as a
// dependant to all of the packages it imports.
func (r *Repo) addPackage(path string, imports []string) {
	var pkgName string

	// if path has vendor/ prefix, that needs to be removed to get the actual
	// package name
	if strings.HasPrefix(path, "vendor/") {
		pkgName = strings.TrimPrefix(path, "vendor/")
	} else {
		// if it doesn't have a vendor/ prefix it means it's part of our module and
		// path should be prefixed with the module name.
		pkgName = r.ModuleName()
		if path != r.path {
			pkgName += "/" + path
		}
	}

	// add the new package to the repo if it didn't exist already
	pkg, exists := r.Packages[pkgName]
	if !exists {
		pkg = &Package{
			Name:         pkgName,
			PartOfModule: r.OwnsPackage(pkgName),
		}
		r.Packages[pkgName] = pkg
	}

	// imports might not be a unique list, but we only want to add pkg as a
	// dependant to those packages once
	alreadyProcessedImports := map[string]interface{}{}
	for _, dependency := range imports {
		if _, alreadyProcessed := alreadyProcessedImports[dependency]; alreadyProcessed {
			continue
		}
		r.addDependant(pkg, dependency)
		alreadyProcessedImports[dependency] = struct{}{}

		// if the dependency is part of an external dependency (defined in go.mod)
		// add the parent module as a dependency as well so that a simple version
		// change would mark this package as changed
		if parent, ok := r.externalModule(dependency); ok {
			if _, alreadyProcessed := alreadyProcessedImports[parent]; alreadyProcessed {
				continue
			}
			r.addDependant(pkg, parent)
			alreadyProcessedImports[parent] = struct{}{}
		}
	}
}

// externalModule checks if the given package is part of one of the modules required
// as dependencies in go.mod. If it is it returns the name of the parent
// package and true.
func (r *Repo) externalModule(pkg string) (string, bool) {
	for _, req := range r.Module.Require {
		if strings.HasPrefix(pkg, req.Mod.Path) {
			return req.Mod.Path, true
		}
	}
	return "", false
}

// addDependant adds dependant as one of the dependants of the package
// identified by dependencyName (if it doesn't exist yet, it will be created).
func (r *Repo) addDependant(dependant *Package, dependencyName string) {
	dependency, exists := r.Packages[dependencyName]
	if !exists {
		dependency = &Package{
			Name:         dependencyName,
			PartOfModule: r.OwnsPackage(dependencyName),
		}
		r.Packages[dependencyName] = dependency
	}

	dependency.Dependants = append(dependency.Dependants, dependant)
}

// detectInternalChangesFrom will run a git diff (revision...HEAD) and flag as
// changed any packages (part of the module in repo or vendored packages) that
// have *.go files that are part of the that diff and packages that depend on them
func (r *Repo) detectInternalChangesFrom(revision string) error {
	repo, err := git.PlainOpen(r.path)
	if err != nil {
		return err
	}

	head, err := repo.Head()
	if err != nil {
		return err
	}

	// Get the HEAD commit
	now, err := repo.CommitObject(head.Hash())
	if err != nil {
		return err
	}

	// Get the tree for HEAD
	nowTree, err := now.Tree()
	if err != nil {
		return err
	}

	ref, err := repo.ResolveRevision(plumbing.Revision(revision))
	if err != nil {
		return err
	}

	// Find the commit for given revision
	then, err := repo.CommitObject(*ref)
	if err != nil {
		return err
	}

	// Get the tree for given revision
	thenTree, err := then.Tree()
	if err != nil {
		return err
	}

	// Get a diff between the two trees
	diff, err := nowTree.Diff(thenTree)
	if err != nil {
		return err
	}

	for _, change := range diff {
		// we're only interested in Go files
		if !strings.HasSuffix(change.From.Name, ".go") {
			continue
		}

		var pkgName string
		// if the changed file is in vendor/ stripping "vendor/" will give us the
		// package name
		if strings.HasPrefix(change.From.Name, "vendor/") {
			pkgName = strings.TrimPrefix(filepath.Dir(change.From.Name), "vendor/")
		}

		// package is part of our module
		if pkgName == "" {
			pkgName = r.ModuleName() + "/" + filepath.Dir(change.From.Name)
		}

		r.flagPackageAsChanged(pkgName)
	}

	return nil
}

// detectGoModulesChanges finds differences in dependencies required by
// HEAD:go.mod and {revision}:go.mod and flags as changed any packages
// depending on any of the changed dependencies.
func (r *Repo) detectGoModulesChanges(revision string) error {
	oldGoMod, err := r.getGoModFromRevision(revision)
	if err != nil {
		return err
	}

	differentModules := goModDifferences(oldGoMod, r.Module)
	for _, module := range differentModules {
		r.flagPackageAsChanged(module)
	}

	return nil
}

// getGoModFromRevision returns (if found) the go.mod file from the given
// revision.
func (r *Repo) getGoModFromRevision(revision string) (*modfile.File, error) {
	repo, err := git.PlainOpen(r.path)
	if err != nil {
		return nil, err
	}

	ref, err := repo.ResolveRevision(plumbing.Revision(revision))
	if err != nil {
		return nil, err
	}

	then, err := repo.CommitObject(*ref)
	if err != nil {
		return nil, err
	}

	file, err := then.File("go.mod")
	if err != nil {
		return nil, err
	}

	reader, err := file.Reader()
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := reader.Close(); err != nil {
			panic(err)
		}
	}()

	b, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	mod, err := modfile.Parse(filepath.Join(r.path, "go.mod"), b, nil)
	if err != nil {
		return nil, err
	}

	return mod, nil
}

// flagPackageAsChanged flags the package with the given name and all of its
// dependant as changed, recursively.
func (r *Repo) flagPackageAsChanged(name string) {
	pkg, exists := r.Packages[name]
	if !exists {
		return
	}

	if pkg.Changed {
		// assume change was already acked and save
		// some computation
		return
	}

	for _, d := range pkg.Dependants {
		r.flagPackageAsChanged(d.Name)
	}
	pkg.Changed = true
}

func (r *Repo) ModuleName() string {
	return r.Module.Module.Mod.Path
}

func (r *Repo) OwnsPackage(pkgName string) bool {
	return strings.HasPrefix(pkgName, r.ModuleName())
}

func directoryShouldBeIgnored(path string) bool {
	return strings.Contains(path, ".git")
}

// goModDifferences returns the list of packages name that were added, removed
// and/or updated between the two go.mod files
func goModDifferences(a, b *modfile.File) []string {
	differences := map[string]interface{}{} // keeping a map of unique differences
	// map is [package name]: version
	oldRequires := map[string]string{}
	for _, r := range a.Require {
		oldRequires[r.Mod.Path] = r.Mod.Version
	}

	newRequires := map[string]string{}
	for _, r := range b.Require {
		newRequires[r.Mod.Path] = r.Mod.Version
	}

	for oldPkg, oldVersion := range oldRequires {
		newVersion, exists := newRequires[oldPkg]
		if !exists {
			differences[oldPkg] = struct{}{}
			continue
		}

		if oldVersion != newVersion {
			differences[oldPkg] = struct{}{}
			continue
		}
	}

	for newPkg, newVersion := range newRequires {
		oldVersion, exists := oldRequires[newPkg]
		if !exists {
			differences[newPkg] = struct{}{}
			continue
		}

		if oldVersion != newVersion {
			differences[newPkg] = struct{}{}
			continue
		}
	}

	var results []string
	for pkg := range differences {
		results = append(results, pkg)
	}

	return results
}
