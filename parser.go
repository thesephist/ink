package main

import (
	"log"
	"math"
)

const (
	infty = math.MaxInt32
)

func Parse(tokenStream <-chan Tok, nodes chan<- Node, done chan<- bool) {
	tokens := make([]Tok, 0)
	for tok := range tokenStream {
		log.Println(tok)
		tokens = append(tokens, tok)
	}

	log.Println()

	idx, length := 0, len(tokens)
	for idx < length {
		expr, incr := parseExpression(tokens[idx:])
		idx += incr
		nodes <- expr
	}
	close(nodes)

	done <- true
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
	target      Node
	expressions []Node
}

type MatchExprNode struct {
	condition Node
	clauses   []MatchClauseNode
}

func parseExpression(tokens []Tok) (Node, int) {
	idx := 0

	consumeDanglingSeparator := func() {
		if idx < len(tokens) && tokens[idx].kind == Separator {
			idx++
		}
	}

	switch tokens[0].kind {
	case NegationOp:
		atom, idx := parseAtom(tokens[1:])
		consumeDanglingSeparator()
		return UnaryExprNode{
			tokens[0],
			atom,
		}, idx
	case Separator:
		idx++
	}

	atom, incr := parseAtom(tokens[idx:])
	idx += incr

	next := tokens[idx]
	idx++
	switch next.kind {
	case Separator:
		// TODO: is this call to consumeDanglingSeparator() necessary?
		consumeDanglingSeparator()
		return atom, idx
	case KeyValueSeparator:
		// we don't consume the ':' because it's not
		// 	implicit in an expression
		return atom, idx - 1
	case RightParen, RightBracket, RightBrace:
		consumeDanglingSeparator()
		return atom, idx - 1
	case AddOp, SubtractOp, MultiplyOp, DivideOp, ModulusOp,
		GreaterThanOp, LessThanOp, EqualOp, EqRefOp, DefineOp, AccessorOp:
		rightOperand, incr := parseAtom(tokens[idx:])
		idx += incr
		// TODO: impl order of operations.
		// 	'.' > '%' > '*/' > '+-' > everything else
		consumeDanglingSeparator()
		return BinaryExprNode{
			next,
			atom,
			rightOperand,
		}, idx
	case MatchColon:
		idx++ // assume next token is LeftBrace for now
		clauses := make([]MatchClauseNode, 0)
		for tokens[idx].kind != RightBrace {
			clauseNode, incr := parseMatchClause(tokens[idx:])
			idx += incr
			clauses = append(clauses, clauseNode)
		}
		idx++ // RightBrace
		consumeDanglingSeparator()
		return MatchExprNode{
			atom,
			clauses,
		}, idx
	default:
		log.Fatalf("syntax error: unexpected end of expression with %s", tokens[idx])
		consumeDanglingSeparator()
		return nil, infty
	}

	log.Fatalf("syntax error: unexpected end of expression with %s", tokens[idx])
	consumeDanglingSeparator()
	return nil, infty
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
	body      []Node
}

func parseAtom(tokens []Tok) (Node, int) {
	tok := tokens[0]
	switch tok.kind {
	case EmptyIdentifier:
		return EmptyIdentifierNode{}, 1
	case Identifier:
		var atom Node
		var idx int
		if tokens[1].kind == FunctionArrow {
			atom, idx = parseFunctionLiteral(tokens)
		} else {
			atom, idx = IdentifierNode{tok.stringVal()}, 1
		}
		if idx < len(tokens) && tokens[idx].kind == LeftParen {
			// may be a function call
			fnCall, incr := parseFunctionCall(atom, tokens[idx:])
			idx += incr
			return fnCall, idx
		} else {
			return atom, idx
		}
	case NumberLiteral:
		return NumberLiteralNode{tok.numberVal()}, 1
	case StringLiteral:
		return StringLiteralNode{tok.stringVal()}, 1
	case TrueLiteral:
		return BooleanLiteralNode{true}, 1
	case FalseLiteral:
		return BooleanLiteralNode{false}, 1
	case NullLiteral:
		return NullLiteralNode{}, 1
	case LeftParen:
		// grouped expression or function literal
		idx := 1 // LeftParen
		var expr Node
		var incr int
		for tokens[idx].kind != RightParen {
			expr, incr = parseExpression(tokens[idx:])
			idx += incr
		}
		idx++ // RightParen
		if tokens[idx].kind == FunctionArrow {
			expr, incr = parseFunctionLiteral(tokens)
			// we assign instead of incrementing here because
			// 	we started from index 0
			idx = incr
		}
		if idx < len(tokens) && tokens[idx].kind == LeftParen {
			fnCall, incr := parseFunctionCall(expr, tokens[idx:])
			idx += incr
			return fnCall, idx
		} else {
			return expr, idx
		}
	case LeftBrace:
		idx := 1
		entries := make([]ObjectEntryNode, 0)
		for tokens[idx].kind != RightBrace {
			keyExpr, keyIncr := parseExpression(tokens[idx:])
			idx += keyIncr
			idx++ // KeyValueSeparator
			valExpr, valIncr := parseExpression(tokens[idx:])
			// Separator consumed by parseExpression
			idx += valIncr
			entries = append(entries, ObjectEntryNode{keyExpr, valExpr})
		}
		idx++ // RightBrace
		return ObjectLiteralNode{entries}, idx
	case LeftBracket:
		idx := 1 // LeftBracket
		vals := make([]Node, 0)
		for tokens[idx].kind != RightBracket {
			expr, incr := parseExpression(tokens[idx:])
			idx += incr
			vals = append(vals, expr)
		}
		idx++ // RightBracket
		return ListLiteralNode{vals}, idx
	}

	log.Fatalf("syntax error: unexpected end of atom, found %s", tok)
	return IdentifierNode{}, infty
}

func parseMatchClause(tokens []Tok) (MatchClauseNode, int) {
	idx := 0
	atom, incr := parseAtom(tokens)
	idx += incr

	if tokens[idx].kind != CaseArrow {
		log.Fatalf("expected CaseArrow, but got %s", tokens[idx])
	}
	idx++

	block, incr := parseBlock(tokens[idx:])
	idx += incr

	return MatchClauseNode{
		atom,
		block,
	}, idx
}

func parseObjectLiteral(tokens []Tok) (Node, int) {
	return ObjectLiteralNode{}, infty
}

func parseFunctionLiteral(tokens []Tok) (FunctionLiteralNode, int) {
	idx := 0
	arguments := make([]IdentifierNode, 0)
	switch tokens[0].kind {
	case LeftParen:
		idx++ // LeftParen
		for tokens[idx].kind != RightParen {
			if tokens[idx].kind == Identifier {
				idNode := IdentifierNode{tokens[idx].stringVal()}
				arguments = append(arguments, idNode)
				idx++
			} else {
				log.Fatal("invalid syntax: expected an identifier in arguments list")
			}

			if tokens[idx].kind != Separator {
				log.Fatal("invalid syntax: expected a comma separated arguments list")
			}
			idx++ // Separator
		}

		idx++ // RightParen
	case Identifier:
		idNode := IdentifierNode{tokens[0].stringVal()}
		arguments = append(arguments, idNode)
		idx++
	default:
		log.Fatal("invalid syntax: malformed arguments list in function")
	}

	if tokens[idx].kind != FunctionArrow {
		log.Fatal("invalid syntax: expected FunctionArrow")
	}
	idx++

	blockResult, incr := parseBlock(tokens[idx:])
	idx += incr
	return FunctionLiteralNode{
		arguments,
		blockResult,
	}, idx
}

func parseFunctionCall(function Node, tokens []Tok) (FunctionCallNode, int) {
	idx := 1 // LeftParen
	arguments := make([]Node, 0)
	for tokens[idx].kind != RightParen {
		exprNode, incr := parseExpression(tokens[idx:])
		idx += incr
		arguments = append(arguments, exprNode)
	}
	idx++ // RightParen
	return FunctionCallNode{
		function,
		arguments,
	}, idx

}

func parseBlock(tokens []Tok) ([]Node, int) {
	idx := 0
	expressions := make([]Node, 0)
	switch tokens[0].kind {
	case LeftBrace:
		// parse until rlight brace
		idx++ // left brace
		for tokens[idx].kind != RightBrace {
			exprResult, incr := parseExpression(tokens[idx:])
			idx += incr
			expressions = append(expressions, exprResult)
		}
		idx++ // right brace
		if tokens[idx].kind == Separator {
			idx++
		}
	default:
		var exprResult Node
		exprResult, idx = parseExpression(tokens)
		expressions = append(expressions, exprResult)
	}

	return expressions, idx
}
