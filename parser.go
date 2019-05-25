package main

import (
	"log"
	"math"
)

const (
	infty = math.MaxInt32
)

func Parse(tokenStream <-chan Tok, nodes chan<- interface{}, done chan<- bool) {
	tokens := make([]Tok, 0)
	for tok := range tokenStream {
		log.Println(tok)
		tokens = append(tokens, tok)
	}

	log.Println()

	idx, length := 0, len(tokens)
	log.Printf("TOKEN COUNT: %d", length)
	for idx < length {
		expr, idxIncr := parseExpression(tokens[idx:])
		idx += idxIncr
		nodes <- expr
	}
	close(nodes)

	done <- true
}

type UnaryExprNode struct {
	operator Tok
	operand  interface{}
}

type BinaryExprNode struct {
	operator     Tok
	leftOperand  interface{}
	rightOperand interface{}
}

type FunctionCallNode struct {
	function  interface{}
	arguments []interface{}
}

type MatchClauseNode struct {
	target      interface{}
	expressions []interface{}
}

type MatchExprNode struct {
	condition interface{}
	clauses   []MatchClauseNode
}

func parseExpression(tokens []Tok) (interface{}, int) {
	idx := 0

	consumeDanglingSeparator := func() {
		if tokens[idx].kind == Separator {
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

	atom, idxIncr := parseAtom(tokens[idx:])
	idx += idxIncr

	log.Println("ATOM =============")
	log.Println(atom)

	next := tokens[idx]
	log.Println(next)
	idx++
	switch next.kind {
	case Separator:
		consumeDanglingSeparator()
		return atom, idx
	case RightParen, RightBracket, RightBrace:
		consumeDanglingSeparator()
		return atom, idx - 1
	case AddOp, SubtractOp, MultiplyOp, DivideOp, ModulusOp,
		GreaterThanOp, LessThanOp, EqualOp, IsOp, DefineOp, AccessorOp:
		rightOperand, idxIncr := parseAtom(tokens[idx:])
		idx += idxIncr
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
			clauseNode, idxIncr := parseMatchClause(tokens[idx:])
			idx += idxIncr
			clauses = append(clauses, clauseNode)
		}
		idx++ // RightBrace
		consumeDanglingSeparator()
		return MatchExprNode{
			atom,
			clauses,
		}, idx
	default:
		log.Fatal("syntax error: unexpected end of expression")
		consumeDanglingSeparator()
		return []interface{}{}, infty
	}

	log.Fatal("syntax error: unexpected end of expression")
	consumeDanglingSeparator()
	return []interface{}{}, infty
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

type ObjectLiteralNode struct{}

type ListLiteralNode struct{}

type FunctionLiteralNode struct {
	arguments []IdentifierNode
	body      []interface{}
}

func parseAtom(tokens []Tok) (interface{}, int) {
	tok := tokens[0]
	switch tok.kind {
	case EmptyIdentifier:
		return EmptyIdentifierNode{}, 1
	case Identifier:
		var atom interface{}
		var idx int
		if tokens[1].kind == FunctionArrow {
			atom, idx = parseFunctionLiteral(tokens)
		} else {
			atom, idx = IdentifierNode{tok.stringVal()}, 1
		}
		if tokens[idx].kind == LeftParen {
			// may be a function call
			fnCall, idxIncr := parseFunctionCall(atom, tokens[idx:])
			idx += idxIncr
			return fnCall, idx
		} else {
			return atom, idx
		}
	case NumberLiteral:
		return NumberLiteralNode{tok.numberVal()}, 1
	case StringLiteral:
		return StringLiteralNode{tok.stringVal()}, 1
	case TrueLiteral, FalseLiteral, NullLiteral:
		return tok, 1
	case LeftParen:
		// grouped expression or function literal
		idx := 1 // LeftParen
		var expr interface{}
		var idxIncr int
		for tokens[idx].kind != RightParen {
			expr, idxIncr = parseExpression(tokens[idx:])
			idx += idxIncr
		}
		idx++ // RightParen
		if tokens[idx].kind == FunctionArrow {
			expr, idxIncr = parseFunctionLiteral(tokens)
			idx += idxIncr
		}
		if tokens[idx].kind == LeftParen {
			fnCall, idxIncr := parseFunctionCall(expr, tokens[idx:])
			idx += idxIncr
			return fnCall, idx
		} else {
			return expr, idx
		}
	case LeftBrace:
		// object literal
		log.Fatal("syntax error: atom::LeftBrace not implemented")
		return IdentifierNode{}, infty
	case LeftBracket:
		// array literal
		log.Fatal("syntax error: atom::LeftBracket not implemented")
		return IdentifierNode{}, infty
	}

	log.Fatalf("syntax error: unexpected end of atom, found %s", tok)
	return IdentifierNode{}, infty
}

func parseMatchClause(tokens []Tok) (MatchClauseNode, int) {
	idx := 0
	atom, idxIncr := parseAtom(tokens)
	idx += idxIncr

	if tokens[idx].kind != CaseArrow {
		log.Fatalf("expected CaseArrow, but got %s", tokens[idx])
	}
	idx++

	block, idxIncr := parseBlock(tokens[idx:])
	idx += idxIncr

	return MatchClauseNode{
		atom,
		block,
	}, idx
}

func parseObjectLiteral(tokens []Tok) (interface{}, int) {
	return ObjectLiteralNode{}, infty
}

func parseListLiteral(tokens []Tok) (interface{}, int) {
	return ListLiteralNode{}, infty
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

	blockResult, idxIncr := parseBlock(tokens[idx:])
	idx += idxIncr
	return FunctionLiteralNode{
		arguments,
		blockResult,
	}, idx
}

func parseFunctionCall(function interface{}, tokens []Tok) (FunctionCallNode, int) {
	idx := 1 // LeftParen
	arguments := make([]interface{}, 0)
	for tokens[idx].kind != RightParen {
		exprNode, idxIncr := parseExpression(tokens[idx:])
		idx += idxIncr
		arguments = append(arguments, exprNode)
	}
	idx++ // RightParen
	return FunctionCallNode{
		function,
		arguments,
	}, idx

}

func parseBlock(tokens []Tok) ([]interface{}, int) {
	idx := 0
	expressions := make([]interface{}, 0)
	switch tokens[0].kind {
	case LeftBrace:
		// parse until rlight brace
		idx++ // left brace
		for tokens[idx].kind != RightBrace {
			exprResult, idxIncr := parseExpression(tokens[idx:])
			idx += idxIncr
			expressions = append(expressions, exprResult)
		}
		idx++ // right brace
	default:
		var exprResult interface{}
		exprResult, idx = parseExpression(tokens)
		expressions = append(expressions, exprResult)
	}

	return expressions, idx
}
