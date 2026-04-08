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

// twoCharOps maps two-character sequences to their token types
var twoCharOps = map[string]Type{
	"->": ARROW, "==": EQ, "!=": NEQ, "<=": LTE, ">=": GTE, "&&": AND, "||": OR,
}

// singleCharOps maps single characters to their token types
var singleCharOps = map[byte]Type{
	'(': LPAREN, ')': RPAREN, '{': LBRACE, '}': RBRACE,
	'[': LBRACKET, ']': RBRACKET, ',': COMMA, '.': DOT,
	':': COLON, ';': SEMI, '=': ASSIGN,
	'+': PLUS, '-': MINUS, '*': STAR, '/': SLASH, '%': MOD,
	'<': LT, '>': GT, '!': NOT,
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

		if l.skipWhitespace(c) {
			continue
		}
		if tok := l.newline(c); tok != nil {
			return *tok
		}
		if l.skipComment(c) {
			continue
		}
		if tok := l.string(c); tok != nil {
			return *tok
		}
		if tok := l.number(c); tok != nil {
			return *tok
		}
		if tok := l.identifier(c); tok != nil {
			return *tok
		}
		if tok := l.utf8Arrow(); tok != nil {
			return *tok
		}
		if tok := l.twoChar(); tok != nil {
			return *tok
		}
		if tok := l.singleChar(c); tok != nil {
			return *tok
		}

		line, col := l.ln, l.col
		l.advance()
		return Tok{ERROR, fmt.Sprintf("unexpected character %q", c), line, col}
	}
	return Tok{EOF, "", l.ln, l.col}
}

func (l *L) skipWhitespace(c byte) bool {
	if c == ' ' || c == '\t' || c == '\r' {
		l.advance()
		return true
	}
	return false
}

func (l *L) newline(c byte) *Tok {
	if c == '\n' {
		tok := Tok{T: NEWLINE, Lit: "\\n", Line: l.ln, Col: l.col}
		l.pos++
		l.ln++
		l.col = 1
		return &tok
	}
	return nil
}

func (l *L) skipComment(c byte) bool {
	if c == '/' && l.pos+1 < len(l.src) && l.src[l.pos+1] == '/' {
		for l.pos < len(l.src) && l.src[l.pos] != '\n' {
			l.advance()
		}
		return true
	}
	return false
}

func (l *L) string(c byte) *Tok {
	if c != '"' {
		return nil
	}
	line, col := l.ln, l.col
	l.advance() // skip "
	var b []byte
	for l.pos < len(l.src) && l.src[l.pos] != '"' && l.src[l.pos] != '\n' {
		if l.src[l.pos] == '\\' && l.pos+1 < len(l.src) {
			l.advance()
			b = append(b, l.unescape(l.src[l.pos]))
		} else {
			b = append(b, l.src[l.pos])
		}
		l.advance()
	}
	if l.pos < len(l.src) && l.src[l.pos] == '"' {
		l.advance()
		return &Tok{STRING, string(b), line, col}
	}
	return &Tok{ERROR, "unterminated string literal", line, col}
}

func (l *L) unescape(c byte) byte {
	switch c {
	case 'n':
		return '\n'
	case 't':
		return '\t'
	case '"':
		return '"'
	case '\\':
		return '\\'
	default:
		return c
	}
}

func (l *L) number(c byte) *Tok {
	if c < '0' || c > '9' {
		return nil
	}
	line, col := l.ln, l.col
	start := l.pos
	for l.pos < len(l.src) && (l.src[l.pos] >= '0' && l.src[l.pos] <= '9' || l.src[l.pos] == '.') {
		l.advance()
	}
	return &Tok{NUM, string(l.src[start:l.pos]), line, col}
}

func (l *L) identifier(c byte) *Tok {
	if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_') {
		return nil
	}
	line, col := l.ln, l.col
	start := l.pos
	for l.pos < len(l.src) && l.isIdentChar(l.src[l.pos]) {
		l.advance()
	}
	word := string(l.src[start:l.pos])
	if t, ok := keywords[word]; ok {
		return &Tok{t, word, line, col}
	}
	return &Tok{IDENT, word, line, col}
}

func (l *L) isIdentChar(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == '-'
}

func (l *L) utf8Arrow() *Tok {
	if l.pos+2 < len(l.src) && string(l.src[l.pos:l.pos+3]) == "→" {
		tok := Tok{T: ARROW, Lit: "->", Line: l.ln, Col: l.col}
		l.pos += 3
		l.col++
		return &tok
	}
	return nil
}

func (l *L) twoChar() *Tok {
	if l.pos+1 >= len(l.src) {
		return nil
	}
	two := string(l.src[l.pos : l.pos+2])
	if t, ok := twoCharOps[two]; ok {
		tok := Tok{T: t, Lit: two, Line: l.ln, Col: l.col}
		l.pos += 2
		l.col += 2
		return &tok
	}
	return nil
}

func (l *L) singleChar(c byte) *Tok {
	if t, ok := singleCharOps[c]; ok {
		line, col := l.ln, l.col
		l.advance()
		return &Tok{t, string([]byte{c}), line, col}
	}
	return nil
}

func (l *L) advance() {
	l.pos++
	l.col++
}
