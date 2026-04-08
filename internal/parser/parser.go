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
	Cond      Expr
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
func (NavStmt) stmt()   {}
func (BackStmt) stmt()  {}
func (ToastStmt) stmt() {}
func (VibStmt) stmt()   {}
func (NotifStmt) stmt() {}
func (RenderStmt) stmt() {}
func (IfStmt) stmt()    {}
func (ForStmt) stmt()   {}
func (WhileStmt) stmt() {}
func (RetStmt) stmt()   {}
func (AwaitStmt) stmt() {}
func (BgStmt) stmt()    {}
func (NativeStmt) stmt() {}
func (ExprStmt) stmt()  {}
func (HttpStmt) stmt()  {}
func (StateStmt) stmt() {}
func (ShareStmt) stmt() {}
func (OpenStmt) stmt()  {}
func (AlertStmt) stmt() {}

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

func (StrExpr) expr()   {}
func (NumExpr) expr()   {}
func (BoolExpr) expr()  {}
func (IdentExpr) expr() {}
func (BinExpr) expr()   {}
func (UnExpr) expr()    {}
func (CallExpr) expr()  {}
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
	args := []Expr{label}
	args = p.maybeHandler(args)
	return ExprStmt{E: CallExpr{Fn: "__button", Args: args}}
}

func (p *P) inputComp() Stmt {
	p.next() // input
	hint := p.expr()
	args := []Expr{hint}
	args = p.maybeHandler(args)
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
	args := []Expr{label}
	args = p.maybeHandler(args)
	return ExprStmt{E: CallExpr{Fn: "__switch", Args: args}}
}

func (p *P) maybeHandler(args []Expr) []Expr {
	if p.is(lexer.ARROW) {
		p.next()
		st := p.stmt()
		return append(args, &BlockExpr{Body: []Stmt{st}})
	} else if p.is(lexer.LBRACE) {
		body := p.block()
		return append(args, &BlockExpr{Body: body})
	}
	return args
}

func (p *P) stmt() Stmt {
	tok := p.cur().T
	switch tok {
	case lexer.NAVIGATE:
		return p.parseNavigate()
	case lexer.BACK:
		return p.parseBack()
	case lexer.TOAST:
		return p.parseToast()
	case lexer.VIBRATE:
		p.next()
		return VibStmt{}
	case lexer.NOTIFY:
		return p.parseNotify()
	case lexer.HTTP:
		return p.httpStmt()
	case lexer.STATE:
		return p.parseState()
	case lexer.SHARE:
		return p.parseShare()
	case lexer.OPEN:
		return p.parseOpen()
	case lexer.ALERT:
		return p.parseAlert()
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
		p.pos-- // put back
		return ExprStmt{E: p.expr()}
	default:
		return ExprStmt{E: p.expr()}
	}
}

func (p *P) parseNavigate() Stmt {
	p.next()
	p.expect(lexer.LPAREN)
	t := p.expect(lexer.IDENT).Lit
	p.expect(lexer.RPAREN)
	return NavStmt{Target: t}
}

func (p *P) parseBack() Stmt {
	p.next()
	if p.is(lexer.LPAREN) {
		p.next()
		p.expect(lexer.RPAREN)
	}
	return BackStmt{}
}

func (p *P) parseToast() Stmt {
	p.next()
	p.expect(lexer.LPAREN)
	e := p.expr()
	p.expect(lexer.RPAREN)
	return ToastStmt{Msg: e}
}

func (p *P) parseNotify() Stmt {
	p.next()
	p.expect(lexer.LPAREN)
	title := p.expr()
	p.expect(lexer.COMMA)
	msg := p.expr()
	p.expect(lexer.RPAREN)
	return NotifStmt{Title: title, Msg: msg}
}

func (p *P) parseState() Stmt {
	p.next()
	name := p.expect(lexer.IDENT).Lit
	var initial Expr
	if p.is(lexer.ASSIGN) {
		p.next()
		initial = p.expr()
	}
	return StateStmt{Name: name, Initial: initial}
}

func (p *P) parseShare() Stmt {
	p.next()
	p.expect(lexer.LPAREN)
	e := p.expr()
	p.expect(lexer.RPAREN)
	return ShareStmt{Text: e}
}

func (p *P) parseOpen() Stmt {
	p.next()
	p.expect(lexer.LPAREN)
	e := p.expr()
	p.expect(lexer.RPAREN)
	return OpenStmt{URL: e}
}

func (p *P) parseAlert() Stmt {
	p.next()
	p.expect(lexer.LPAREN)
	title := p.expr()
	p.expect(lexer.COMMA)
	msg := p.expr()
	p.expect(lexer.RPAREN)
	return AlertStmt{Title: title, Msg: msg}
}

func (p *P) httpStmt() Stmt {
	p.next() // http
	method := strings.ToUpper(p.expect(lexer.IDENT).Lit)
	url := p.expr()
	var body Expr
	var resultVar string
	if p.is(lexer.ARROW) {
		p.next()
		if p.is(lexer.IDENT) {
			resultVar = p.next().Lit
		}
	} else if p.is(lexer.LBRACE) {
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
	e := p.and()
	for p.is(lexer.OR) {
		op := p.next().Lit
		e = BinExpr{L: e, Op: op, R: p.and()}
	}
	return e
}

func (p *P) and() Expr {
	e := p.eq()
	for p.is(lexer.AND) {
		op := p.next().Lit
		e = BinExpr{L: e, Op: op, R: p.eq()}
	}
	return e
}

func (p *P) eq() Expr {
	e := p.cmp()
	for {
		var op string
		switch {
		case p.is(lexer.EQ):
			op = p.next().Lit
		case p.is(lexer.NEQ):
			op = p.next().Lit
		default:
			return e
		}
		e = BinExpr{L: e, Op: op, R: p.cmp()}
	}
}

func (p *P) cmp() Expr {
	e := p.add()
	for {
		var op string
		switch {
		case p.is(lexer.LT):
			op = p.next().Lit
		case p.is(lexer.GT):
			op = p.next().Lit
		case p.is(lexer.LTE):
			op = p.next().Lit
		case p.is(lexer.GTE):
			op = p.next().Lit
		default:
			return e
		}
		e = BinExpr{L: e, Op: op, R: p.add()}
	}
}

func (p *P) add() Expr {
	e := p.mul()
	for {
		var op string
		switch {
		case p.is(lexer.PLUS):
			op = p.next().Lit
		case p.is(lexer.MINUS):
			op = p.next().Lit
		default:
			return e
		}
		e = BinExpr{L: e, Op: op, R: p.mul()}
	}
}

func (p *P) mul() Expr {
	e := p.unary()
	for {
		var op string
		switch {
		case p.is(lexer.STAR):
			op = p.next().Lit
		case p.is(lexer.SLASH):
			op = p.next().Lit
		case p.is(lexer.MOD):
			op = p.next().Lit
		default:
			return e
		}
		e = BinExpr{L: e, Op: op, R: p.unary()}
	}
}

func (p *P) unary() Expr {
	if p.is(lexer.NOT) || p.is(lexer.MINUS) {
		op := p.next().Lit
		return UnExpr{Op: op, R: p.unary()}
	}
	return p.postfix()
}

func (p *P) postfix() Expr {
	e := p.primary()
	for {
		switch {
		case p.is(lexer.DOT):
			p.next()
			field := p.expect(lexer.IDENT).Lit
			if p.is(lexer.LPAREN) {
				args := p.args()
				e = MethodExpr{Obj: e, Method: field, Args: args}
			} else {
				e = FieldExpr{Obj: e, Field: field}
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
	switch {
	case p.is(lexer.STRING):
		return StrExpr{V: p.next().Lit}
	case p.is(lexer.NUM):
		return NumExpr{V: p.next().Lit}
	case p.is(lexer.BOOL):
		v := p.next().Lit == "true"
		return BoolExpr{V: v}
	case p.is(lexer.IDENT):
		name := p.next().Lit
		if p.is(lexer.LPAREN) {
			args := p.args()
			return CallExpr{Fn: name, Args: args}
		}
		return IdentExpr{Name: name}
	case p.is(lexer.LPAREN):
		p.next()
		e := p.expr()
		p.expect(lexer.RPAREN)
		return e
	case p.is(lexer.LBRACKET):
		return p.array()
	case p.is(lexer.LBRACE):
		return p.blockExpr()
	default:
		p.err("unexpected token: " + p.cur().T.String())
		return nil
	}
}

func (p *P) args() []Expr {
	p.expect(lexer.LPAREN)
	var args []Expr
	for !p.is(lexer.RPAREN) && !p.done() {
		args = append(args, p.expr())
		if p.is(lexer.COMMA) {
			p.next()
		}
	}
	p.expect(lexer.RPAREN)
	return args
}

func (p *P) array() Expr {
	p.expect(lexer.LBRACKET)
	var elems []Expr
	for !p.is(lexer.RBRACKET) && !p.done() {
		elems = append(elems, p.expr())
		if p.is(lexer.COMMA) {
			p.next()
		}
	}
	p.expect(lexer.RBRACKET)
	return ArrExpr{Elems: elems}
}

func (p *P) blockExpr() Expr {
	p.expect(lexer.LBRACE)
	var stmts []Stmt
	for !p.is(lexer.RBRACE) && !p.done() {
		p.skipNL()
		if p.is(lexer.RBRACE) {
			break
		}
		stmts = append(stmts, p.stmt())
		p.skipNL()
	}
	p.expect(lexer.RBRACE)
	return BlockExpr{Body: stmts}
}

// Helpers
func (p *P) cur() lexer.Tok {
	if p.pos >= len(p.toks) {
		return lexer.Tok{T: lexer.EOF}
	}
	return p.toks[p.pos]
}

func (p *P) next() lexer.Tok {
	if p.pos >= len(p.toks) {
		return lexer.Tok{T: lexer.EOF}
	}
	t := p.toks[p.pos]
	p.pos++
	return t
}

func (p *P) is(t lexer.Type) bool { return p.cur().T == t }

func (p *P) expect(t lexer.Type) lexer.Tok {
	if !p.is(t) {
		p.err(fmt.Sprintf("expected %s, got %s", t.String(), p.cur().T.String()))
		return lexer.Tok{}
	}
	return p.next()
}

func (p *P) skipNL() {
	for p.is(lexer.NEWLINE) {
		p.next()
	}
}

func (p *P) done() bool { return p.pos >= len(p.toks) || p.cur().T == lexer.EOF }

func (p *P) err(msg string) { p.errs = append(p.errs, fmt.Sprintf("%d:%d: %s", p.cur().Line, p.cur().Col, msg)) }
