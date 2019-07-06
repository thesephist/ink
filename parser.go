package main

import (
	"fmt"
	"strings"
)

// Node represents an abstract syntax tree (AST) node in an Ink program.
type Node interface {
	String() string
	Eval(*StackFrame, bool) (Value, error)
}

type UnaryExprNode struct {
	operator Tok
	operand  Node
}

func (n UnaryExprNode) opChar() string {
	return "~"
}

func (n UnaryExprNode) String() string {
	return fmt.Sprintf("Unary %s (%s)", n.opChar(), n.operand.String())
}

type BinaryExprNode struct {
	operator     Tok
	leftOperand  Node
	rightOperand Node
}

func (n BinaryExprNode) opChar() string {
	switch n.operator.kind {
	case AddOp:
		return "+"
	case SubtractOp:
		return "-"
	case MultiplyOp:
		return "*"
	case DivideOp:
		return "/"
	case ModulusOp:
		return "%"
	case GreaterThanOp:
		return ">"
	case LessThanOp:
		return "<"
	case EqualOp:
		return "="
	case EqRefOp:
		return "is"
	case LogicalAndOp:
		return "&"
	case LogicalOrOp:
		return "|"
	case LogicalXorOp:
		return "^"
	case DefineOp:
		return ":="
	case AccessorOp:
		return "."
	default:
		logErr(ErrAssert, "unknown operator in binary expression")
		return "?"
	}
}

func (n BinaryExprNode) String() string {
	return fmt.Sprintf("Binary (%s) %s (%s)",
		n.leftOperand.String(),
		n.opChar(),
		n.rightOperand.String())
}

type FunctionCallNode struct {
	function  Node
	arguments []Node
}

func (n FunctionCallNode) String() string {
	args := make([]string, len(n.arguments))
	for i, a := range n.arguments {
		args[i] = a.String()
	}
	return fmt.Sprintf("Call (%s) on (%s)",
		n.function,
		strings.Join(args, ", "))
}

type MatchClauseNode struct {
	target     Node
	expression Node
}

func (n MatchClauseNode) String() string {
	return fmt.Sprintf("Clause (%s) -> (%s)",
		n.target.String(),
		n.expression.String())
}

type MatchExprNode struct {
	condition Node
	clauses   []MatchClauseNode
}

func (n MatchExprNode) String() string {
	clauses := make([]string, len(n.clauses))
	for i, c := range n.clauses {
		clauses[i] = c.String()
	}
	return fmt.Sprintf("Match on (%s) to {%s}",
		n.condition.String(),
		clauses)
}

type ExpressionListNode struct {
	expressions []Node
}

func (n ExpressionListNode) String() string {
	exprs := make([]string, len(n.expressions))
	for i, expr := range n.expressions {
		exprs[i] = expr.String()
	}
	return fmt.Sprintf("Expression List (%s)", strings.Join(exprs, ", "))
}

type EmptyIdentifierNode struct{}

func (n EmptyIdentifierNode) String() string {
	return "Empty Identifier"
}

type IdentifierNode struct {
	val string
}

func (n IdentifierNode) String() string {
	return fmt.Sprintf("Identifier '%s'", n.val)
}

type NumberLiteralNode struct {
	val float64
}

func (n NumberLiteralNode) String() string {
	return fmt.Sprintf("Number %s", nToS(n.val))
}

type StringLiteralNode struct {
	val string
}

func (n StringLiteralNode) String() string {
	return fmt.Sprintf("String '%s'", n.val)
}

type BooleanLiteralNode struct {
	val bool
}

func (n BooleanLiteralNode) String() string {
	return fmt.Sprintf("Boolean %t", n.val)
}

type ObjectLiteralNode struct {
	entries []ObjectEntryNode
}

func (n ObjectLiteralNode) String() string {
	entries := make([]string, len(n.entries))
	for i, e := range n.entries {
		entries[i] = e.String()
	}
	return fmt.Sprintf("Object {%s}",
		strings.Join(entries, ", "))
}

type ObjectEntryNode struct {
	key Node
	val Node
}

func (n ObjectEntryNode) String() string {
	return fmt.Sprintf("Object Entry (%s): (%s)", n.key.String(), n.val.String())
}

type ListLiteralNode struct {
	vals []Node
}

func (n ListLiteralNode) String() string {
	vals := make([]string, len(n.vals))
	for i, v := range n.vals {
		vals[i] = v.String()
	}
	return fmt.Sprintf("List [%s]", strings.Join(vals, ", "))
}

type FunctionLiteralNode struct {
	arguments []Node
	body      Node
}

func (n FunctionLiteralNode) String() string {
	args := make([]string, len(n.arguments))
	for i, a := range n.arguments {
		args[i] = a.String()
	}
	return fmt.Sprintf("Function (%s) => (%s)",
		strings.Join(args, ", "),
		n.body.String(),
	)
}

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

// Parse concurrently transforms a stream of Tok (tokens) to Node (AST nodes).
//	This implementation uses recursive descent parsing.
func Parse(
	tokenStream <-chan Tok,
	nodes chan<- Node,
	fatalError bool,
	debugParser bool,
) {
	defer close(nodes)

	tokens := make([]Tok, 0)
	for tok := range tokenStream {
		tokens = append(tokens, tok)
	}

	idx, length := 0, len(tokens)
	for idx < length {
		if tokens[idx].kind == Separator {
			// this sometimes happens when the repl receives comment inputs
			idx++
			continue
		}

		expr, incr, err := parseExpression(tokens[idx:])
		idx += incr

		if err != nil {
			e, isErr := err.(Err)
			if isErr {
				if fatalError {
					logErr(e.reason, e.message)
				} else {
					logSafeErr(e.reason, e.message)
				}
			} else {
				logErrf(ErrAssert, "err raised that was not of Err type -> %s",
					err.Error())
			}
			return
		}

		if debugParser {
			logDebug("parse ->", expr.String())
		}
		nodes <- expr
	}
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
		return 40

	case GreaterThanOp, LessThanOp, EqualOp, EqRefOp:
		return 30

	case LogicalAndOp:
		return 20
	case LogicalXorOp:
		return 15
	case LogicalOrOp:
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
		LogicalAndOp, LogicalOrOp, LogicalXorOp,
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
		LogicalAndOp, LogicalOrOp, LogicalXorOp,
		GreaterThanOp, LessThanOp, EqualOp, EqRefOp, DefineOp, AccessorOp:
		binExpr, incr, err := parseBinaryExpression(atom, nextTok, tokens[idx:], -1)
		if err != nil {
			return nil, 0, err
		}
		idx += incr

		// Binary expressions are often followed by a match
		if idx < len(tokens) && tokens[idx].kind == MatchColon {
			idx++ // MatchColon

			clauses, incr, err := parseMatchBody(tokens[idx:])
			if err != nil {
				return nil, 0, err
			}
			idx += incr

			consumeDanglingSeparator()
			return MatchExprNode{
				condition: binExpr,
				clauses:   clauses,
			}, idx, nil
		} else {
			consumeDanglingSeparator()
			return binExpr, idx, nil
		}

	case MatchColon:
		clauses, incr, err := parseMatchBody(tokens[idx:])
		if err != nil {
			return nil, 0, err
		}
		idx += incr

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
	case NumberLiteral:
		return NumberLiteralNode{tok.num}, idx, nil
	case StringLiteral:
		return StringLiteralNode{tok.str}, idx, nil
	case TrueLiteral:
		return BooleanLiteralNode{true}, idx, nil
	case FalseLiteral:
		return BooleanLiteralNode{false}, idx, nil
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
			atom = IdentifierNode{tok.str}
		}
		// may be called as a function, so flows beyond
		//	switch case
	case EmptyIdentifier:
		if tokens[idx].kind == FunctionArrow {
			var err error
			atom, idx, err = parseFunctionLiteral(tokens)
			if err != nil {
				return nil, 0, err
			}

			// parseAtom should not consume trailing Separators, but
			// 	parseFunctionLiteral does because it ends with expressions.
			// 	so we backtrack one token.
			return atom, idx - 1, nil
		} else {
			return EmptyIdentifierNode{}, idx, nil
		}
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

// parses everything that follows MatchColon
// 	does not consume dangling separator -- that's for parseExpression
func parseMatchBody(tokens []Tok) ([]MatchClauseNode, int, error) {
	idx := 1 // LeftBrace
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

	return clauses, idx, nil
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
	arguments := make([]Node, 0)

	err := guardUnexpectedInputEnd(tokens, idx)
	if err != nil {
		return FunctionLiteralNode{}, 0, err
	}

	switch tok.kind {
	case LeftParen:
		for {
			if tokens[idx].kind == Identifier {
				idNode := IdentifierNode{tokens[idx].str}
				arguments = append(arguments, idNode)
			} else if tokens[idx].kind == EmptyIdentifier {
				idNode := EmptyIdentifierNode{}
				arguments = append(arguments, idNode)
			} else {
				break
			}
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
		idNode := IdentifierNode{tok.str}
		arguments = append(arguments, idNode)
	case EmptyIdentifier:
		idNode := EmptyIdentifierNode{}
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
