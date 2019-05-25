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

	idx, length := 0, len(tokens)
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

type FunctionCallExprNode struct {
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
	tok := tokens[0]
	if tok.kind == NegationOp {
		atom, idx := parseAtom(tokens[1:])
		return UnaryExprNode{
			tok,
			atom,
		}, idx
	}

	atom, idx := parseAtom(tokens)
	defer func() {
		idx++ // clean up separator
	}()

	next := tokens[idx]
	idx++
	switch next.kind {
	case Separator:
		return atom, idx
	case AddOp, SubtractOp, MultiplyOp, DivideOp, ModulusOp,
		GreaterThanOp, LessThanOp, EqualOp, IsOp, DefineOp, AccessorOp:
		rightOperand, idxIncr := parseAtom(tokens[idx:])
		idx += idxIncr
		return BinaryExprNode{
			next,
			atom,
			rightOperand,
		}, idx
	case LeftParen:
		arguments := make([]interface{}, 0)
		for tokens[idx].kind != RightParen {
			exprNode, idxIncr := parseExpression(tokens[idx:])
			idxIncr += idx
			arguments = append(arguments, exprNode)
		}
		return FunctionCallExprNode{
			atom,
			arguments,
		}, idx
	case MatchColon:
		idx++ // assume next token is RightBrace for now
		clauses := make([]MatchClauseNode, 0)
		for tokens[idx].kind != RightBrace {
			clauseNode, idxIncr := parseMatchClause(tokens[idx:])
			idx += idxIncr
			clauses = append(clauses, clauseNode)
		}
		return MatchExprNode{
			atom,
			clauses,
		}, idx
	default:
		log.Fatal("syntax error")
		return []interface{}{}, infty
	}

	log.Fatal("syntax error")
	return []interface{}{}, infty
}

type IdentifierNode struct {
	val string
}

type NumberLiteralNode struct {
	val float64
}

type StringLiteralNode struct {
	val string
}

type ObjectLiteralNode struct {
}

type ListLiteralNode struct {
}

type FunctionLiteralNode struct {
	arguments []IdentifierNode
	body      []interface{}
}

func parseAtom(tokens []Tok) (interface{}, int) {
	tok := tokens[0]
	switch tok.kind {
	case Identifier:
		if tokens[1].kind == FunctionArrow {
			return parseFunctionLiteral(tokens)
		} else {
			return IdentifierNode{tok.stringVal()}, 1
		}
	case NumberLiteral:
		return NumberLiteralNode{tok.numberVal()}, 1
	case StringLiteral:
		return StringLiteralNode{tok.stringVal()}, 1
	case TrueLiteral, FalseLiteral, NullLiteral:
		return tok, 1
	case LeftParen:
		// grouped expression
		return IdentifierNode{}, infty
	case LeftBrace:
		// object literal
		return IdentifierNode{}, infty
	case LeftBracket:
		// array literal
		return IdentifierNode{}, infty
	default:
		// may be function literal
		return IdentifierNode{}, infty
	}
	return IdentifierNode{}, infty
}

func parseMatchClause(tokens []Tok) (MatchClauseNode, int) {
	return MatchClauseNode{}, infty
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
		idx++
		for tokens[idx].kind == Identifier {
			idNode := IdentifierNode{tokens[idx].stringVal()}
			arguments = append(arguments, idNode)
			idx++
		}
		// next is right paren
		idx++
	case Identifier:
		idNode := IdentifierNode{tokens[0].stringVal()}
		arguments = append(arguments, idNode)
		idx++
	default:
		log.Fatal("Invalid syntax")
	}

	if tokens[idx].kind != FunctionArrow {
		log.Fatal("Invalid syntax")
	}
	idx++

	blockResult, idxIncr := parseBlock(tokens[idx:])
	idx += idxIncr
	return FunctionLiteralNode{
		arguments,
		blockResult,
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
