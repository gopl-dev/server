// Package main implements a custom analyzer that checks for the usage of the 'call' wrapper
// in public methods of the Service struct, ensuring consistent validation via 'Normalize'.
package main

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(Analyzer)
}

var Analyzer = &analysis.Analyzer{
	Name: "service_call_guard",
	Doc:  "enforces the use of 'Normalize' in public Service methods",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			fn, ok := n.(*ast.FuncDecl)
			if !ok {
				return true
			}

			if fn.Recv == nil || len(fn.Recv.List) == 0 {
				return true
			}

			recvType := fn.Recv.List[0].Type
			starExpr, ok := recvType.(*ast.StarExpr)
			if !ok {
				return true
			}
			ident, ok := starExpr.X.(*ast.Ident)
			if !ok || ident.Name != "Service" {
				return true
			}

			if !fn.Name.IsExported() {
				return true
			}

			hasNormalize := false
			ast.Inspect(fn.Body, func(n ast.Node) bool {
				callExpr, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}

				var funcName string
				if fun, ok := callExpr.Fun.(*ast.Ident); ok {
					funcName = fun.Name
				}

				if funcName == "Normalize" {
					hasNormalize = true
					return false
				}
				return true
			})

			if !hasNormalize {
				pass.Reportf(fn.Pos(), "public Service method %s must call 'Normalize' for input validation", fn.Name.Name)
			}

			return true
		})
	}

	return nil, nil //nolint:nilnil
}
