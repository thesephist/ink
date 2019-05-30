package main

import (
	"log"
	"math"
)

const (
	maxIdx = math.MaxInt32
)

func Parse(tokenStream <-chan Tok, nodes chan<- Node, done chan<- bool) {
	tokens := make([]Tok, 0)
	for tok := range tokenStream {
		// log.Println(tok)
		tokens = append(tokens, tok)
	}

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

func parseExpression(tokens []Tok) (Node, int) {
	idx := 0

	consumeDanglingSeparator := func() {
		// bounds check in case parseExpress() called at some point
		//	consumed end token
		if idx < len(tokens) && tokens[idx].kind == Separator {
			idx++
		}
	}

	if tokens[0].kind == NegationOp {
		atom, idx := parseAtom(tokens[1:])
		consumeDanglingSeparator()
		return UnaryExprNode{tokens[0], atom}, idx
	}

	atom, incr := parseAtom(tokens[idx:])
	idx += incr

	nextTok := tokens[idx]
	idx++

	switch nextTok.kind {
	case Separator:
		return atom, idx // consuming separator
	case KeyValueSeparator:
		return atom, idx - 1 // not consuming KeyValueSeparator
	case AddOp, SubtractOp, MultiplyOp, DivideOp, ModulusOp,
		GreaterThanOp, LessThanOp, EqualOp, EqRefOp, DefineOp, AccessorOp:
		rightOperand, incr := parseAtom(tokens[idx:])
		idx += incr
		// TODO: implement order of operations.
		// 	'.' > '%' > '*/' > '+-' > everything else
		consumeDanglingSeparator()
		return BinaryExprNode{
			nextTok,
			atom,
			rightOperand,
		}, idx
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
			atom,
			clauses,
		}, idx
	default:
		log.Fatalf("syntax error: unexpected token in  expression with %s", tokens[idx])
		consumeDanglingSeparator()
		return nil, maxIdx
	}
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

func parseAtom(tokens []Tok) (Node, int) {
	tok, idx := tokens[0], 1

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
		} else {
			atom = IdentifierNode{tok.stringVal()}
		}
		// may be called as a function, so flows beyodn
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
		} else {
			atom = ExpressionListNode{exprs}
		}
		// may be called as a function, so flows beyodn
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
			entries = append(entries, ObjectEntryNode{keyExpr, valExpr})
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
		log.Fatalf("syntax error: unexpected start of atom, found %s", tok)
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
		log.Fatalf("expected CaseArrow, but got %s", tokens[idx].String())
	}
	idx++

	expr, incr := parseExpression(tokens[idx:])
	idx += incr

	return MatchClauseNode{
		atom,
		expr,
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
				log.Fatalf("invalid syntax: expected a comma separated arguments list, found %s",
					tokens[idx].String())
			}
			idx++ // Separator
		}
		if tokens[idx].kind != RightParen {
			log.Fatalf("invalid syntax: expected arguments list to terminate with a RightParen, found %s",
				tokens[idx].String())
		}
		idx++ // RightParen
	case Identifier:
		idNode := IdentifierNode{tok.stringVal()}
		arguments = append(arguments, idNode)
	default:
		log.Fatalf("invalid syntax: malformed arguments list in function at %s", tok.String())
	}

	if tokens[idx].kind != FunctionArrow {
		log.Fatalf("invalid syntax: expected FunctionArrow but found %s", tokens[idx].String())
	}
	idx++

	body, incr := parseExpression(tokens[idx:])
	idx += incr
	return FunctionLiteralNode{
		arguments,
		body,
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
		function,
		arguments,
	}, idx

}
