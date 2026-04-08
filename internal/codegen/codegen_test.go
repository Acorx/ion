package codegen

import (
	"strings"
	"testing"

	"github.com/Acorx/ion/internal/lexer"
	"github.com/Acorx/ion/internal/parser"
)

func parse(src string) *parser.Prog {
	toks := lexer.New([]byte(src)).All()
	prog, _ := parser.New(toks).Parse()
	return prog
}

func TestGenMainActivity(t *testing.T) {
	src := `app TestApp
screen Main {
  text "Hello"
}`
	prog := parse(src)
	files := New("com.example.testapp").Gen(prog)
	if _, ok := files["MainActivity.kt"]; !ok {
		t.Error("expected MainActivity.kt")
	}
	if !strings.Contains(files["MainActivity.kt"], "package com.example.testapp") {
		t.Error("expected package declaration")
	}
}

func TestGenScreenActivity(t *testing.T) {
	src := `app Test
screen Settings {
  text "Settings"
}`
	prog := parse(src)
	files := New("com.example.test").Gen(prog)
	if _, ok := files["SettingsActivity.kt"]; !ok {
		t.Error("expected SettingsActivity.kt")
	}
}

func TestGenLayout(t *testing.T) {
	src := `app Test
screen Main {
  text "Hello"
  button "Click"
}`
	prog := parse(src)
	files := New("com.example.test").Gen(prog)
	if _, ok := files["activity_main.xml"]; !ok {
		t.Error("expected activity_main.xml")
	}
	if !strings.Contains(files["activity_main.xml"], "LinearLayout") {
		t.Error("expected LinearLayout in layout")
	}
}

func TestGenManifest(t *testing.T) {
	src := `app TestApp
screen Main {}`
	prog := parse(src)
	files := New("com.example.testapp").Gen(prog)
	if _, ok := files["AndroidManifest.xml"]; !ok {
		t.Error("expected AndroidManifest.xml")
	}
	if !strings.Contains(files["AndroidManifest.xml"], "package=\"com.example.testapp\"") {
		t.Error("expected package in manifest")
	}
}

func TestGenBuildGradle(t *testing.T) {
	src := `app Test {}`
	prog := parse(src)
	files := New("com.example.test").Gen(prog)
	if _, ok := files["build.gradle"]; !ok {
		t.Error("expected build.gradle")
	}
}

func TestGenFunction(t *testing.T) {
	src := `app Test
screen Main {
  button "Click" -> greet()
}
fn greet() {
  toast("Hello!")
}`
	prog := parse(src)
	files := New("com.example.test").Gen(prog)
	// Function should be inlined in MainActivity
	if !strings.Contains(files["MainActivity.kt"], "fun greet()") {
		t.Error("expected greet function in MainActivity")
	}
}
