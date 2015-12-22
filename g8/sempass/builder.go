package sempass

import (
	"e8vm.io/e8vm/g8/ast"
	"e8vm.io/e8vm/g8/tast"
	"e8vm.io/e8vm/lex8"
	"e8vm.io/e8vm/sym8"
)

type builder struct {
	*lex8.ErrorList
	path string

	this  *tast.Ref
	scope *sym8.Scope

	exprFunc func(b *builder, expr ast.Expr) tast.Expr
	stmtFunc func(b *builder, stmt ast.Stmt) tast.Stmt
}
