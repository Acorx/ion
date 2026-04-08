package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Acorx/ion/internal/codegen"
	"github.com/Acorx/ion/internal/formatter"
	"github.com/Acorx/ion/internal/lexer"
	"github.com/Acorx/ion/internal/parser"
	"github.com/fatih/color"
)

var (
	cyan   = color.New(color.FgCyan).SprintFunc()
	green  = color.New(color.FgGreen).SprintFunc()
	red    = color.New(color.FgRed).SprintFunc()
	yellow = color.New(color.FgYellow).SprintFunc()
	bold   = color.New(color.Bold).SprintFunc()
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	cmd := os.Args[1]

	switch cmd {
	case "-v", "--version":
		fmt.Println(cyan("ion"), "0.2.0")

	case "build":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, red("Error:"), "missing input file")
			usage()
			os.Exit(1)
		}
		input := os.Args[2]
		out := "output"
		for i, a := range os.Args {
			if a == "--out" && i+1 < len(os.Args) {
				out = os.Args[i+1]
			}
		}
		build(input, out)

	case "transpile":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, red("Error:"), "missing input file")
			os.Exit(1)
		}
		prog := compile(os.Args[2])
		files := codegen.New(pkgName(prog)).Gen(prog)
		for n, c := range files {
			fmt.Printf("%s\n%s\n", bold("// === "+n+" ==="), c)
		}

	case "format":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, red("Error:"), "missing input file")
			os.Exit(1)
		}
		prog := compile(os.Args[2])
		fmt.Println(formatter.Format(prog))

	case "check":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, red("Error:"), "missing input file")
			os.Exit(1)
		}
		src, err := os.ReadFile(os.Args[2])
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s %v\n", red("Error:"), err)
			os.Exit(1)
		}
		toks := lexer.New(src).All()
		prog, errs := parser.New(toks).Parse()
		if len(errs) > 0 {
			for _, e := range errs {
				fmt.Fprintf(os.Stderr, "%s %s\n", red("✗"), e)
			}
			os.Exit(1)
		}
		fmt.Println(green("✓"), os.Args[2], "valid")
		fmt.Printf("  %d screens, %d functions\n", len(prog.Screens), len(prog.Funcs))

	case "watch":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, red("Error:"), "missing input file")
			os.Exit(1)
		}
		input := os.Args[2]
		out := "output"
		for i, a := range os.Args {
			if a == "--out" && i+1 < len(os.Args) {
				out = os.Args[i+1]
			}
		}
		watch(input, out)

	default:
		// Legacy: ion <file.ion>
		input := cmd
		out := "output"
		for i, a := range os.Args {
			if a == "--out" && i+1 < len(os.Args) {
				out = os.Args[i+1]
			}
		}
		build(input, out)
	}
}

func usage() {
	fmt.Print(bold("ion"), " — ", cyan("minimalist language for Android"), "\n\n")
	fmt.Println(bold("Usage:"))
	fmt.Printf("  %s <file.ion> [--out dir]     Build app\n", cyan("ion"))
	fmt.Printf("  %s build <file.ion> [--out]   Build app\n", cyan("ion"))
	fmt.Printf("  %s transpile <file.ion>       Show Kotlin output\n", cyan("ion"))
	fmt.Printf("  %s format <file.ion>          Format source\n", cyan("ion"))
	fmt.Printf("  %s check <file.ion>           Validate without build\n", cyan("ion"))
	fmt.Printf("  %s watch <file.ion> [--out]   Rebuild on change\n", cyan("ion"))
	fmt.Printf("  %s -v                         Show version\n", cyan("ion"))
}

func build(input, out string) {
	start := time.Now()
	prog := compile(input)
	files := codegen.New(pkgName(prog)).Gen(prog)
	for n, c := range files {
		p := filepath.Join(out, n)
		os.MkdirAll(filepath.Dir(p), 0755)
		os.WriteFile(p, []byte(c), 0644)
		fmt.Printf("  %s %s\n", green("→"), p)
	}
	fmt.Printf("\n%s %d files → %s/ (%s)\n", 
		green("✓"), len(files), out, 
		yellow(fmt.Sprintf("%.0fms", time.Since(start).Seconds()*1000)))
}

func compile(path string) *parser.Prog {
	src, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %v\n", red("Error:"), err)
		os.Exit(1)
	}
	toks := lexer.New(src).All()
	prog, errs := parser.New(toks).Parse()
	if len(errs) > 0 {
		for _, e := range errs {
			fmt.Fprintf(os.Stderr, "%s %s\n", red("✗"), e)
		}
		os.Exit(1)
	}
	return prog
}

func watch(input, out string) {
	fmt.Printf("%s Watching %s...\n", cyan("⏳"), input)
	
	lastMod := time.Now()
	for {
		time.Sleep(500 * time.Millisecond)
		info, err := os.Stat(input)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s %v\n", red("Error:"), err)
			continue
		}
		if info.ModTime().After(lastMod) {
			lastMod = info.ModTime()
			fmt.Printf("\n%s %s changed, rebuilding...\n", yellow("⟳"), input)
			build(input, out)
		}
	}
}

func pkgName(p *parser.Prog) string {
	n := strings.ToLower(p.Name)
	if n == "" {
		n = "ionapp"
	}
	return "com.example." + n
}
