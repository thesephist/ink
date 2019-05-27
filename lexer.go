package main

import (
	"fmt"
	"log"
	"strconv"
	"unicode"
)

const (
	_ = iota
	Block
	UnaryExpr
	BinaryExpr
	MatchExpr
	MatchClause

	Identifier
	EmptyIdentifier

	FunctionCall

	NumberLiteral
	StringLiteral
	ObjectLiteral
	ListLiteral
	FunctionLiteral

	TrueLiteral
	FalseLiteral
	NullLiteral

	// ambiguous operators and symbols
	AccessorOp

	IsOp

	// =
	EqualOp
	FunctionArrow

	// :
	DefineOp
	MatchColon

	// -
	CaseArrow
	SubtractOp

	// single char, unambiguous
	NegationOp
	AddOp
	MultiplyOp
	DivideOp
	ModulusOp
	GreaterThanOp
	LessThanOp

	Separator
	LeftParen
	RightParen
	LeftBracket
	RightBracket
	LeftBrace
	RightBrace
)

type span struct {
	startLine, startCol int
	endLine, endCol     int
}

type Tok struct {
	val  interface{}
	kind int
	span
}

func (tok *Tok) stringVal() string {
	switch v := tok.val.(type) {
	case string:
		return string(v)
	case float64:
		return strconv.FormatFloat(float64(v), 'f', -1, 64)
	default:
		return ""
	}
}

func (tok *Tok) numberVal() float64 {
	switch v := tok.val.(type) {
	case float64:
		return float64(v)
	default:
		return 0
	}
}

func (tok Tok) String() string {
	return fmt.Sprintf("Token %s: \t%s", tokKindToName(tok.kind), tok.stringVal())
}

func Tokenize(input <-chan rune, tokens chan<- Tok, done chan<- bool) {
	lastTokKind := Separator
	buf := ""
	strbuf := ""
	var strbufStartLine, strbufStartCol int

	lineNo := 0
	colNo := 0

	simpleCommit := func(tok Tok) {
		lastTokKind = tok.kind
		tok.endLine = lineNo
		tok.endCol = colNo
		tokens <- tok
	}
	simpleCommitChar := func(kind int) {
		simpleCommit(Tok{
			"",
			kind,
			span{lineNo, colNo, lineNo, colNo + 1},
		})
	}
	commitClear := func() {
		if buf != "" {
			switch buf {
			case ".":
				simpleCommitChar(AccessorOp)
			case "is":
				simpleCommitChar(IsOp)
			case "=":
				simpleCommitChar(EqualOp)
			case "=>":
				simpleCommitChar(FunctionArrow)
			case ":=":
				simpleCommitChar(DefineOp)
			case "::":
				simpleCommitChar(MatchColon)
			case "-":
				simpleCommitChar(SubtractOp)
			case "->":
				simpleCommitChar(CaseArrow)
			case ">":
				simpleCommitChar(GreaterThanOp)
			case "true":
				simpleCommitChar(TrueLiteral)
			case "false":
				simpleCommitChar(FalseLiteral)
			case "null":
				simpleCommitChar(NullLiteral)
			default:
				if unicode.IsDigit(rune(buf[0])) {
					f, err := strconv.ParseFloat(buf, 64)
					if err != nil {
						log.Fatalf("Parsing error in number at %d:%d, %s", lineNo, colNo, err.Error())
					}
					simpleCommit(Tok{
						f,
						NumberLiteral,
						span{lineNo, colNo - len(buf), lineNo, colNo + 1},
					})
				} else {
					simpleCommit(Tok{
						buf,
						Identifier,
						span{lineNo, colNo - len(buf), lineNo, colNo + 1},
					})
				}
			}
			buf = ""
		}
	}
	commit := func(tok Tok) {
		commitClear()
		simpleCommit(tok)
	}
	commitChar := func(kind int) {
		commit(Tok{
			"",
			kind,
			span{lineNo, colNo, lineNo, colNo + 1},
		})
	}
	ensureSeparator := func() {
		commitClear()
		switch lastTokKind {
		case Separator, LeftParen, LeftBracket, LeftBrace:
			// do nothing
		default:
			commitChar(Separator)
		}
	}

	inStringLiteral := false

	// parse stuff
	go func() {
		for char := range input {
			switch {
			case char == '\'':
				if inStringLiteral {
					commit(Tok{
						strbuf,
						StringLiteral,
						span{strbufStartLine, strbufStartCol, lineNo, colNo + 1},
					})
				} else {
					strbuf = ""
					strbufStartLine, strbufStartCol = lineNo, colNo
				}
				inStringLiteral = !inStringLiteral
			case inStringLiteral:
				strbuf += string(char)
			case char == '`':
				nextChar := <-input
				for nextChar != '`' {
					nextChar = <-input
				}
			case char == '\n':
				lineNo++
				colNo = 0
				ensureSeparator()
			case unicode.IsSpace(char):
				commitClear()
			case char == '_':
				commitChar(EmptyIdentifier)
			case char == '~':
				commitChar(NegationOp)
			case char == '+':
				commitChar(AddOp)
			case char == '*':
				commitChar(MultiplyOp)
			case char == '/':
				commitChar(DivideOp)
			case char == '<':
				commitChar(LessThanOp)
			case char == ',':
				commitChar(Separator)
			case char == '(':
				commitChar(LeftParen)
			case char == ')':
				ensureSeparator()
				commitChar(RightParen)
			case char == '[':
				commitChar(LeftBracket)
			case char == ']':
				ensureSeparator()
				commitChar(RightBracket)
			case char == '{':
				commitChar(LeftBrace)
			case char == '}':
				ensureSeparator()
				commitChar(RightBrace)
			default:
				buf += string(char)
			}
			colNo++
		}

		close(tokens)
		done <- true
	}()
}

func isValidIdentifierStartChar(char rune) bool {
	if unicode.IsLetter(char) {
		return true
	}

	switch char {
	case '@', '!', '?':
		return true
	default:
		return false
	}
}

func isValidIdentifierChar(char rune) bool {
	return isValidIdentifierStartChar(char) || unicode.IsDigit(char)
}

func tokKindToName(kind int) string {
	switch kind {
	case Block:
		return "Block"
	case UnaryExpr:
		return "UnaryExpr"
	case BinaryExpr:
		return "BinaryExpr"
	case MatchExpr:
		return "MatchExpr"
	case MatchClause:
		return "MatchClause"

	case Identifier:
		return "Identifier"
	case EmptyIdentifier:
		return "EmptyIdentifier"

	case FunctionCall:
		return "FunctionCall"

	case NumberLiteral:
		return "NumberLiteral"
	case StringLiteral:
		return "StringLiteral"
	case ObjectLiteral:
		return "ObjectLiteral"
	case ListLiteral:
		return "ListLiteral"
	case FunctionLiteral:
		return "FunctionLiteral"

	case TrueLiteral:
		return "TrueLiteral"
	case FalseLiteral:
		return "FalseLiteral"
	case NullLiteral:
		return "NullLiteral"

	case AccessorOp:
		return "AccessorOp"

	case IsOp:
		return "IsOp"

	case EqualOp:
		return "EqualOp"
	case FunctionArrow:
		return "FunctionArrow"

	case DefineOp:
		return "DefineOp"
	case MatchColon:
		return "MatchColon"

	case CaseArrow:
		return "CaseArrow"
	case SubtractOp:
		return "SubtractOp"

	case NegationOp:
		return "NegationOp"
	case AddOp:
		return "AddOp"
	case MultiplyOp:
		return "MultiplyOp"
	case DivideOp:
		return "DivideOp"
	case ModulusOp:
		return "ModulusOp"
	case GreaterThanOp:
		return "GreaterThanOp"
	case LessThanOp:
		return "LessThanOp"

	case Separator:
		return "Separator"
	case LeftParen:
		return "LeftParen"
	case RightParen:
		return "RightParen"
	case LeftBracket:
		return "LeftBracket"
	case RightBracket:
		return "RightBracket"
	case LeftBrace:
		return "LeftBrace"
	case RightBrace:
		return "RightBrace"

	default:
		return "UnknownToken"
	}
}
