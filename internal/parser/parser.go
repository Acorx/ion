package parser

import (
	"fmt"
	"strings"

	"github.com/Acorx/ion/internal/lexer"
)

// AST
type Prog struct {
	Name    string
	Screens []Screen
	Funcs   []Func
	Natives []string
}

type Screen struct {
	Name string
	Body []Stmt
}

type Func struct {
	Name   string
	Params []string
	Body   []Stmt
}

// Stmt
type Stmt interface{ stmt() }

type AssignStmt struct {
	Name string
	Val  Expr
}
type NavStmt struct{ Target string }
type BackStmt struct{}
type ToastStmt struct{ Msg Expr }
type VibStmt struct{}
type NotifStmt struct{ Title, Msg Expr }
type RenderStmt struct{ Comp Node }
type IfStmt struct {
	Cond       Expr
	Then, Else []Stmt
}
type ForStmt struct {
	Var  string
	Iter Expr
	Body []Stmt
}
type WhileStmt struct {
	Cond Expr
	Body []Stmt
}
type RetStmt struct{ Val Expr }
type AwaitStmt struct{ Call Expr }
type BgStmt struct{ Body []Stmt }
type NativeStmt struct{ Code string }
type ExprStmt struct{ E Expr }
type HttpStmt struct {
	Method    string
	URL       Expr
	Body      Expr
	ResultVar string
}
type StateStmt struct {
	Name    string
	Initial Expr
}
type ShareStmt struct{ Text Expr }
type OpenStmt struct{ URL Expr }
type AlertStmt struct{ Title, Msg Expr }

func (AssignStmt) stmt() {}
func (NavStmt) stmt()    {}
func (BackStmt) stmt()   {}
func (ToastStmt) stmt()  {}
func (VibStmt) stmt()    {}
func (NotifStmt) stmt()  {}
func (RenderStmt) stmt() {}
func (IfStmt) stmt()     {}
func (ForStmt) stmt()    {}
func (WhileStmt) stmt()  {}
func (RetStmt) stmt()    {}
func (AwaitStmt) stmt()  {}
func (BgStmt) stmt()     {}
func (NativeStmt) stmt() {}
func (ExprStmt) stmt()   {}
func (HttpStmt) stmt()   {}
func (StateStmt) stmt()  {}
func (ShareStmt) stmt()  {}
func (OpenStmt) stmt()   {}
func (AlertStmt) stmt()  {}

// UI Node
type Node interface{ node() }

type TextNode struct{ Val Expr }
type BtnNode struct {
	Label   Expr
	Handler []Stmt
}
type InputNode struct {
	Hint    Expr
	Handler []Stmt
}
type ListNode struct {
	Items Expr
	Item  []Node
}
type ImgNode struct{ Src Expr }
type SwitchNode struct {
	Label   Expr
	Handler []Stmt
}

func (TextNode) node()   {}
func (BtnNode) node()    {}
func (InputNode) node()  {}
func (ListNode) node()   {}
func (ImgNode) node()    {}
func (SwitchNode) node() {}

// Expr
type Expr interface{ expr() }

type StrExpr struct{ V string }
type NumExpr struct{ V string }
type BoolExpr struct{ V bool }
type IdentExpr struct{ Name string }
type BinExpr struct {
	L  Expr
	Op string
	R  Expr
}
type UnExpr struct {
	Op string
	R  Expr
}
type CallExpr struct {
	Fn   string
	Args []Expr
}
type MethodExpr struct {
	Obj    Expr
	Method string
	Args   []Expr
}
type FieldExpr struct {
	Obj   Expr
	Field string
}
type IdxExpr struct {
	Obj Expr
	Idx Expr
}
type ArrExpr struct{ Elems []Expr }
type MapExpr struct{ Entries map[string]Expr }
type BlockExpr struct{ Body []Stmt }

func (StrExpr) expr()    {}
func (NumExpr) expr()    {}
func (BoolExpr) expr()   {}
func (IdentExpr) expr()  {}
func (BinExpr) expr()    {}
func (UnExpr) expr()     {}
func (CallExpr) expr()   {}
func (MethodExpr) expr() {}
func (FieldExpr) expr()  {}
func (IdxExpr) expr()    {}
func (ArrExpr) expr()    {}
func (MapExpr) expr()    {}
func (BlockExpr) expr()  {}

// Parser
type P struct {
	toks []lexer.Tok
	pos  int
	errs []string
}

func New(toks []lexer.Tok) *P { return &P{toks: toks} }

func (p *P) Parse() (*Prog, []string) {
	pr := &Prog{}
	p.skipNL()
	if p.is(lexer.APP) {
		p.next()
		pr.Name = p.expect(lexer.IDENT).Lit
	}
	p.skipNL()
	for !p.done() {
		switch {
		case p.is(lexer.SCREEN):
			pr.Screens = append(pr.Screens, p.screen())
		case p.is(lexer.FN):
			pr.Funcs = append(pr.Funcs, p.fn())
		case p.is(lexer.NATIVE):
			p.next()
			p.expect(lexer.ARROW)
			pr.Natives = append(pr.Natives, p.expect(lexer.STRING).Lit)
		default:
			p.err("expected screen/fn/native, got " + p.cur().T.String())
			p.next()
		}
		p.skipNL()
	}
	return pr, p.errs
}

func (p *P) screen() Screen {
	p.next() // screen
	s := Screen{Name: p.expect(lexer.IDENT).Lit}
	p.expect(lexer.LBRACE)
	for !p.is(lexer.RBRACE) && !p.done() {
		p.skipNL()
		if p.is(lexer.RBRACE) {
			break
		}
		s.Body = append(s.Body, p.stmtOrComp())
		p.skipNL()
	}
	p.expect(lexer.RBRACE)
	return s
}

func (p *P) fn() Func {
	p.next() // fn
	f := Func{Name: p.expect(lexer.IDENT).Lit}
	p.expect(lexer.LPAREN)
	for !p.is(lexer.RPAREN) && !p.done() {
		f.Params = append(f.Params, p.expect(lexer.IDENT).Lit)
		if p.is(lexer.COMMA) {
			p.next()
		}
	}
	p.expect(lexer.RPAREN)
	f.Body = p.block()
	return f
}

func (p *P) stmtOrComp() Stmt {
	// Check if it's a UI component
	switch p.cur().T {
	case lexer.TEXT:
		return p.textComp()
	case lexer.BUTTON:
		return p.btnComp()
	case lexer.INPUT:
		return p.inputComp()
	case lexer.LIST:
		return p.listComp()
	case lexer.IMAGE:
		return p.imgComp()
	case lexer.SWITCH:
		return p.switchComp()
	case lexer.PROGRESS:
		return p.progressComp()
	}
	return p.stmt()
}

func (p *P) textComp() Stmt {
	p.next() // text
	return ExprStmt{E: CallExpr{Fn: "__text", Args: []Expr{p.expr()}}}
}

func (p *P) btnComp() Stmt {
	p.next() // button
	label := p.expr()
	var args []Expr = []Expr{label}
	if p.is(lexer.ARROW) {
		p.next()
		st := p.stmt()
		args = append(args, &BlockExpr{Body: []Stmt{st}})
	} else if p.is(lexer.LBRACE) {
		body := p.block()
		args = append(args, &BlockExpr{Body: body})
	}
	return ExprStmt{E: CallExpr{Fn: "__button", Args: args}}
}

func (p *P) inputComp() Stmt {
	p.next() // input
	hint := p.expr()
	var args []Expr = []Expr{hint}
	if p.is(lexer.ARROW) {
		p.next()
		st := p.stmt()
		args = append(args, &BlockExpr{Body: []Stmt{st}})
	} else if p.is(lexer.LBRACE) {
		body := p.block()
		args = append(args, &BlockExpr{Body: body})
	}
	return ExprStmt{E: CallExpr{Fn: "__input", Args: args}}
}

func (p *P) listComp() Stmt {
	p.next() // list
	items := p.expr()
	if p.is(lexer.LBRACE) {
		_ = p.block()
	}
	return ExprStmt{E: CallExpr{Fn: "__list", Args: []Expr{items}}}
}

func (p *P) imgComp() Stmt {
	p.next() // image
	return ExprStmt{E: CallExpr{Fn: "__image", Args: []Expr{p.expr()}}}
}

func (p *P) switchComp() Stmt {
	p.next() // switch
	label := p.expr()
	var args []Expr = []Expr{label}
	if p.is(lexer.ARROW) {
		p.next()
		st := p.stmt()
		args = append(args, &BlockExpr{Body: []Stmt{st}})
	} else if p.is(lexer.LBRACE) {
		body := p.block()
		args = append(args, &BlockExpr{Body: body})
	}
	return ExprStmt{E: CallExpr{Fn: "__switch", Args: args}}
}

func (p *P) progressComp() Stmt {
	p.next() // progress
	return ExprStmt{E: CallExpr{Fn: "__progress"}}
}

func (p *P) stmt() Stmt {
	switch p.cur().T {
	case lexer.NAVIGATE:
		p.next()
		p.expect(lexer.LPAREN)
		t := p.expect(lexer.IDENT).Lit
		p.expect(lexer.RPAREN)
		return NavStmt{Target: t}

	case lexer.BACK:
		p.next()
		if p.is(lexer.LPAREN) {
			p.next()
			p.expect(lexer.RPAREN)
		}
		return BackStmt{}

	case lexer.TOAST:
		p.next()
		p.expect(lexer.LPAREN)
		e := p.expr()
		p.expect(lexer.RPAREN)
		return ToastStmt{Msg: e}

	case lexer.VIBRATE:
		p.next()
		return VibStmt{}

	case lexer.NOTIFY:
		p.next()
		p.expect(lexer.LPAREN)
		title := p.expr()
		p.expect(lexer.COMMA)
		msg := p.expr()
		p.expect(lexer.RPAREN)
		return NotifStmt{Title: title, Msg: msg}

	case lexer.HTTP:
		return p.httpStmt()

	case lexer.STATE:
		p.next()
		name := p.expect(lexer.IDENT).Lit
		var initial Expr
		if p.is(lexer.ASSIGN) {
			p.next()
			initial = p.expr()
		}
		return StateStmt{Name: name, Initial: initial}

	case lexer.SHARE:
		p.next()
		p.expect(lexer.LPAREN)
		e := p.expr()
		p.expect(lexer.RPAREN)
		return ShareStmt{Text: e}

	case lexer.OPEN:
		p.next()
		p.expect(lexer.LPAREN)
		e := p.expr()
		p.expect(lexer.RPAREN)
		return OpenStmt{URL: e}

	case lexer.ALERT:
		p.next()
		p.expect(lexer.LPAREN)
		title := p.expr()
		p.expect(lexer.COMMA)
		msg := p.expr()
		p.expect(lexer.RPAREN)
		return AlertStmt{Title: title, Msg: msg}

	case lexer.IF:
		return p.ifStmt()

	case lexer.FOR:
		return p.forStmt()

	case lexer.WHILE:
		return p.whileStmt()

	case lexer.RETURN:
		p.next()
		return RetStmt{Val: p.expr()}

	case lexer.AWAIT:
		p.next()
		return AwaitStmt{Call: p.expr()}

	case lexer.BACKGROUND:
		p.next()
		return BgStmt{Body: p.block()}

	case lexer.NATIVE:
		p.next()
		p.expect(lexer.ARROW)
		return NativeStmt{Code: p.expect(lexer.STRING).Lit}

	case lexer.IDENT:
		name := p.next().Lit
		if p.is(lexer.ASSIGN) {
			p.next()
			return AssignStmt{Name: name, Val: p.expr()}
		}
		// expression statement
		p.pos-- // put back
		return ExprStmt{E: p.expr()}

	default:
		return ExprStmt{E: p.expr()}
	}
}

func (p *P) httpStmt() Stmt {
	p.next()                                             // http
	method := strings.ToUpper(p.expect(lexer.IDENT).Lit) // get/post/put/delete
	url := p.expr()
	var body Expr
	var resultVar string
	// Optional: -> variable or { body }
	if p.is(lexer.ARROW) {
		p.next()
		if p.is(lexer.IDENT) {
			resultVar = p.next().Lit
		}
	} else if p.is(lexer.LBRACE) {
		// body for POST
		_ = p.block()
	}
	return HttpStmt{Method: method, URL: url, Body: body, ResultVar: resultVar}
}

func (p *P) ifStmt() Stmt {
	p.next() // if
	cond := p.expr()
	then := p.block()
	var els []Stmt
	if p.is(lexer.ELSE) {
		p.next()
		els = p.block()
	}
	return IfStmt{Cond: cond, Then: then, Else: els}
}

func (p *P) forStmt() Stmt {
	p.next() // for
	v := p.expect(lexer.IDENT).Lit
	p.expect(lexer.IN)
	iter := p.expr()
	return ForStmt{Var: v, Iter: iter, Body: p.block()}
}

func (p *P) whileStmt() Stmt {
	p.next() // while
	cond := p.expr()
	return WhileStmt{Cond: cond, Body: p.block()}
}

func (p *P) block() []Stmt {
	p.expect(lexer.LBRACE)
	var s []Stmt
	for !p.is(lexer.RBRACE) && !p.done() {
		p.skipNL()
		if p.is(lexer.RBRACE) {
			break
		}
		s = append(s, p.stmt())
		p.skipNL()
	}
	p.expect(lexer.RBRACE)
	return s
}

// Expressions
func (p *P) expr() Expr { return p.or() }

func (p *P) or() Expr {
	l := p.and()
	for p.is(lexer.OR) {
		op := p.next().Lit
		r := p.and()
		l = BinExpr{L: l, Op: op, R: r}
	}
	return l
}

func (p *P) and() Expr {
	l := p.eq()
	for p.is(lexer.AND) {
		op := p.next().Lit
		r := p.eq()
		l = BinExpr{L: l, Op: op, R: r}
	}
	return l
}

func (p *P) eq() Expr {
	l := p.cmp()
	for p.is(lexer.EQ) || p.is(lexer.NEQ) {
		op := p.next().Lit
		r := p.cmp()
		l = BinExpr{L: l, Op: op, R: r}
	}
	return l
}

func (p *P) cmp() Expr {
	l := p.add()
	for p.is(lexer.LT) || p.is(lexer.GT) || p.is(lexer.LTE) || p.is(lexer.GTE) {
		op := p.next().Lit
		r := p.add()
		l = BinExpr{L: l, Op: op, R: r}
	}
	return l
}

func (p *P) add() Expr {
	l := p.mul()
	for p.is(lexer.PLUS) || p.is(lexer.MINUS) {
		op := p.next().Lit
		r := p.mul()
		l = BinExpr{L: l, Op: op, R: r}
	}
	return l
}

func (p *P) mul() Expr {
	l := p.unary()
	for p.is(lexer.STAR) || p.is(lexer.SLASH) || p.is(lexer.MOD) {
		op := p.next().Lit
		r := p.unary()
		l = BinExpr{L: l, Op: op, R: r}
	}
	return l
}

func (p *P) unary() Expr {
	if p.is(lexer.NOT) || p.is(lexer.MINUS) {
		op := p.next().Lit
		return UnExpr{Op: op, R: p.primary()}
	}
	return p.postfix()
}

func (p *P) postfix() Expr {
	e := p.primary()
	for {
		switch {
		case p.is(lexer.DOT):
			p.next()
			f := p.expect(lexer.IDENT).Lit
			if p.is(lexer.LPAREN) {
				e = MethodExpr{Obj: e, Method: f, Args: p.args()}
			} else {
				e = FieldExpr{Obj: e, Field: f}
			}
		case p.is(lexer.LBRACKET):
			p.next()
			idx := p.expr()
			p.expect(lexer.RBRACKET)
			e = IdxExpr{Obj: e, Idx: idx}
		default:
			return e
		}
	}
}

func (p *P) primary() Expr {
	switch p.cur().T {
	case lexer.STRING:
		return StrExpr{V: p.next().Lit}
	case lexer.NUM:
		return NumExpr{V: p.next().Lit}
	case lexer.BOOL:
		return BoolExpr{V: p.next().Lit == "true"}
	case lexer.IDENT:
		name := p.next().Lit
		if p.is(lexer.LPAREN) {
			return CallExpr{Fn: name, Args: p.args()}
		}
		return IdentExpr{Name: name}
	case lexer.LBRACKET:
		p.next()
		var elems []Expr
		for !p.is(lexer.RBRACKET) && !p.done() {
			elems = append(elems, p.expr())
			if p.is(lexer.COMMA) {
				p.next()
			}
		}
		p.expect(lexer.RBRACKET)
		return ArrExpr{Elems: elems}
	case lexer.LPAREN:
		p.next()
		e := p.expr()
		p.expect(lexer.RPAREN)
		return e
	default:
		p.err("unexpected " + p.cur().T.String())
		p.next()
		return nil
	}
}

func (p *P) args() []Expr {
	p.expect(lexer.LPAREN)
	var a []Expr
	for !p.is(lexer.RPAREN) && !p.done() {
		a = append(a, p.expr())
		if p.is(lexer.COMMA) {
			p.next()
		}
	}
	p.expect(lexer.RPAREN)
	return a
}

// Helpers
func (p *P) cur() lexer.Tok {
	if p.done() {
		return lexer.Tok{T: lexer.EOF}
	}
	return p.toks[p.pos]
}

func (p *P) next() lexer.Tok {
	t := p.cur()
	p.pos++
	return t
}

func (p *P) is(t lexer.Type) bool { return p.cur().T == t }

func (p *P) expect(t lexer.Type) lexer.Tok {
	if p.is(t) {
		return p.next()
	}
	p.err(fmt.Sprintf("expected %s, got %s", t, p.cur().T))
	return lexer.Tok{Lit: ""}
}

func (p *P) done() bool { return p.pos >= len(p.toks) || p.toks[p.pos].T == lexer.EOF }

func (p *P) skipNL() {
	for p.is(lexer.NEWLINE) || p.is(lexer.SEMI) {
		p.next()
	}
}

func (p *P) err(msg string) {
	p.errs = append(p.errs, fmt.Sprintf("line %d: %s", p.cur().Line, msg))
}
