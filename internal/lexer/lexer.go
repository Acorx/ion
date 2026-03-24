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
}

var keywords = map[string]Type{
	"app": APP, "screen": SCREEN, "fn": FN, "if": IF, "else": ELSE,
	"for": FOR, "in": IN, "while": WHILE, "return": RETURN,
	"native": NATIVE, "import": IMPORT, "await": AWAIT, "background": BACKGROUND,
	"text": TEXT, "button": BUTTON, "input": INPUT, "list": LIST, "image": IMAGE,
	"switch": SWITCH, "slider": SLIDER, "checkbox": CHECKBOX, "progress": PROGRESS,
	"navigate": NAVIGATE, "back": BACK, "toast": TOAST, "vibrate": VIBRATE,
	"notify": NOTIFY, "render": RENDER,
	"true": BOOL, "false": BOOL,
}

type L struct {
	src []byte
	pos int
	ln  int
}

func New(src []byte) *L { return &L{src: src, pos: 0, ln: 1} }

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
			l.pos++
			continue
		}

		// newline
		if c == '\n' {
			l.pos++
			l.ln++
			return Tok{NEWLINE, "\\n", l.ln}
		}

		// comment
		if c == '/' && l.pos+1 < len(l.src) && l.src[l.pos+1] == '/' {
			for l.pos < len(l.src) && l.src[l.pos] != '\n' {
				l.pos++
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
			l.pos += 3
			return Tok{ARROW, "->", l.ln}
		}

		// two-char
		if l.pos+1 < len(l.src) {
			two := string(l.src[l.pos : l.pos+2])
			switch two {
			case "->":
				l.pos += 2
				return Tok{ARROW, "->", l.ln}
			case "==":
				l.pos += 2
				return Tok{EQ, "==", l.ln}
			case "!=":
				l.pos += 2
				return Tok{NEQ, "!=", l.ln}
			case "<=":
				l.pos += 2
				return Tok{LTE, "<=", l.ln}
			case ">=":
				l.pos += 2
				return Tok{GTE, ">=", l.ln}
			case "&&":
				l.pos += 2
				return Tok{AND, "&&", l.ln}
			case "||":
				l.pos += 2
				return Tok{OR, "||", l.ln}
			}
		}

		// single char
		l.pos++
		switch c {
		case '(':
			return Tok{LPAREN, "(", l.ln}
		case ')':
			return Tok{RPAREN, ")", l.ln}
		case '{':
			return Tok{LBRACE, "{", l.ln}
		case '}':
			return Tok{RBRACE, "}", l.ln}
		case '[':
			return Tok{LBRACKET, "[", l.ln}
		case ']':
			return Tok{RBRACKET, "]", l.ln}
		case ',':
			return Tok{COMMA, ",", l.ln}
		case '.':
			return Tok{DOT, ".", l.ln}
		case ':':
			return Tok{COLON, ":", l.ln}
		case ';':
			return Tok{SEMI, ";", l.ln}
		case '=':
			return Tok{ASSIGN, "=", l.ln}
		case '+':
			return Tok{PLUS, "+", l.ln}
		case '-':
			return Tok{MINUS, "-", l.ln}
		case '*':
			return Tok{STAR, "*", l.ln}
		case '/':
			return Tok{SLASH, "/", l.ln}
		case '%':
			return Tok{MOD, "%", l.ln}
		case '<':
			return Tok{LT, "<", l.ln}
		case '>':
			return Tok{GT, ">", l.ln}
		case '!':
			return Tok{NOT, "!", l.ln}
		default:
			return Tok{ERROR, string(c), l.ln}
		}
	}
	return Tok{EOF, "", l.ln}
}

func (l *L) str() Tok {
	l.pos++ // skip "
	start := l.pos
	for l.pos < len(l.src) && l.src[l.pos] != '"' && l.src[l.pos] != '\n' {
		if l.src[l.pos] == '\\' {
			l.pos++
		}
		l.pos++
	}
	s := string(l.src[start:l.pos])
	if l.pos < len(l.src) && l.src[l.pos] == '"' {
		l.pos++
	}
	return Tok{STRING, s, l.ln}
}

func (l *L) num() Tok {
	start := l.pos
	for l.pos < len(l.src) && (l.src[l.pos] >= '0' && l.src[l.pos] <= '9' || l.src[l.pos] == '.') {
		l.pos++
	}
	return Tok{NUM, string(l.src[start:l.pos]), l.ln}
}

func (l *L) ident() Tok {
	start := l.pos
	for l.pos < len(l.src) && ((l.src[l.pos] >= 'a' && l.src[l.pos] <= 'z') ||
		(l.src[l.pos] >= 'A' && l.src[l.pos] <= 'Z') ||
		(l.src[l.pos] >= '0' && l.src[l.pos] <= '9') || l.src[l.pos] == '_' || l.src[l.pos] == '-') {
		l.pos++
	}
	word := string(l.src[start:l.pos])
	if t, ok := keywords[word]; ok {
		return Tok{t, word, l.ln}
	}
	return Tok{IDENT, word, l.ln}
}
