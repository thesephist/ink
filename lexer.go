package main

import (
	"fmt"
	"strconv"
	"unicode"
)

const (
	Separator = iota

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

	// ambiguous operators and symbols
	AccessorOp

	EqRefOp

	// =
	EqualOp
	FunctionArrow

	// :
	KeyValueSeparator
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

	// TODO: Bitwise/Logical & | ^ (and or xor)
	//	and order of operations / implement thru language to eval

	LeftParen
	RightParen
	LeftBracket
	RightBracket
	LeftBrace
	RightBrace
)

type position struct {
	startLine, startCol int
}

func (sp *position) String() string {
	return fmt.Sprintf("%d:%d", sp.startLine, sp.startCol)
}

type Tok struct {
	kind int
	// str and num are both present to implement Tok
	//	as a monomorphic type for all tokens; will be zero
	//	values often.
	str string
	num float64
	position
}

func (tok Tok) String() string {
	label := tokKindToName(tok.kind)

	switch tok.kind {
	case Identifier, StringLiteral:

		return fmt.Sprintf("%s %s [%s]",
			label,
			tok.str,
			tok.position.String())
	case NumberLiteral:
		return fmt.Sprintf("%s %s [%s]",
			label,
			nToS(tok.num),
			tok.position.String())
	default:
		return fmt.Sprintf("%s [%s]",
			label,
			tok.position.String())
	}
}

func Tokenize(
	input <-chan rune,
	tokens chan<- Tok,
	errors chan<- Err,
	debugLexer bool,
) {
	var buf, strbuf string
	var strbufStartLine, strbufStartCol int

	syntaxErrored := false

	lastTokKind := Separator
	lineNo := 1
	colNo := 1

	simpleCommit := func(tok Tok) {
		lastTokKind = tok.kind
		if debugLexer {
			logDebug("lex ->", tok.String())
		}
		tokens <- tok
	}
	simpleCommitChar := func(kind int) {
		simpleCommit(Tok{
			kind:     kind,
			position: position{lineNo, colNo},
		})
	}
	forbidden := func() {
		// no-op, re-bound below
		logErrf(ErrAssert, "this function should never run")
	}
	ensureSeparator := forbidden
	commitClear := forbidden
	commitClear = func() {
		if buf != "" {
			cbuf := buf
			buf = ""
			switch cbuf {
			case "is":
				simpleCommitChar(EqRefOp)
			case "true":
				simpleCommitChar(TrueLiteral)
			case "false":
				simpleCommitChar(FalseLiteral)
			default:
				if unicode.IsDigit(rune(cbuf[0])) {
					f, err := strconv.ParseFloat(cbuf, 64)
					if err != nil {
						syntaxErrored = true
						errors <- Err{
							ErrSyntax,
							fmt.Sprintf("parsing error in number at %d:%d, %s", lineNo, colNo, err.Error()),
						}
						close(tokens)
						close(errors)
						// trick to prevent sending tokens, errors on closed channels
						simpleCommit = func(tok Tok) {}
						commitClear = func() {}
						return
					}
					simpleCommit(Tok{
						num:      f,
						kind:     NumberLiteral,
						position: position{lineNo, colNo - len(cbuf)},
					})
				} else {
					simpleCommit(Tok{
						str:      cbuf,
						kind:     Identifier,
						position: position{lineNo, colNo - len(cbuf)},
					})
				}
			}
		}
	}
	commit := func(tok Tok) {
		commitClear()
		simpleCommit(tok)
	}
	commitChar := func(kind int) {
		commit(Tok{
			kind:     kind,
			position: position{lineNo, colNo},
		})
	}
	ensureSeparator = func() {
		commitClear()
		switch lastTokKind {
		case Separator, LeftParen, LeftBracket, LeftBrace,
			AddOp, SubtractOp, MultiplyOp, DivideOp, ModulusOp, NegationOp,
			GreaterThanOp, LessThanOp, EqualOp, EqRefOp, DefineOp, AccessorOp,
			KeyValueSeparator, FunctionArrow, MatchColon, CaseArrow:
			// do nothing
		default:
			commitChar(Separator)
		}
	}

	go func() {
		// Ink requires max 1 lookahead, so rather than allowing backtracking
		//	from the lexer's reader, we implement a streaming lexer with a buffer
		//	of 1, implemented as this lastChar character. Every loop we take char
		//	from lastChar if not zero, from input channel otherwise.
		var char, lastChar rune
		inStringLiteral := false
		for {
			if lastChar != 0 {
				char = lastChar
				lastChar = 0
			} else {
				char = <-input
				if char == 0 {
					break
				}
			}
			switch {
			case char == '\'':
				if inStringLiteral {
					commit(Tok{
						str:      strbuf,
						kind:     StringLiteral,
						position: position{strbufStartLine, strbufStartCol},
					})
				} else {
					strbuf = ""
					strbufStartLine, strbufStartCol = lineNo, colNo
				}
				inStringLiteral = !inStringLiteral
			case inStringLiteral:
				if char == '\n' {
					lineNo++
					colNo = 0
					strbuf += string(char)
				} else if char == '\\' {
					// backtick escapes like in most other languages,
					//	so just consume whatever the next char is into
					//	the current string buffer
					strbuf += string(<-input)
					colNo++
				} else {
					strbuf += string(char)
				}
			case char == '`':
				nextChar := <-input
				if nextChar == '`' {
					for nextChar != '\n' {
						nextChar = <-input
					}
					lineNo++
					colNo = 0
					ensureSeparator()
				} else {
					for nextChar != '`' {
						nextChar = <-input
						if nextChar == '\n' {
							lineNo++
							colNo = 0
						}
						colNo++
					}
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
			case char == '%':
				commitChar(ModulusOp)
			case char == '<':
				commitChar(LessThanOp)
			case char == '>':
				commitChar(GreaterThanOp)
			case char == ',':
				commitChar(Separator)
			case char == '.':
				// only non-AccessorOp case is [Number token] . [Number],
				//	so we commit and bail early if the buf is empty or contains
				//	a clearly non-numeric token. Note that this means all numbers
				//	must start with a digit. i.e. .5 is not 0.5 but a syntax error.
				//	This is the case since we don't know what the last token was,
				//	and I think streaming parse is worth the tradeoffs of losing
				//	that context.
				committed := false
				for _, d := range buf {
					// XXX: there's probably a faster way
					if !unicode.IsDigit(d) {
						commitChar(AccessorOp)
						committed = true
						break
					}
				}
				if !committed {
					if buf == "" {
						commitChar(AccessorOp)
					} else {
						buf += "."
					}
				}
			case char == ':':
				nextChar := <-input
				colNo++
				if nextChar == '=' {
					commitChar(DefineOp)
				} else if nextChar == ':' {
					commitChar(MatchColon)
				} else {
					// key is parsed as expression, so make sure
					//	we mark expression end (Separator)
					ensureSeparator()
					commitChar(KeyValueSeparator)
					lastChar = nextChar
				}
			case char == '=':
				nextChar := <-input
				colNo++
				if nextChar == '>' {
					commitChar(FunctionArrow)
				} else {
					commitChar(EqualOp)
					lastChar = nextChar
				}
			case char == '-':
				nextChar := <-input
				colNo++
				if nextChar == '>' {
					commitChar(CaseArrow)
				} else {
					commitChar(SubtractOp)
					lastChar = nextChar
				}
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

		ensureSeparator()

		if !syntaxErrored {
			// must not have closed channels yet
			close(tokens)
			close(errors)
		}
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

	case AccessorOp:
		return "AccessorOp"

	case EqRefOp:
		return "EqRefOp"

	case EqualOp:
		return "EqualOp"
	case FunctionArrow:
		return "FunctionArrow"

	case KeyValueSeparator:
		return "KeyValueSeparator"
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
