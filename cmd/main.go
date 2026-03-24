package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Acorx/ion/internal/codegen"
	"github.com/Acorx/ion/internal/lexer"
	"github.com/Acorx/ion/internal/parser"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Print("ion — minimalist language for Android\n\nUsage:\n  ion <file.ion> [--out dir]   Build\n  ion transpile <file.ion>      Show Kotlin\n  ion -v                        Version\n")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "-v", "--version":
		fmt.Println("ion 0.1.0")
	case "transpile":
		if len(os.Args) < 3 {
			os.Exit(1)
		}
		prog := compile(os.Args[2])
		files := codegen.New(pkgName(prog)).Gen(prog)
		for n, c := range files {
			fmt.Printf("// === %s ===\n%s\n", n, c)
		}
	default:
		input := os.Args[1]
		out := "output"
		for i, a := range os.Args {
			if a == "--out" && i+1 < len(os.Args) {
				out = os.Args[i+1]
			}
		}
		prog := compile(input)
		files := codegen.New(pkgName(prog)).Gen(prog)
		for n, c := range files {
			p := filepath.Join(out, n)
			os.MkdirAll(filepath.Dir(p), 0755)
			os.WriteFile(p, []byte(c), 0644)
			fmt.Printf("  → %s\n", p)
		}
		fmt.Printf("\n✅ %d files → %s/\n", len(files), out)
	}
}

func compile(path string) *parser.Prog {
	src, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	toks := lexer.New(src).All()
	prog, errs := parser.New(toks).Parse()
	if len(errs) > 0 {
		for _, e := range errs {
			fmt.Fprintf(os.Stderr, "  ✗ %s\n", e)
		}
		os.Exit(1)
	}
	return prog
}

func pkgName(p *parser.Prog) string {
	n := strings.ToLower(p.Name)
	if n == "" {
		n = "ionapp"
	}
	return "com.example." + n
}
