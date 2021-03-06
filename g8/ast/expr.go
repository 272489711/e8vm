package ast

import (
	"e8vm.io/e8vm/lex8"
)

// Operand is an operand expression
type Operand struct {
	*lex8.Token
}

// MemberExpr is an expression of form A.B
type MemberExpr struct {
	Expr Expr
	Dot  *lex8.Token
	Sub  *lex8.Token
}

// OpExpr is a binary or unary operation that uses an operator
type OpExpr struct {
	A  Expr
	Op *lex8.Token
	B  Expr
}

// StarExpr is an expression after a '*'
type StarExpr struct {
	Star *lex8.Token
	Expr Expr
}

// ParenExpr is an expression in a pair of parenthesis
type ParenExpr struct {
	Lparen *lex8.Token
	Expr
	Rparen *lex8.Token
}

// ExprList is a list of expressions
type ExprList struct {
	Exprs  []Expr
	Commas []*lex8.Token
}

// Len returns the length of the expression list
func (list *ExprList) Len() int {
	if list == nil {
		return 0
	}
	return len(list.Exprs)
}

// CallExpr is a function call expression
type CallExpr struct {
	Func   Expr
	Lparen *lex8.Token
	Args   *ExprList
	Rparen *lex8.Token
}

// IndexExpr is fetching an element in an array or slice
type IndexExpr struct {
	Array    Expr
	Lbrack   *lex8.Token
	Index    Expr
	Colon    *lex8.Token
	IndexEnd Expr
	Rbrack   *lex8.Token
}

// ArrayTypeExpr is the type expression of an array or a slice
type ArrayTypeExpr struct {
	Lbrack *lex8.Token
	Len    Expr // optional
	Rbrack *lex8.Token
	Type   Expr
}

// ArrayLiteral is an array or slice literal
type ArrayLiteral struct {
	Type   *ArrayTypeExpr
	Lbrace *lex8.Token
	Exprs  *ExprList
	Rbrace *lex8.Token
}

// FuncTypeExpr is the type expression of a function pointer
type FuncTypeExpr struct {
	Kw      *lex8.Token
	FuncSig *FuncSig
}
