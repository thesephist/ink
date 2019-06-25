package main

import (
	"math"
)

const (
	maxIdx = math.MaxInt32
)

func Parse(tokenStream <-chan Tok, nodes chan<- Node, debugParser bool) {
	tokens := make([]Tok, 0)
	for tok := range tokenStream {
		tokens = append(tokens, tok)
	}

	idx, length := 0, len(tokens)
	for idx < length {
		expr, incr := parseExpression(tokens[idx:])
		idx += incr
		if debugParser {
			logDebug("parse ->", expr.String())
		}
		nodes <- expr
	}
	close(nodes)
}

type UnaryExprNode struct {
	operator Tok
	operand  Node
}

type BinaryExprNode struct {
	operator     Tok
	leftOperand  Node
	rightOperand Node
}

type FunctionCallNode struct {
	function  Node
	arguments []Node
}

type MatchClauseNode struct {
	target     Node
	expression Node
}

type MatchExprNode struct {
	condition Node
	clauses   []MatchClauseNode
}

type ExpressionListNode struct {
	expressions []Node
}

type EmptyIdentifierNode struct{}

type IdentifierNode struct {
	val string
}

type NumberLiteralNode struct {
	val float64
}

type StringLiteralNode struct {
	val string
}

type BooleanLiteralNode struct {
	val bool
}

type NullLiteralNode struct{}

type ObjectLiteralNode struct {
	entries []ObjectEntryNode
}

type ObjectEntryNode struct {
	key Node
	val Node
}

type ListLiteralNode struct {
	vals []Node
}

type FunctionLiteralNode struct {
	arguments []IdentifierNode
	body      Node
}

func getOpPriority(t Tok) int {
	// higher == greater priority
	switch t.kind {
	case AccessorOp:
		return 100
	case ModulusOp:
		return 80
	case MultiplyOp, DivideOp:
		return 50
	case AddOp, SubtractOp:
		return 20
	case GreaterThanOp, LessThanOp, EqualOp, EqRefOp:
		return 10
	case DefineOp:
		return 0
	default:
		return -1
	}
}

func isBinaryOp(t Tok) bool {
	switch t.kind {
	case AddOp, SubtractOp, MultiplyOp, DivideOp, ModulusOp,
		GreaterThanOp, LessThanOp, EqualOp, EqRefOp, DefineOp, AccessorOp:
		return true
	default:
		return false
	}
}

func parseBinaryExpression(
	leftOperand Node,
	operator Tok,
	tokens []Tok,
	previousPriority int,
) (Node, int) {
	rightAtom, idx := parseAtom(tokens)
	incr := 0

	ops := make([]Tok, 1)
	nodes := make([]Node, 2)
	ops[0] = operator
	nodes[0] = leftOperand
	nodes[1] = rightAtom

	// build up a list of binary operations, with tree nodes
	//	where there are higher-priority binary ops
	for len(tokens) > idx && isBinaryOp(tokens[idx]) {
		if previousPriority >= getOpPriority(tokens[idx]) {
			// Priority is lower than the calling function's last op,
			//  so return control to the parent binary op
			break
		} else if getOpPriority(ops[len(ops)-1]) >= getOpPriority(tokens[idx]) {
			// Priority is lower than the previous op (but higher than parent),
			//	so it's ok to be left-heavy in this tree
			ops = append(ops, tokens[idx])
			idx++
			rightAtom, incr = parseAtom(tokens[idx:])
			nodes = append(nodes, rightAtom)
			idx += incr
		} else {
			// Priority is higher than previous ops,
			//	so make it a right-heavy tree
			subtree, incr := parseBinaryExpression(
				nodes[len(nodes)-1],
				tokens[idx],
				tokens[idx+1:],
				getOpPriority(ops[len(ops)-1]),
			)
			nodes[len(nodes)-1] = subtree
			idx += incr + 1
		}
	}

	// ops, nodes -> left-biased binary expression tree
	tree := nodes[0]
	nodes = nodes[1:]
	for len(ops) > 0 {
		tree = BinaryExprNode{
			operator:     ops[0],
			leftOperand:  tree,
			rightOperand: nodes[0],
		}
		ops = ops[1:]
		nodes = nodes[1:]
	}

	return tree, idx
}

func parseExpression(tokens []Tok) (Node, int) {
	idx := 0

	consumeDanglingSeparator := func() {
		// bounds check in case parseExpress() called at some point
		//	consumed end token
		if idx < len(tokens) && tokens[idx].kind == Separator {
			idx++
		}
	}

	atom, incr := parseAtom(tokens[idx:])
	idx += incr

	nextTok := tokens[idx]
	idx++

	switch nextTok.kind {
	case Separator:
		return atom, idx // consuming dangling separator
	case KeyValueSeparator, RightParen:
		// these belong to the parent atom that contains this expression,
		//	so return without consuming token (idx - 1)
		return atom, idx - 1
	case AddOp, SubtractOp, MultiplyOp, DivideOp, ModulusOp,
		GreaterThanOp, LessThanOp, EqualOp, EqRefOp, DefineOp, AccessorOp:
		n, incr := parseBinaryExpression(atom, nextTok, tokens[idx:], -1)
		idx += incr
		consumeDanglingSeparator()
		return n, idx
	case MatchColon:
		idx++ // LeftBrace
		clauses := make([]MatchClauseNode, 0)
		for tokens[idx].kind != RightBrace {
			clauseNode, incr := parseMatchClause(tokens[idx:])
			idx += incr
			clauses = append(clauses, clauseNode)
		}
		idx++ // RightBrace
		consumeDanglingSeparator()
		return MatchExprNode{
			condition: atom,
			clauses:   clauses,
		}, idx
	default:
		logErrf(ErrSyntax, "unexpected token %s following an expression", nextTok)
		consumeDanglingSeparator()
		return nil, maxIdx
	}
}

func parseAtom(tokens []Tok) (Node, int) {
	tok, idx := tokens[0], 1

	if tok.kind == NegationOp {
		atom, idx := parseAtom(tokens[1:])
		return UnaryExprNode{
			operator: tok,
			operand:  atom,
		}, idx + 1
	}

	var atom Node
	switch tok.kind {
	case EmptyIdentifier:
		return EmptyIdentifierNode{}, idx
	case NumberLiteral:
		return NumberLiteralNode{tok.numberVal()}, idx
	case StringLiteral:
		return StringLiteralNode{tok.stringVal()}, idx
	case TrueLiteral:
		return BooleanLiteralNode{true}, idx
	case FalseLiteral:
		return BooleanLiteralNode{false}, idx
	case NullLiteral:
		return NullLiteralNode{}, idx
	case Identifier:
		if tokens[idx].kind == FunctionArrow {
			atom, idx = parseFunctionLiteral(tokens)
			// parseAtom should not consume trailing Separators, but
			// 	parseFunctionLiteral does because it ends with expressions.
			// 	so we backtrack one token.
			idx--
		} else {
			atom = IdentifierNode{tok.stringVal()}
		}
		// may be called as a function, so flows beyond
		//	switch case
	case LeftParen:
		// grouped expression or function literal
		exprs := make([]Node, 0)
		for tokens[idx].kind != RightParen {
			expr, incr := parseExpression(tokens[idx:])
			idx += incr
			exprs = append(exprs, expr)
		}
		idx++ // RightParen
		if tokens[idx].kind == FunctionArrow {
			atom, idx = parseFunctionLiteral(tokens)
			// parseAtom should not consume trailing Separators, but
			// 	parseFunctionLiteral does because it ends with expressions.
			// 	so we backtrack one token.
			idx--
		} else {
			atom = ExpressionListNode{exprs}
		}
		// may be called as a function, so flows beyond
		//	switch case
	case LeftBrace:
		entries := make([]ObjectEntryNode, 0)
		for tokens[idx].kind != RightBrace {
			keyExpr, keyIncr := parseExpression(tokens[idx:])
			idx += keyIncr
			idx++ // KeyValueSeparator
			valExpr, valIncr := parseExpression(tokens[idx:])
			// Separator consumed by parseExpression
			idx += valIncr
			entries = append(entries, ObjectEntryNode{key: keyExpr, val: valExpr})
		}
		idx++ // RightBrace
		return ObjectLiteralNode{entries}, idx
	case LeftBracket:
		vals := make([]Node, 0)
		for tokens[idx].kind != RightBracket {
			expr, incr := parseExpression(tokens[idx:])
			idx += incr
			vals = append(vals, expr)
		}
		idx++ // RightBracket
		return ListLiteralNode{vals}, idx
	default:
		logErrf(ErrSyntax, "unexpected start of atom, found %s", tok)
		return IdentifierNode{}, maxIdx
	}

	// bounds check here because parseExpression may have
	//	consumed all tokens before this
	for idx < len(tokens) && tokens[idx].kind == LeftParen {
		var incr int
		atom, incr = parseFunctionCall(atom, tokens[idx:])
		idx += incr
	}

	return atom, idx
}

func parseMatchClause(tokens []Tok) (MatchClauseNode, int) {
	atom, idx := parseAtom(tokens)

	if tokens[idx].kind != CaseArrow {
		logErrf(ErrSyntax, "expected token '->', but got %s", tokens[idx].String())
	}
	idx++

	expr, incr := parseExpression(tokens[idx:])
	idx += incr

	return MatchClauseNode{
		target:     atom,
		expression: expr,
	}, idx
}

func parseFunctionLiteral(tokens []Tok) (FunctionLiteralNode, int) {
	tok, idx := tokens[0], 1
	arguments := make([]IdentifierNode, 0)

	switch tok.kind {
	case LeftParen:
		for tokens[idx].kind == Identifier {
			idNode := IdentifierNode{tokens[idx].stringVal()}
			arguments = append(arguments, idNode)
			idx++

			if tokens[idx].kind != Separator {
				logErrf(ErrSyntax, "expected a comma separated arguments list, found %s",
					tokens[idx].String())
			}
			idx++ // Separator
		}
		if tokens[idx].kind != RightParen {
			logErrf(ErrSyntax, "expected arguments list to terminate with a RightParen, found %s",
				tokens[idx].String())
		}
		idx++ // RightParen
	case Identifier:
		idNode := IdentifierNode{tok.stringVal()}
		arguments = append(arguments, idNode)
	default:
		logErrf(ErrSyntax, "malformed arguments list in function at %s", tok.String())
	}

	if tokens[idx].kind != FunctionArrow {
		logErrf(ErrSyntax, "expected FunctionArrow but found %s", tokens[idx].String())
	}
	idx++

	body, incr := parseExpression(tokens[idx:])
	idx += incr
	return FunctionLiteralNode{
		arguments: arguments,
		body:      body,
	}, idx
}

func parseFunctionCall(function Node, tokens []Tok) (FunctionCallNode, int) {
	idx := 1
	arguments := make([]Node, 0)
	for tokens[idx].kind != RightParen {
		expr, incr := parseExpression(tokens[idx:])
		idx += incr
		arguments = append(arguments, expr)
	}
	idx++ // RightParen
	return FunctionCallNode{
		function:  function,
		arguments: arguments,
	}, idx

}
