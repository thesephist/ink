package main

import (
	"fmt"
)

func guardUnexpectedInputEnd(tokens []Tok, idx int) error {
	if idx >= len(tokens) {
		if len(tokens) > 0 {
			return Err{
				ErrSyntax,
				fmt.Sprintf("unexpected end of input at %s", tokens[len(tokens)-1].String()),
			}
		} else {
			return Err{
				ErrSyntax,
				fmt.Sprintf("unexpected end of input"),
			}
		}
	}

	return nil
}

func Parse(
	tokenStream <-chan Tok,
	nodes chan<- Node,
	errors chan<- Err,
	debugParser bool,
) {
	tokens := make([]Tok, 0)
	for tok := range tokenStream {
		tokens = append(tokens, tok)
	}

	idx, length := 0, len(tokens)
	for idx < length {
		expr, incr, err := parseExpression(tokens[idx:])
		idx += incr

		if err != nil {
			e, isErr := err.(Err)
			if isErr {
				errors <- e
			} else {
				logErrf(ErrAssert, "err raised that was not of Err type -> %s",
					err.Error())
			}
			close(nodes)
			close(errors)
			return
		}

		if debugParser {
			logDebug("parse ->", expr.String())
		}
		nodes <- expr
	}
	close(nodes)
	close(errors)
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
) (Node, int, error) {
	rightAtom, idx, err := parseAtom(tokens)
	if err != nil {
		return nil, 0, err
	}
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

			err := guardUnexpectedInputEnd(tokens, idx)
			if err != nil {
				return nil, 0, err
			}

			rightAtom, incr, err = parseAtom(tokens[idx:])
			if err != nil {
				return nil, 0, err
			}
			nodes = append(nodes, rightAtom)
			idx += incr
		} else {
			err := guardUnexpectedInputEnd(tokens, idx+1)
			if err != nil {
				return nil, 0, err
			}

			// Priority is higher than previous ops,
			//	so make it a right-heavy tree
			subtree, incr, err := parseBinaryExpression(
				nodes[len(nodes)-1],
				tokens[idx],
				tokens[idx+1:],
				getOpPriority(ops[len(ops)-1]),
			)
			if err != nil {
				return nil, 0, err
			}
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

	return tree, idx, nil
}

func parseExpression(tokens []Tok) (Node, int, error) {
	idx := 0

	consumeDanglingSeparator := func() {
		// bounds check in case parseExpress() called at some point
		//	consumed end token
		if idx < len(tokens) && tokens[idx].kind == Separator {
			idx++
		}
	}

	atom, incr, err := parseAtom(tokens[idx:])
	if err != nil {
		return nil, 0, err
	}
	idx += incr

	err = guardUnexpectedInputEnd(tokens, idx)
	if err != nil {
		return nil, 0, err
	}
	nextTok := tokens[idx]
	idx++

	switch nextTok.kind {
	case Separator:
		// consuming dangling separator
		return atom, idx, nil
	case KeyValueSeparator, RightParen:
		// these belong to the parent atom that contains this expression,
		//	so return without consuming token (idx - 1)
		return atom, idx - 1, nil
	case AddOp, SubtractOp, MultiplyOp, DivideOp, ModulusOp,
		GreaterThanOp, LessThanOp, EqualOp, EqRefOp, DefineOp, AccessorOp:
		n, incr, err := parseBinaryExpression(atom, nextTok, tokens[idx:], -1)
		if err != nil {
			return nil, 0, err
		}

		idx += incr
		consumeDanglingSeparator()
		return n, idx, nil
	case MatchColon:
		idx++ // LeftBrace
		clauses := make([]MatchClauseNode, 0)
		err := guardUnexpectedInputEnd(tokens, idx)
		if err != nil {
			return nil, 0, err
		}
		for tokens[idx].kind != RightBrace {
			clauseNode, incr, err := parseMatchClause(tokens[idx:])
			if err != nil {
				return nil, 0, err
			}
			idx += incr

			clauses = append(clauses, clauseNode)

			err = guardUnexpectedInputEnd(tokens, idx)
			if err != nil {
				return nil, 0, err
			}
		}
		idx++ // RightBrace
		consumeDanglingSeparator()
		return MatchExprNode{
			condition: atom,
			clauses:   clauses,
		}, idx, nil
	default:
		return nil, 0, Err{
			ErrSyntax,
			fmt.Sprintf("unexpected token %s following an expression", nextTok),
		}
	}
}

func parseAtom(tokens []Tok) (Node, int, error) {
	err := guardUnexpectedInputEnd(tokens, 0)
	if err != nil {
		return nil, 0, err
	}

	tok, idx := tokens[0], 1

	if tok.kind == NegationOp {
		atom, idx, err := parseAtom(tokens[idx:])
		if err != nil {
			return nil, 0, err
		}
		return UnaryExprNode{
			operator: tok,
			operand:  atom,
		}, idx + 1, nil
	}

	err = guardUnexpectedInputEnd(tokens, idx)
	if err != nil {
		return nil, 0, err
	}

	var atom Node
	switch tok.kind {
	case EmptyIdentifier:
		return EmptyIdentifierNode{}, idx, nil
	case NumberLiteral:
		return NumberLiteralNode{tok.numberVal()}, idx, nil
	case StringLiteral:
		return StringLiteralNode{tok.stringVal()}, idx, nil
	case TrueLiteral:
		return BooleanLiteralNode{true}, idx, nil
	case FalseLiteral:
		return BooleanLiteralNode{false}, idx, nil
	case NullLiteral:
		return NullLiteralNode{}, idx, nil
	case Identifier:
		if tokens[idx].kind == FunctionArrow {
			var err error
			atom, idx, err = parseFunctionLiteral(tokens)
			if err != nil {
				return nil, 0, err
			}

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
			expr, incr, err := parseExpression(tokens[idx:])
			if err != nil {
				return nil, 0, err
			}

			idx += incr
			exprs = append(exprs, expr)

			err = guardUnexpectedInputEnd(tokens, idx)
			if err != nil {
				return nil, 0, err
			}
		}
		idx++ // RightParen

		err = guardUnexpectedInputEnd(tokens, idx)
		if err != nil {
			return nil, 0, err
		}

		if tokens[idx].kind == FunctionArrow {
			var err error
			atom, idx, err = parseFunctionLiteral(tokens)
			if err != nil {
				return nil, 0, err
			}

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
			keyExpr, keyIncr, err := parseExpression(tokens[idx:])
			if err != nil {
				return nil, 0, err
			}

			idx += keyIncr
			idx++ // KeyValueSeparator

			err = guardUnexpectedInputEnd(tokens, idx)
			if err != nil {
				return nil, 0, err
			}

			valExpr, valIncr, err := parseExpression(tokens[idx:])
			if err != nil {
				return nil, 0, err
			}

			// Separator consumed by parseExpression
			idx += valIncr
			entries = append(entries, ObjectEntryNode{key: keyExpr, val: valExpr})

			err = guardUnexpectedInputEnd(tokens, idx)
			if err != nil {
				return nil, 0, err
			}
		}
		idx++ // RightBrace
		return ObjectLiteralNode{entries}, idx, nil
	case LeftBracket:
		vals := make([]Node, 0)
		for tokens[idx].kind != RightBracket {
			expr, incr, err := parseExpression(tokens[idx:])
			if err != nil {
				return nil, 0, err
			}

			idx += incr
			vals = append(vals, expr)

			err = guardUnexpectedInputEnd(tokens, idx)
			if err != nil {
				return nil, 0, err
			}
		}
		idx++ // RightBracket
		return ListLiteralNode{vals}, idx, nil
	default:
		return IdentifierNode{}, 0, Err{
			ErrSyntax,
			fmt.Sprintf("unexpected start of atom, found %s", tok),
		}
	}

	// bounds check here because parseExpression may have
	//	consumed all tokens before this
	for idx < len(tokens) && tokens[idx].kind == LeftParen {
		var incr int
		var err error
		atom, incr, err = parseFunctionCall(atom, tokens[idx:])
		if err != nil {
			return nil, 0, err
		}
		idx += incr

		err = guardUnexpectedInputEnd(tokens, idx)
		if err != nil {
			return nil, 0, err
		}
	}

	return atom, idx, nil
}

func parseMatchClause(tokens []Tok) (MatchClauseNode, int, error) {
	atom, idx, err := parseAtom(tokens)
	if err != nil {
		return MatchClauseNode{}, 0, err
	}

	err = guardUnexpectedInputEnd(tokens, idx)
	if err != nil {
		return MatchClauseNode{}, 0, err
	}

	if tokens[idx].kind != CaseArrow {
		return MatchClauseNode{}, 0, Err{
			ErrSyntax,
			fmt.Sprintf("expected token '->', but got %s", tokens[idx].String()),
		}
	}
	idx++

	err = guardUnexpectedInputEnd(tokens, idx)
	if err != nil {
		return MatchClauseNode{}, 0, err
	}

	expr, incr, err := parseExpression(tokens[idx:])
	if err != nil {
		return MatchClauseNode{}, 0, err
	}
	idx += incr

	return MatchClauseNode{
		target:     atom,
		expression: expr,
	}, idx, nil
}

func parseFunctionLiteral(tokens []Tok) (FunctionLiteralNode, int, error) {

	tok, idx := tokens[0], 1
	arguments := make([]IdentifierNode, 0)

	err := guardUnexpectedInputEnd(tokens, idx)
	if err != nil {
		return FunctionLiteralNode{}, 0, err
	}

	switch tok.kind {
	case LeftParen:
		for tokens[idx].kind == Identifier {
			idNode := IdentifierNode{tokens[idx].stringVal()}
			arguments = append(arguments, idNode)
			idx++

			err := guardUnexpectedInputEnd(tokens, idx)
			if err != nil {
				return FunctionLiteralNode{}, 0, err
			}

			if tokens[idx].kind != Separator {
				return FunctionLiteralNode{}, 0, Err{
					ErrSyntax,
					fmt.Sprintf("expected a comma separated arguments list, found %s",
						tokens[idx].String()),
				}
			}
			idx++ // Separator
		}

		err := guardUnexpectedInputEnd(tokens, idx)
		if err != nil {
			return FunctionLiteralNode{}, 0, err
		}
		if tokens[idx].kind != RightParen {
			return FunctionLiteralNode{}, 0, Err{
				ErrSyntax,
				fmt.Sprintf("expected arguments list to terminate with a RightParen, found %s",
					tokens[idx].String()),
			}
		}
		idx++ // RightParen
	case Identifier:
		idNode := IdentifierNode{tok.stringVal()}
		arguments = append(arguments, idNode)
	default:
		return FunctionLiteralNode{}, 0, Err{
			ErrSyntax,
			fmt.Sprintf("malformed arguments list in function at %s", tok.String()),
		}
	}

	err = guardUnexpectedInputEnd(tokens, idx)
	if err != nil {
		return FunctionLiteralNode{}, 0, err
	}

	if tokens[idx].kind != FunctionArrow {
		return FunctionLiteralNode{}, 0, Err{
			ErrSyntax,
			fmt.Sprintf("expected FunctionArrow but found %s", tokens[idx].String()),
		}
	}
	idx++

	body, incr, err := parseExpression(tokens[idx:])
	if err != nil {
		return FunctionLiteralNode{}, 0, err
	}
	idx += incr

	return FunctionLiteralNode{
		arguments: arguments,
		body:      body,
	}, idx, nil
}

func parseFunctionCall(function Node, tokens []Tok) (FunctionCallNode, int, error) {
	idx := 1
	arguments := make([]Node, 0)

	err := guardUnexpectedInputEnd(tokens, idx)
	if err != nil {
		return FunctionCallNode{}, 0, err
	}

	for tokens[idx].kind != RightParen {
		expr, incr, err := parseExpression(tokens[idx:])
		if err != nil {
			return FunctionCallNode{}, 0, err
		}

		idx += incr
		arguments = append(arguments, expr)

		err = guardUnexpectedInputEnd(tokens, idx)
		if err != nil {
			return FunctionCallNode{}, 0, err
		}
	}
	idx++ // RightParen

	return FunctionCallNode{
		function:  function,
		arguments: arguments,
	}, idx, nil
}
