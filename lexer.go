package main

import (
	"fmt"
	"strconv"
	"unicode"
)

type Kind int

const (
	Separator Kind = iota

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

	LogicalAndOp
	LogicalOrOp
	LogicalXorOp

	LeftParen
	RightParen
	LeftBracket
	RightBracket
	LeftBrace
	RightBrace
)

type position struct {
	line, col int
}

func (p position) String() string {
	return fmt.Sprintf("%d:%d", p.line, p.col)
}

// Tok is the monomorphic struct representing all Ink program tokens
//	in the lexer.
type Tok struct {
	kind Kind
	// str and num are both present to implement Tok
	//	as a monomorphic type for all tokens; will be zero
	//	values often.
	str string
	num float64
	position
}

func (tok Tok) String() string {
	switch tok.kind {
	case Identifier, StringLiteral:
		return fmt.Sprintf("%s '%s' [%s]",
			tok.kind.String(),
			tok.str,
			tok.position.String())
	case NumberLiteral:
		return fmt.Sprintf("%s '%s' [%s]",
			tok.kind.String(),
			nToS(tok.num),
			tok.position.String())
	default:
		return fmt.Sprintf("%s [%s]",
			tok.kind.String(),
			tok.position.String())
	}
}

// Tokenize takes a stream of characters and concurrently transforms it into a stream of Tok (tokens).
func Tokenize(
	input <-chan rune,
	tokens chan<- Tok,
	fatalError bool,
	debugLexer bool,
) {
	defer close(tokens)

	var buf, strbuf string
	var strbufStartLine, strbufStartCol int

	lastKind := Separator
	lineNo, colNo := 1, 1

	simpleCommit := func(tok Tok) {
		lastKind = tok.kind
		if debugLexer {
			logDebug("lex ->", tok.String())
		}
		tokens <- tok
	}
	simpleCommitChar := func(kind Kind) {
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
			case "true":
				simpleCommitChar(TrueLiteral)
			case "false":
				simpleCommitChar(FalseLiteral)
			default:
				if unicode.IsDigit(rune(cbuf[0])) {
					f, err := strconv.ParseFloat(cbuf, 64)
					if err != nil {
						e := Err{
							ErrSyntax,
							fmt.Sprintf("parsing error in number at %d:%d, %s",
								lineNo, colNo, err.Error()),
						}
						if fatalError {
							logErr(e.reason, e.message)
						} else {
							logSafeErr(e.reason, e.message)
						}
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
	commitChar := func(kind Kind) {
		commit(Tok{
			kind:     kind,
			position: position{lineNo, colNo},
		})
	}
	ensureSeparator = func() {
		commitClear()
		switch lastKind {
		case Separator, LeftParen, LeftBracket, LeftBrace,
			AddOp, SubtractOp, MultiplyOp, DivideOp, ModulusOp, NegationOp,
			GreaterThanOp, LessThanOp, EqualOp, DefineOp, AccessorOp,
			KeyValueSeparator, FunctionArrow, MatchColon, CaseArrow:
			// do nothing
		default:
			commitChar(Separator)
		}
	}

	// Ink requires max 1 lookahead, so rather than allowing backtracking
	//	from the lexer's reader, we implement a streaming lexer with a buffer
	//	of 1, implemented as this lastChar character. Every loop we take char
	//	from lastChar if not zero, from input channel otherwise.
	var char, lastChar rune
	inStringLiteral := false

lexLoop:
	for {
		if lastChar != 0 {
			char, lastChar = lastChar, 0
		} else {
			var open bool
			char, open = <-input
			if !open {
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
				// single-line comment, keep taking until EOL
				for nextChar != '\n' {
					var ok bool
					nextChar, ok = <-input
					if !ok {
						break lexLoop // input closed
					}
				}

				ensureSeparator()
				lineNo++
				colNo = 0
			} else {
				// multi-line block comment, keep taking until end of block
				for nextChar != '`' {
					var ok bool
					nextChar, ok = <-input
					if !ok {
						break lexLoop // input closed
					}
					if nextChar == '\n' {
						lineNo++
						colNo = 0
					}
					colNo++
				}
			}
		case char == '\n':
			ensureSeparator()
			lineNo++
			colNo = 0
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
		case char == '&':
			commitChar(LogicalAndOp)
		case char == '|':
			commitChar(LogicalOrOp)
		case char == '^':
			commitChar(LogicalXorOp)
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

func (kind Kind) String() string {
	switch kind {
	case Block:
		return "expression block"
	case UnaryExpr:
		return "unary expression"
	case BinaryExpr:
		return "binary expression"
	case MatchExpr:
		return "match expression"
	case MatchClause:
		return "match clause"

	case Identifier:
		return "identifier"
	case EmptyIdentifier:
		return "'_'"

	case FunctionCall:
		return "function call"

	case NumberLiteral:
		return "number literal"
	case StringLiteral:
		return "string literal"
	case ObjectLiteral:
		return "composite literal"
	case ListLiteral:
		return "list composite literal"
	case FunctionLiteral:
		return "function literal"

	case TrueLiteral:
		return "'true'"
	case FalseLiteral:
		return "'false'"

	case AccessorOp:
		return "'.'"

	case EqualOp:
		return "'='"
	case FunctionArrow:
		return "'=>'"

	case KeyValueSeparator:
		return "':'"
	case DefineOp:
		return "':='"
	case MatchColon:
		return "'::'"

	case CaseArrow:
		return "'->'"
	case SubtractOp:
		return "'-'"

	case NegationOp:
		return "'~'"
	case AddOp:
		return "'+'"
	case MultiplyOp:
		return "'*'"
	case DivideOp:
		return "'/'"
	case ModulusOp:
		return "'%'"
	case GreaterThanOp:
		return "'>'"
	case LessThanOp:
		return "'<'"

	case LogicalAndOp:
		return "'&'"
	case LogicalOrOp:
		return "'|'"
	case LogicalXorOp:
		return "'^'"

	case Separator:
		return "','"
	case LeftParen:
		return "'('"
	case RightParen:
		return "')'"
	case LeftBracket:
		return "'['"
	case RightBracket:
		return "']'"
	case LeftBrace:
		return "'{'"
	case RightBrace:
		return "'}'"

	default:
		return "unknown token"
	}
}
