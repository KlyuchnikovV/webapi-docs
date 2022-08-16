package fun

import (
	"go/ast"
)

type Body struct {
	Statements *ast.BlockStmt
}

func NewBody(statements *ast.BlockStmt) *Body {
	return &Body{
		Statements: statements,
	}
}

func (b *Body) ReturnStatements() []ast.ReturnStmt {
	var result = make([]ast.ReturnStmt, 0)

	ast.Inspect(b.Statements, func(n ast.Node) bool {
		ret, ok := n.(*ast.ReturnStmt)
		if ok {
			result = append(result, *ret)
		}
		return true
	})

	return result
}
