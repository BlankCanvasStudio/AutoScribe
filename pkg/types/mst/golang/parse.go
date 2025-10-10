package golang

import (
	"fmt"

	"golang.org/x/tools/go/packages"

	log "github.com/sirupsen/logrus"

	"github.com/BlankCanvasStudio/AutoScribe/pkg/types/mst"
)

/*
Summary:
Parses all Go packages under the given foldername using packages.Load with a configured set of needs, constructs a PackageNode for each loaded package, and delegates to PopulatePackageInformation to fill imports, type definitions, and function declarations. Returns the list of PackageNode pointers on success; on failure, returns an error and may return a partially populated slice.

Signature:
func ParsePackages(foldername string) ([]*PackageNode, error)

Parameters:
- foldername: string
  The folder path from which to load packages.

Returns:
- ([]*PackageNode, error)
  - On success: a slice of *PackageNode and nil error.
  - On load failure: nil slice and a non-nil error.
  - On per-package population failure: a non-nil error and a partially populated slice of PackageNode objects.

Errors/Exceptions:
- "failed to load package %v: %v" if packages.Load fails.
- "failed to populate package information: %v" if PopulatePackageInformation fails for any package (returns the current partial pkgNodes along with the error).

Side Effects:
- Logs debug information about the number of packages seen and declared functions.
- Allocates and returns new PackageNode objects; Mutates no external state beyond the returned data.

Edge Cases & Assumptions:
- If no packages are found, returns an empty slice and nil error.
- Assumes valid foldername and accessible filesystem; errors propagate as above.
- For each PackageNode, FunctionDecls is initialized as an empty []mst.FunctionDecl and Imports as an empty map[string]string before population.

*/
func ParsePackages(foldername string) ([]*PackageNode, error) {
	cfg := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedFiles |
			packages.NeedSyntax |
			packages.NeedCompiledGoFiles |
			packages.NeedSyntax |
			packages.NeedTypes |
			packages.NeedTypesInfo |
			packages.NeedImports |
			packages.NeedDeps,
	}

	pkgs, err := packages.Load(cfg, foldername)
	if err != nil {
		return nil, fmt.Errorf("failed to load package %v: %v", foldername, err)
	}

	pkgNodes := []*PackageNode{}

	log.Debugf("# packages seen: %v", len(pkgs))

	for _, pkg := range pkgs {
		pkgNode := &PackageNode{
			Package:       pkg,
			FunctionDecls: []mst.FunctionDecl{},
			// TypeDefinitions:      []*ast.TypeSpec{},
			Imports: make(map[string]string),
		}

		err = pkgNode.PopulatePackageInformation()
		if err != nil {
			return pkgNodes, fmt.Errorf("failed to populate package information: %v", err)
		}

		log.Debugf("%v functions declared", len(pkgNode.FunctionDecls))

		pkgNodes = append(pkgNodes, pkgNode)
	}

	return pkgNodes, nil
}
