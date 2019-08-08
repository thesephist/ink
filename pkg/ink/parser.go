package ink

import (
	"fmt"
	"strings"
)

// Node represents an abstract syntax tree (AST) node in an Ink program.
type Node interface {
	String() string
	Position() position
	Eval(*StackFrame, bool) (Value, error)
}

// a string representation of the Position of a given node,
//	appropriate for an error message
func poss(n Node) string {
	return n.Position().String()
}

type UnaryExprNode struct {
	operator Kind
	operand  Node
	position
}

func (n UnaryExprNode) String() string {
	return fmt.Sprintf("Unary %s (%s)", n.operator, n.operand)
}

func (n UnaryExprNode) Position() position {
	return n.position
}

type BinaryExprNode struct {
	operator     Kind
	leftOperand  Node
	rightOperand Node
	position
}

func (n BinaryExprNode) String() string {
	return fmt.Sprintf("Binary (%s) %s (%s)", n.leftOperand, n.operator, n.rightOperand)
}

func (n BinaryExprNode) Position() position {
	return n.position
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

func (n FunctionCallNode) Position() position {
	return n.function.Position()
}

type MatchClauseNode struct {
	target     Node
	expression Node
}

func (n MatchClauseNode) String() string {
	return fmt.Sprintf("Clause (%s) -> (%s)", n.target, n.expression)
}

func (n MatchClauseNode) Position() position {
	return n.target.Position()
}

type MatchExprNode struct {
	condition Node
	clauses   []MatchClauseNode
	position
}

func (n MatchExprNode) String() string {
	clauses := make([]string, len(n.clauses))
	for i, c := range n.clauses {
		clauses[i] = c.String()
	}
	return fmt.Sprintf("Match on (%s) to {%s}", n.condition, clauses)
}

func (n MatchExprNode) Position() position {
	return n.position
}

type ExpressionListNode struct {
	expressions []Node
	position
}

func (n ExpressionListNode) String() string {
	exprs := make([]string, len(n.expressions))
	for i, expr := range n.expressions {
		exprs[i] = expr.String()
	}
	return fmt.Sprintf("Expression List (%s)", strings.Join(exprs, ", "))
}

func (n ExpressionListNode) Position() position {
	return n.position
}

type EmptyIdentifierNode struct {
	position
}

func (n EmptyIdentifierNode) String() string {
	return "Empty Identifier"
}

func (n EmptyIdentifierNode) Position() position {
	return n.position
}

type IdentifierNode struct {
	val string
	position
}

func (n IdentifierNode) String() string {
	return fmt.Sprintf("Identifier '%s'", n.val)
}

func (n IdentifierNode) Position() position {
	return n.position
}

type NumberLiteralNode struct {
	val float64
	position
}

func (n NumberLiteralNode) String() string {
	return fmt.Sprintf("Number %s", nToS(n.val))
}

func (n NumberLiteralNode) Position() position {
	return n.position
}

type StringLiteralNode struct {
	val string
	position
}

func (n StringLiteralNode) String() string {
	return fmt.Sprintf("String '%s'", n.val)
}

func (n StringLiteralNode) Position() position {
	return n.position
}

type BooleanLiteralNode struct {
	val bool
	position
}

func (n BooleanLiteralNode) String() string {
	return fmt.Sprintf("Boolean %t", n.val)
}

func (n BooleanLiteralNode) Position() position {
	return n.position
}

type ObjectLiteralNode struct {
	entries []ObjectEntryNode
	position
}

func (n ObjectLiteralNode) String() string {
	entries := make([]string, len(n.entries))
	for i, e := range n.entries {
		entries[i] = e.String()
	}
	return fmt.Sprintf("Object {%s}",
		strings.Join(entries, ", "))
}

func (n ObjectLiteralNode) Position() position {
	return n.position
}

type ObjectEntryNode struct {
	key Node
	val Node
	position
}

func (n ObjectEntryNode) String() string {
	return fmt.Sprintf("Object Entry (%s): (%s)", n.key, n.val)
}

type ListLiteralNode struct {
	vals []Node
	position
}

func (n ListLiteralNode) String() string {
	vals := make([]string, len(n.vals))
	for i, v := range n.vals {
		vals[i] = v.String()
	}
	return fmt.Sprintf("List [%s]", strings.Join(vals, ", "))
}

func (n ListLiteralNode) Position() position {
	return n.position
}

type FunctionLiteralNode struct {
	arguments []Node
	body      Node
	position
}

func (n FunctionLiteralNode) String() string {
	args := make([]string, len(n.arguments))
	for i, a := range n.arguments {
		args[i] = a.String()
	}
	return fmt.Sprintf("Function (%s) => (%s)", strings.Join(args, ", "), n.body)
}

func (n FunctionLiteralNode) Position() position {
	return n.position
}

func guardUnexpectedInputEnd(tokens []Tok, idx int) error {
	if idx >= len(tokens) {
		if len(tokens) > 0 {
			return Err{
				ErrSyntax,
				fmt.Sprintf("unexpected end of input at %s", tokens[len(tokens)-1]),
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
					LogErr(e.reason, e.message)
				} else {
					LogSafeErr(e.reason, e.message)
				}
			} else {
				LogErrf(ErrAssert, "err raised that was not of Err type -> %s",
					err.Error())
			}
			return
		}

		if debugParser {
			LogDebug("parse ->", expr.String())
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

	case GreaterThanOp, LessThanOp, EqualOp:
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
		GreaterThanOp, LessThanOp, EqualOp, DefineOp, AccessorOp:
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
			operator:     ops[0].kind,
			leftOperand:  tree,
			rightOperand: nodes[0],
			position:     ops[0].position,
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
		GreaterThanOp, LessThanOp, EqualOp, DefineOp, AccessorOp:
		binExpr, incr, err := parseBinaryExpression(atom, nextTok, tokens[idx:], -1)
		if err != nil {
			return nil, 0, err
		}
		idx += incr

		// Binary expressions are often followed by a match
		if idx < len(tokens) && tokens[idx].kind == MatchColon {
			colonPos := tokens[idx].position
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
				position:  colonPos,
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
			position:  nextTok.position,
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
			operator: tok.kind,
			operand:  atom,
			position: tok.position,
		}, idx + 1, nil
	}

	err = guardUnexpectedInputEnd(tokens, idx)
	if err != nil {
		return nil, 0, err
	}

	var atom Node
	switch tok.kind {
	case NumberLiteral:
		return NumberLiteralNode{tok.num, tok.position}, idx, nil
	case StringLiteral:
		return StringLiteralNode{tok.str, tok.position}, idx, nil
	case TrueLiteral:
		return BooleanLiteralNode{true, tok.position}, idx, nil
	case FalseLiteral:
		return BooleanLiteralNode{false, tok.position}, idx, nil
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
			atom = IdentifierNode{tok.str, tok.position}
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
			return EmptyIdentifierNode{tok.position}, idx, nil
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
			atom = ExpressionListNode{
				expressions: exprs,
				position:    tok.position,
			}
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
			if tokens[idx].kind == KeyValueSeparator {
				idx++
			} else {
				return nil, 0, Err{
					ErrSyntax,
					fmt.Sprintf("expected %s after composite key, found %s",
						KeyValueSeparator.String(), tokens[idx]),
				}
			}

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
			entries = append(entries, ObjectEntryNode{
				key:      keyExpr,
				val:      valExpr,
				position: keyExpr.Position(),
			})

			err = guardUnexpectedInputEnd(tokens, idx)
			if err != nil {
				return nil, 0, err
			}
		}
		idx++ // RightBrace

		return ObjectLiteralNode{
			entries:  entries,
			position: tok.position,
		}, idx, nil
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

		return ListLiteralNode{
			vals:     vals,
			position: tok.position,
		}, idx, nil
	default:
		return nil, 0, Err{
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
			fmt.Sprintf("expected %s, but got %s", CaseArrow, tokens[idx]),
		}
	}
	idx++ // CaseArrow

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
			tk := tokens[idx]
			if tk.kind == Identifier {
				idNode := IdentifierNode{tk.str, tk.position}
				arguments = append(arguments, idNode)
			} else if tk.kind == EmptyIdentifier {
				idNode := EmptyIdentifierNode{tk.position}
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
					fmt.Sprintf("expected arguments in a list separated by %s, found %s",
						Separator, tokens[idx]),
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
				fmt.Sprintf("expected arguments list to terminate with %s, found %s",
					RightParen, tokens[idx]),
			}
		}
		idx++ // RightParen
	case Identifier:
		idNode := IdentifierNode{tok.str, tok.position}
		arguments = append(arguments, idNode)
	case EmptyIdentifier:
		idNode := EmptyIdentifierNode{tok.position}
		arguments = append(arguments, idNode)
	default:
		return FunctionLiteralNode{}, 0, Err{
			ErrSyntax,
			fmt.Sprintf("malformed arguments list in function at %s", tok),
		}
	}

	err = guardUnexpectedInputEnd(tokens, idx)
	if err != nil {
		return FunctionLiteralNode{}, 0, err
	}

	if tokens[idx].kind != FunctionArrow {
		return FunctionLiteralNode{}, 0, Err{
			ErrSyntax,
			fmt.Sprintf("expected %s but found %s", FunctionArrow, tokens[idx]),
		}
	}
	idx++ // FunctionArrow

	body, incr, err := parseExpression(tokens[idx:])
	if err != nil {
		return FunctionLiteralNode{}, 0, err
	}
	idx += incr

	return FunctionLiteralNode{
		arguments: arguments,
		body:      body,
		position:  tokens[0].position,
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
