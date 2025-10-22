package golang

import (
	"fmt"

	"golang.org/x/tools/go/packages"

	log "github.com/sirupsen/logrus"

	"github.com/BlankCanvasStudio/AutoScribe/pkg/types/mst"
)

/*
Summary:
ParsePackages loads Go packages from the given folder, builds a PackageNode for each loaded package,
and populates per-package metadata by calling PopulatePackageInformation on each node. The result
is a slice of *PackageNode ready for analysis, or an error if loading or population fails.

Signature:
func ParsePackages(foldername string) ([]*PackageNode, error)

Parameters:
- foldername: string — path to the root folder to load packages from.

Returns:
- []*PackageNode: a slice of PackageNode pointers, one per loaded package (order as returned by packages.Load).
- error: non-nil if loading packages fails or if PopulatePackageInformation fails for any package.

Errors/Exceptions:
- non-nil error when packages.Load fails: "failed to load package %v: %v"
- non-nil error when a package cannot be populated: "failed to populate package information: %v"

Side Effects:
- Logs debugging information: number of packages seen, and number of functions declared per package.
- Constructs and mutates in-memory PackageNode structures (Package, FunctionDecls, Imports) per package.

Edge Cases & Assumptions:
- If no syntax files are present, ParsePackages returns an empty slice and nil error.
- Each pkgNode.PopulatePackageInformation() populates imports, type definitions, and function declarations
  for its corresponding package; any error short-circuits and is propagated.
- Assumes alignment between syntax files and compiled Go files within each package when populating data.

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
