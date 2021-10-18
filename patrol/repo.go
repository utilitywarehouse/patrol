package patrol

import (
	"go/parser"
	"go/token"
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

func NewRepo(path string) (*Repo, error) {
	repo := &Repo{
		path:     path,
		Packages: map[string]*Package{},
	}

	b, err := os.ReadFile(filepath.Join(path, "go.mod"))
	if err != nil {
		return nil, err
	}

	mod, err := modfile.Parse(filepath.Join(path, "go.mod"), b, nil)
	if err != nil {
		return nil, err
	}

	repo.Module = mod

	err = filepath.Walk(path, func(p string, f os.FileInfo, err error) error {
		if f.IsDir() && !directoryShouldBeIgnored(p) {
			fset := token.NewFileSet()
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
					for _, imp := range file.Imports {
						imports = append(imports, strings.ReplaceAll(imp.Path.Value, `"`, ""))
					}
				}
				repo.AddPackage(strings.TrimPrefix(p, path+"/"), imports)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return repo, nil
}

func (r *Repo) AddPackage(path string, imports []string) {
	var pkgName string

	if strings.HasPrefix(path, "vendor/") {
		pkgName = strings.TrimPrefix(path, "vendor/")
	} else {
		pkgName = r.ModuleName()
		if path != r.path {
			pkgName += "/" + path
		}
	}

	pkg, exists := r.Packages[pkgName]
	if !exists {
		pkg = &Package{
			Name:         pkgName,
			PartOfModule: r.OwnsPackage(pkgName),
		}
		r.Packages[pkgName] = pkg
	}

	alreadyProcessedImports := map[string]interface{}{}
	for _, dependency := range imports {
		if _, alreadyProcessed := alreadyProcessedImports[dependency]; alreadyProcessed {
			continue
		}
		r.AddDependant(pkg, dependency)
		alreadyProcessedImports[dependency] = struct{}{}
	}
}

func (r *Repo) AddDependant(dependant *Package, dependencyName string) {
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

func (r *Repo) ChangesFrom(revision string) ([]string, error) {
	err := r.detectInternalChangesFrom(revision)
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

func (r *Repo) detectInternalChangesFrom(revision string) error {
	// git diff go files
	repo, err := git.PlainOpen(r.path)
	if err != nil {
		return err
	}

	head, err := repo.Head()
	if err != nil {
		return err
	}

	now, err := repo.CommitObject(head.Hash())
	if err != nil {
		return err
	}

	nowTree, err := now.Tree()
	if err != nil {
		return err
	}

	ref := plumbing.NewHash(revision)
	then, err := repo.CommitObject(ref)
	if err != nil {
		return err
	}

	thenTree, err := then.Tree()
	if err != nil {
		return err
	}

	diff, err := nowTree.Diff(thenTree)
	if err != nil {
		return err
	}

	for _, change := range diff {
		if !strings.HasSuffix(change.From.Name, ".go") {
			continue
		}

		pkgName := r.ModuleName() + "/" + filepath.Dir(change.From.Name)
		r.flagPackageAsChanged(pkgName)
	}

	return nil
}

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

	pkg.Changed = true
	for _, d := range pkg.Dependants {
		d.Changed = true
	}
}

func (r *Repo) ModuleName() string {
	return r.Module.Module.Mod.Path
}

func (r *Repo) OwnsPackage(pkgName string) bool {
	return strings.HasPrefix(pkgName, r.ModuleName())
}

type Package struct {
	Name         string
	PartOfModule bool
	Dependants   []*Package
	Changed      bool
}

func directoryShouldBeIgnored(path string) bool {
	return strings.Contains(path, ".git")
}
