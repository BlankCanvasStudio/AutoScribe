package golang

import (
    "fmt"

    "golang.org/x/tools/go/packages"

    log "github.com/sirupsen/logrus"

    "github.com/BlankCanvasStudio/AutoScribe/pkg/types/mst"
)


func ParsePackages(foldername string) ([]*PackageNode, error) {
    cfg := &packages.Config{
        Mode: packages.NeedName            |
              packages.NeedFiles           |
              packages.NeedSyntax          |
              packages.NeedCompiledGoFiles |
              packages.NeedSyntax          |
              packages.NeedTypes           |
              packages.NeedTypesInfo       |
              packages.NeedImports         |
              packages.NeedDeps,
    }

    pkgs, err := packages.Load(cfg, foldername)
    if err != nil {
        return nil, fmt.Errorf("failed to load package %v: %v", foldername, err)
    }

    pkgNodes := []*PackageNode{}

    log.Debugf("# packages seen: %v", len(pkgs))

    for _, pkg := range pkgs {
        pkgNode := &PackageNode {
            Package: pkg,
            FunctionDecls: []mst.FunctionDecl{},
            // TypeDefinitions:      []*ast.TypeSpec{},
            Imports:              make(map[string]string),
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

