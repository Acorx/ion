package lexer

import (
	"testing"
)

func TestString(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{`"hello"`, "hello"},
		{`"hello world"`, "hello world"},
		{`"hello\"world"`, "hello\"world"},
	}
	for _, tt := range tests {
		toks := New([]byte(tt.input)).All()
		if len(toks) < 2 {
			t.Errorf("expected at least 2 tokens, got %d", len(toks))
			continue
		}
		if toks[0].T != STRING {
			t.Errorf("expected STRING, got %s", toks[0].T)
		}
		if toks[0].Lit != tt.want {
			t.Errorf("expected %q, got %q", tt.want, toks[0].Lit)
		}
	}
}

func TestNumber(t *testing.T) {
	tests := []string{"42", "3.14", "0"}
	for _, tt := range tests {
		toks := New([]byte(tt)).All()
		if len(toks) < 1 {
			t.Errorf("expected at least 1 token")
			continue
		}
		if toks[0].T != NUM {
			t.Errorf("expected NUM for %s, got %s", tt, toks[0].T)
		}
		if toks[0].Lit != tt {
			t.Errorf("expected %q, got %q", tt, toks[0].Lit)
		}
	}
}

func TestKeyword(t *testing.T) {
	keywords := []struct {
		input string
		typ   Type
	}{
		{"app", APP},
		{"screen", SCREEN},
		{"fn", FN},
		{"if", IF},
		{"else", ELSE},
		{"for", FOR},
		{"while", WHILE},
		{"return", RETURN},
		{"text", TEXT},
		{"button", BUTTON},
		{"input", INPUT},
		{"toast", TOAST},
	}
	for _, tt := range keywords {
		toks := New([]byte(tt.input)).All()
		if len(toks) < 1 {
			t.Errorf("expected token for %s", tt.input)
			continue
		}
		if toks[0].T != tt.typ {
			t.Errorf("expected %s, got %s", tt.typ, toks[0].T)
		}
	}
}

func TestOperator(t *testing.T) {
	ops := []struct {
		input string
		typ   Type
	}{
		{"->", ARROW},
		{"==", EQ},
		{"!=", NEQ},
		{"<=", LTE},
		{">=", GTE},
		{"&&", AND},
		{"||", OR},
		{"=", ASSIGN},
		{"+", PLUS},
		{"-", MINUS},
		{"*", STAR},
	}
	for _, tt := range ops {
		toks := New([]byte(tt.input)).All()
		if len(toks) < 1 {
			t.Errorf("expected token for %s", tt.input)
			continue
		}
		if toks[0].T != tt.typ {
			t.Errorf("expected %s, got %s", tt.typ, toks[0].T)
		}
	}
}

func TestPunctuation(t *testing.T) {
	punct := []struct {
		input string
		typ   Type
	}{
		{"(", LPAREN},
		{")", RPAREN},
		{"{", LBRACE},
		{"}", RBRACE},
		{"[", LBRACKET},
		{"]", RBRACKET},
		{",", COMMA},
		{":", COLON},
	}
	for _, tt := range punct {
		toks := New([]byte(tt.input)).All()
		if len(toks) < 1 {
			t.Errorf("expected token for %s", tt.input)
			continue
		}
		if toks[0].T != tt.typ {
			t.Errorf("expected %s, got %s", tt.typ, toks[0].T)
		}
	}
}

func TestComment(t *testing.T) {
	input := `// this is a comment
text "hello"`
	toks := New([]byte(input)).All()
	// Should skip comment and return TEXT
	for _, tok := range toks {
		if tok.T == TEXT {
			return
		}
	}
	t.Errorf("expected TEXT token after comment")
}

func TestArrowUTF8(t *testing.T) {
	input := "button → toast"
	toks := New([]byte(input)).All()
	// Should tokenize UTF-8 arrow as ARROW
	found := false
	for _, tok := range toks {
		if tok.T == ARROW {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected ARROW token for UTF-8 arrow")
	}
}

func TestPosition(t *testing.T) {
	input := "app\n  text"
	toks := New([]byte(input)).All()
	if len(toks) < 2 {
		t.Errorf("expected at least 2 tokens")
		return
	}
	// First token on line 1
	if toks[0].Line != 1 {
		t.Errorf("expected line 1, got %d", toks[0].Line)
	}
	// Find TEXT token on line 2
	for _, tok := range toks {
		if tok.T == TEXT && tok.Line != 2 {
			t.Errorf("expected TEXT on line 2, got line %d", tok.Line)
		}
	}
}
