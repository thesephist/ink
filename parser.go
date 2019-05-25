package main

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
