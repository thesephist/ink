package main

const (
	Block = iota
	UnaryExpr
	BinaryExpr
	FunctionCallExpr
	MatchExpr
	MatchClause
	Identifier

	NumberLiteral
	StringLiteral
	BooleanLiteral
	NullLiteral
	ObjectLiteral
	ListLiteral
	FunctionLiteral

	UnaryOp
	BinaryOp
)

type Node struct {
	kind int
}

func Parse(input <-chan Tok) Node {
	return Node{}
}

type BlockNode struct {
	expressions []Node
	Node
}

func parseBlock(input <-chan Tok) Node {
	return Node{}
}

func parseExpression(input <-chan Tok) Node {
	return Node{}
}

func parseMatchClauses(input <-chan Tok) []Node {
	return []Node{}
}

func parseMatchClause(input <-chan Tok) Node {
	return Node{}
}

func parseAtom(input <-chan Tok) Node {
	return Node{}
}

func parseLiteral(input <-chan Tok) Node {
	return Node{}
}

func parseObjectLiteral(input <-chan Tok) Node {
	return Node{}
}

func parseListLiteral(input <-chan Tok) Node {
	return Node{}
}

func parseFunctionLiteral(input <-chan Tok) Node {
	return Node{}
}

func isUnaryOp(tok string) bool {
	switch tok {
	case "~":
		return true
	default:
		return false
	}
}

func isBinaryOp(tok string) bool {
	switch tok {
	case "+", "-", "/", "*", "%", ">", "<", "==", "is", ":=", ".":
		return true
	default:
		return false
	}
}
