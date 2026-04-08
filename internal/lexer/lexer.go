package lexer

import "fmt"

type Type int

const (
	EOF Type = iota
	NEWLINE
	ERROR
	STRING
	NUM
	BOOL
	IDENT
	APP
	SCREEN
	FN
	IF
	ELSE
	FOR
	IN
	WHILE
	RETURN
	NATIVE
	IMPORT
	AWAIT
	BACKGROUND
	TEXT
	BUTTON
	INPUT
	LIST
	IMAGE
	SWITCH
	SLIDER
	CHECKBOX
	PROGRESS
	NAVIGATE
	BACK
	TOAST
	VIBRATE
	NOTIFY
	RENDER
	HTTP
	STATE
	SHARE
	OPEN
	TIMER
	ALERT
	LPAREN
	RPAREN
	LBRACE
	RBRACE
	LBRACKET
	RBRACKET
	COMMA
	DOT
	COLON
	SEMI
	ASSIGN
	ARROW
	PLUS
	MINUS
	STAR
	SLASH
	MOD
	EQ
	NEQ
	LT
	GT
	LTE
	GTE
	AND
	OR
	NOT
)

func (t Type) String() string {
	names := [...]string{
		"EOF", "NEWLINE", "ERROR",
		"STRING", "NUM", "BOOL", "IDENT",
		"APP", "SCREEN", "FN", "IF", "ELSE", "FOR", "IN", "WHILE", "RETURN", "NATIVE", "IMPORT",
		"AWAIT", "BACKGROUND",
		"TEXT", "BUTTON", "INPUT", "LIST", "IMAGE", "SWITCH", "SLIDER", "CHECKBOX", "PROGRESS",
		"NAVIGATE", "BACK", "TOAST", "VIBRATE", "NOTIFY", "RENDER",
		"HTTP", "STATE", "SHARE", "OPEN", "TIMER", "ALERT",
		"LPAREN", "RPAREN", "LBRACE", "RBRACE", "LBRACKET", "RBRACKET",
		"COMMA", "DOT", "COLON", "SEMI",
		"ASSIGN", "ARROW",
		"PLUS", "MINUS", "STAR", "SLASH", "MOD",
		"EQ", "NEQ", "LT", "GT", "LTE", "GTE",
		"AND", "OR", "NOT",
	}
	if int(t) < len(names) {
		return names[t]
	}
	return fmt.Sprintf("T%d", t)
}

type Tok struct {
	T    Type
	Lit  string
	Line int
	Col  int
}

var keywords = map[string]Type{
	"app": APP, "screen": SCREEN, "fn": FN, "if": IF, "else": ELSE,
	"for": FOR, "in": IN, "while": WHILE, "return": RETURN,
	"native": NATIVE, "import": IMPORT, "await": AWAIT, "background": BACKGROUND,
	"text": TEXT, "button": BUTTON, "input": INPUT, "list": LIST, "image": IMAGE,
	"switch": SWITCH, "slider": SLIDER, "checkbox": CHECKBOX, "progress": PROGRESS,
	"navigate": NAVIGATE, "back": BACK, "toast": TOAST, "vibrate": VIBRATE,
	"notify": NOTIFY, "render": RENDER, "http": HTTP, "state": STATE,
	"share": SHARE, "open": OPEN, "timer": TIMER, "alert": ALERT,
	"true": BOOL, "false": BOOL,
}

type L struct {
	src []byte
	pos int
	ln  int
	col int
}

func New(src []byte) *L { return &L{src: src, pos: 0, ln: 1, col: 1} }

func (l *L) All() []Tok {
	var toks []Tok
	for {
		t := l.Next()
		toks = append(toks, t)
		if t.T == EOF {
			break
		}
	}
	return toks
}

func (l *L) Next() Tok {
	for l.pos < len(l.src) {
		c := l.src[l.pos]

		// skip spaces/tabs
		if c == ' ' || c == '\t' || c == '\r' {
			l.advance()
			continue
		}

		// newline
		if c == '\n' {
			tok := Tok{T: NEWLINE, Lit: "\\n", Line: l.ln, Col: l.col}
			l.pos++
			l.ln++
			l.col = 1
			return tok
		}

		// comment
		if c == '/' && l.pos+1 < len(l.src) && l.src[l.pos+1] == '/' {
			for l.pos < len(l.src) && l.src[l.pos] != '\n' {
				l.advance()
			}
			continue
		}

		// string
		if c == '"' {
			return l.str()
		}

		// number
		if c >= '0' && c <= '9' {
			return l.num()
		}

		// identifier/keyword
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_' {
			return l.ident()
		}

		// arrow → (UTF-8)
		if l.pos+2 < len(l.src) && string(l.src[l.pos:l.pos+3]) == "→" {
			tok := Tok{T: ARROW, Lit: "->", Line: l.ln, Col: l.col}
			l.pos += 3
			l.col++
			return tok
		}

		// two-char
		if l.pos+1 < len(l.src) {
			two := string(l.src[l.pos : l.pos+2])
			switch two {
			case "->":
				tok := Tok{T: ARROW, Lit: "->", Line: l.ln, Col: l.col}
				l.pos += 2
				l.col += 2
				return tok
			case "==":
				tok := Tok{T: EQ, Lit: "==", Line: l.ln, Col: l.col}
				l.pos += 2
				l.col += 2
				return tok
			case "!=":
				tok := Tok{T: NEQ, Lit: "!=", Line: l.ln, Col: l.col}
				l.pos += 2
				l.col += 2
				return tok
			case "<=":
				tok := Tok{T: LTE, Lit: "<=", Line: l.ln, Col: l.col}
				l.pos += 2
				l.col += 2
				return tok
			case ">=":
				tok := Tok{T: GTE, Lit: ">=", Line: l.ln, Col: l.col}
				l.pos += 2
				l.col += 2
				return tok
			case "&&":
				tok := Tok{T: AND, Lit: "&&", Line: l.ln, Col: l.col}
				l.pos += 2
				l.col += 2
				return tok
			case "||":
				tok := Tok{T: OR, Lit: "||", Line: l.ln, Col: l.col}
				l.pos += 2
				l.col += 2
				return tok
			}
		}

		// single char
		line, col := l.ln, l.col
		l.advance()
		switch c {
		case '(':
			return Tok{LPAREN, "(", line, col}
		case ')':
			return Tok{RPAREN, ")", line, col}
		case '{':
			return Tok{LBRACE, "{", line, col}
		case '}':
			return Tok{RBRACE, "}", line, col}
		case '[':
			return Tok{LBRACKET, "[", line, col}
		case ']':
			return Tok{RBRACKET, "]", line, col}
		case ',':
			return Tok{COMMA, ",", line, col}
		case '.':
			return Tok{DOT, ".", line, col}
		case ':':
			return Tok{COLON, ":", line, col}
		case ';':
			return Tok{SEMI, ";", line, col}
		case '=':
			return Tok{ASSIGN, "=", line, col}
		case '+':
			return Tok{PLUS, "+", line, col}
		case '-':
			return Tok{MINUS, "-", line, col}
		case '*':
			return Tok{STAR, "*", line, col}
		case '/':
			return Tok{SLASH, "/", line, col}
		case '%':
			return Tok{MOD, "%", line, col}
		case '<':
			return Tok{LT, "<", line, col}
		case '>':
			return Tok{GT, ">", line, col}
		case '!':
			return Tok{NOT, "!", line, col}
		default:
			return Tok{ERROR, fmt.Sprintf("unexpected character %q", c), line, col}
		}
	}
	return Tok{EOF, "", l.ln, l.col}
}

func (l *L) str() Tok {
	line, col := l.ln, l.col
	l.advance() // skip "
	var b []byte
	for l.pos < len(l.src) && l.src[l.pos] != '"' && l.src[l.pos] != '\n' {
		if l.src[l.pos] == '\\' && l.pos+1 < len(l.src) {
			l.advance()
			switch l.src[l.pos] {
			case 'n':
				b = append(b, '\n')
			case 't':
				b = append(b, '\t')
			case '"':
				b = append(b, '"')
			case '\\':
				b = append(b, '\\')
			default:
				b = append(b, l.src[l.pos])
			}
		} else {
			b = append(b, l.src[l.pos])
		}
		l.advance()
	}
	if l.pos < len(l.src) && l.src[l.pos] == '"' {
		l.advance()
		return Tok{STRING, string(b), line, col}
	}
	return Tok{ERROR, "unterminated string literal", line, col}
}

func (l *L) num() Tok {
	line, col := l.ln, l.col
	start := l.pos
	for l.pos < len(l.src) && (l.src[l.pos] >= '0' && l.src[l.pos] <= '9' || l.src[l.pos] == '.') {
		l.advance()
	}
	return Tok{NUM, string(l.src[start:l.pos]), line, col}
}

func (l *L) ident() Tok {
	line, col := l.ln, l.col
	start := l.pos
	for l.pos < len(l.src) && ((l.src[l.pos] >= 'a' && l.src[l.pos] <= 'z') ||
		(l.src[l.pos] >= 'A' && l.src[l.pos] <= 'Z') ||
		(l.src[l.pos] >= '0' && l.src[l.pos] <= '9') || l.src[l.pos] == '_' || l.src[l.pos] == '-') {
		l.advance()
	}
	word := string(l.src[start:l.pos])
	if t, ok := keywords[word]; ok {
		return Tok{t, word, line, col}
	}
	return Tok{IDENT, word, line, col}
}

func (l *L) advance() {
	l.pos++
	l.col++
}
