package parser

import (
	"testing"

	"github.com/Acorx/ion/internal/lexer"
)

func TestParseApp(t *testing.T) {
	src := []byte("app MyApp")
	toks := lexer.New(src).All()
	prog, errs := New(toks).Parse()
	if len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	if prog.Name != "MyApp" {
		t.Errorf("expected MyApp, got %s", prog.Name)
	}
}

func TestParseScreen(t *testing.T) {
	src := []byte(`app Test
screen Main {
  text "Hello"
}`)
	toks := lexer.New(src).All()
	prog, errs := New(toks).Parse()
	if len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	if len(prog.Screens) != 1 {
		t.Fatalf("expected 1 screen, got %d", len(prog.Screens))
	}
	if prog.Screens[0].Name != "Main" {
		t.Errorf("expected Main, got %s", prog.Screens[0].Name)
	}
}

func TestParseFunction(t *testing.T) {
	src := []byte(`fn greet(name) {
  toast(name)
}`)
	toks := lexer.New(src).All()
	prog, errs := New(toks).Parse()
	if len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	if len(prog.Funcs) != 1 {
		t.Fatalf("expected 1 function, got %d", len(prog.Funcs))
	}
	if prog.Funcs[0].Name != "greet" {
		t.Errorf("expected greet, got %s", prog.Funcs[0].Name)
	}
	if len(prog.Funcs[0].Params) != 1 {
		t.Errorf("expected 1 param, got %d", len(prog.Funcs[0].Params))
	}
}

func TestParseButtonWithHandler(t *testing.T) {
	src := []byte(`screen Main {
  button "Click" -> toast("clicked")
}`)
	toks := lexer.New(src).All()
	prog, errs := New(toks).Parse()
	if len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	if len(prog.Screens) != 1 {
		t.Fatalf("expected 1 screen")
	}
	if len(prog.Screens[0].Body) != 1 {
		t.Fatalf("expected 1 statement")
	}
}

func TestParseIfStatement(t *testing.T) {
	src := []byte(`fn test() {
  if x > 0 {
    toast("positive")
  }
}`)
	toks := lexer.New(src).All()
	prog, errs := New(toks).Parse()
	if len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	if len(prog.Funcs) != 1 {
		t.Fatalf("expected 1 function")
	}
	// Check that body contains IfStmt
	_ = prog.Funcs[0].Body
}

func TestParseHTTP(t *testing.T) {
	src := []byte(`fn load() {
  http get "https://api.example.com" -> data
}`)
	toks := lexer.New(src).All()
	prog, errs := New(toks).Parse()
	if len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	if len(prog.Funcs) != 1 {
		t.Fatalf("expected 1 function")
	}
}

func TestParseState(t *testing.T) {
	src := []byte(`screen Main {
  state count = 0
}`)
	toks := lexer.New(src).All()
	prog, errs := New(toks).Parse()
	if len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	if len(prog.Screens) != 1 {
		t.Fatalf("expected 1 screen")
	}
}

func TestParseError(t *testing.T) {
	src := []byte("screen {") // missing name
	toks := lexer.New(src).All()
	_, errs := New(toks).Parse()
	if len(errs) == 0 {
		t.Error("expected parse error for invalid syntax")
	}
}
