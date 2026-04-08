package formatter

import (
	"strconv"
	"strings"

	"github.com/Acorx/ion/internal/parser"
)

type F struct {
	b strings.Builder
}

func Format(p *parser.Prog) string {
	f := &F{}
	f.program(p)
	return f.b.String()
}

func (f *F) program(p *parser.Prog) {
	if p.Name != "" {
		f.line(0, "app "+p.Name)
		f.b.WriteByte('\n')
	}

	for _, n := range p.Natives {
		f.line(0, "native -> "+strconv.Quote(n))
	}
	if len(p.Natives) > 0 && (len(p.Screens) > 0 || len(p.Funcs) > 0) {
		f.b.WriteByte('\n')
	}

	for i, s := range p.Screens {
		if i > 0 {
			f.b.WriteByte('\n')
		}
		f.line(0, "screen "+s.Name+" {")
		f.stmts(s.Body, 1)
		f.line(0, "}")
	}

	if len(p.Screens) > 0 && len(p.Funcs) > 0 {
		f.b.WriteByte('\n')
	}

	for i, fn := range p.Funcs {
		if i > 0 {
			f.b.WriteByte('\n')
		}
		f.line(0, "fn "+fn.Name+"("+strings.Join(fn.Params, ", ")+") {")
		f.stmts(fn.Body, 1)
		f.line(0, "}")
	}
}

func (f *F) stmts(stmts []parser.Stmt, indent int) {
	for _, st := range stmts {
		f.stmt(st, indent)
	}
}

func (f *F) stmt(st parser.Stmt, indent int) {
	switch s := st.(type) {
	case parser.AssignStmt:
		f.line(indent, s.Name+" = "+f.expr(s.Val))
	case parser.NavStmt:
		f.line(indent, "navigate("+s.Target+")")
	case parser.BackStmt:
		f.line(indent, "back()")
	case parser.ToastStmt:
		f.line(indent, "toast("+f.expr(s.Msg)+")")
	case parser.VibStmt:
		f.line(indent, "vibrate()")
	case parser.NotifStmt:
		f.line(indent, "notify("+f.expr(s.Title)+", "+f.expr(s.Msg)+")")
	case parser.IfStmt:
		f.ifStmt(s, indent)
	case parser.ForStmt:
		f.line(indent, "for "+s.Var+" in "+f.expr(s.Iter)+" {")
		f.stmts(s.Body, indent+1)
		f.line(indent, "}")
	case parser.WhileStmt:
		f.line(indent, "while "+f.expr(s.Cond)+" {")
		f.stmts(s.Body, indent+1)
		f.line(indent, "}")
	case parser.RetStmt:
		f.line(indent, "return "+f.expr(s.Val))
	case parser.AwaitStmt:
		f.line(indent, "await "+f.expr(s.Call))
	case parser.BgStmt:
		f.line(indent, "background {")
		f.stmts(s.Body, indent+1)
		f.line(indent, "}")
	case parser.NativeStmt:
		f.line(indent, "native -> "+strconv.Quote(s.Code))
	case parser.HttpStmt:
		f.httpStmt(s, indent)
	case parser.StateStmt:
		f.stateStmt(s, indent)
	case parser.ShareStmt:
		f.line(indent, "share("+f.expr(s.Text)+")")
	case parser.OpenStmt:
		f.line(indent, "open("+f.expr(s.URL)+")")
	case parser.AlertStmt:
		f.line(indent, "alert("+f.expr(s.Title)+", "+f.expr(s.Msg)+")")
	case parser.ExprStmt:
		f.exprStmt(s, indent)
	}
}

func (f *F) ifStmt(s parser.IfStmt, indent int) {
	f.line(indent, "if "+f.expr(s.Cond)+" {")
	f.stmts(s.Then, indent+1)
	if len(s.Else) == 0 {
		f.line(indent, "}")
		return
	}
	f.line(indent, "} else {")
	f.stmts(s.Else, indent+1)
	f.line(indent, "}")
}

func (f *F) httpStmt(s parser.HttpStmt, indent int) {
	line := "http " + strings.ToLower(s.Method) + " " + f.expr(s.URL)
	if s.ResultVar != "" {
		line += " -> " + s.ResultVar
	}
	f.line(indent, line)
}

func (f *F) stateStmt(s parser.StateStmt, indent int) {
	line := "state " + s.Name
	if s.Initial != nil {
		line += " = " + f.expr(s.Initial)
	}
	f.line(indent, line)
}

func (f *F) exprStmt(s parser.ExprStmt, indent int) {
	if line, ok := f.component(s.E); ok {
		f.line(indent, line)
		return
	}
	f.line(indent, f.expr(s.E))
}

func (f *F) component(e parser.Expr) (string, bool) {
	call, ok := e.(parser.CallExpr)
	if !ok {
		return "", false
	}

	switch call.Fn {
	case "__text":
		return "text " + f.expr(call.Args[0]), true
	case "__image":
		return "image " + f.expr(call.Args[0]), true
	case "__list":
		return "list " + f.expr(call.Args[0]), true
	case "__button":
		return f.componentWithHandler("button", call.Args), true
	case "__input":
		return f.componentWithHandler("input", call.Args), true
	case "__switch":
		return f.componentWithHandler("switch", call.Args), true
	default:
		return "", false
	}
}

func (f *F) componentWithHandler(name string, args []parser.Expr) string {
	if len(args) == 0 {
		return name
	}
	line := name + " " + f.expr(args[0])
	if len(args) < 2 {
		return line
	}
	if blk, ok := args[1].(*parser.BlockExpr); ok && len(blk.Body) == 1 {
		return line + " -> " + f.inlineStmt(blk.Body[0])
	}
	return line
}

func (f *F) inlineStmt(st parser.Stmt) string {
	switch s := st.(type) {
	case parser.NavStmt:
		return "navigate(" + s.Target + ")"
	case parser.BackStmt:
		return "back()"
	case parser.ToastStmt:
		return "toast(" + f.expr(s.Msg) + ")"
	case parser.VibStmt:
		return "vibrate()"
	case parser.ShareStmt:
		return "share(" + f.expr(s.Text) + ")"
	case parser.OpenStmt:
		return "open(" + f.expr(s.URL) + ")"
	case parser.AlertStmt:
		return "alert(" + f.expr(s.Title) + ", " + f.expr(s.Msg) + ")"
	case parser.ExprStmt:
		return f.expr(s.E)
	default:
		return f.expr(parser.CallExpr{Fn: "__unsupported_inline", Args: nil})
	}
}

func (f *F) expr(e parser.Expr) string {
	if e == nil {
		return "null"
	}
	switch ex := e.(type) {
	case parser.StrExpr:
		return strconv.Quote(ex.V)
	case parser.NumExpr:
		return ex.V
	case parser.BoolExpr:
		if ex.V {
			return "true"
		}
		return "false"
	case parser.IdentExpr:
		return ex.Name
	case parser.BinExpr:
		return f.expr(ex.L) + " " + ex.Op + " " + f.expr(ex.R)
	case parser.UnExpr:
		return ex.Op + f.expr(ex.R)
	case parser.CallExpr:
		return ex.Fn + "(" + f.joinExprs(ex.Args) + ")"
	case parser.MethodExpr:
		return f.expr(ex.Obj) + "." + ex.Method + "(" + f.joinExprs(ex.Args) + ")"
	case parser.FieldExpr:
		return f.expr(ex.Obj) + "." + ex.Field
	case parser.IdxExpr:
		return f.expr(ex.Obj) + "[" + f.expr(ex.Idx) + "]"
	case parser.ArrExpr:
		return "[" + f.joinExprs(ex.Elems) + "]"
	default:
		return "null"
	}
}

func (f *F) joinExprs(exprs []parser.Expr) string {
	parts := make([]string, len(exprs))
	for i, e := range exprs {
		parts[i] = f.expr(e)
	}
	return strings.Join(parts, ", ")
}

func (f *F) line(indent int, s string) {
	f.b.WriteString(strings.Repeat(" ", indent))
	f.b.WriteString(s)
	f.b.WriteByte('\n')
}
