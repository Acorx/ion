package codegen

import (
	"fmt"
	"strings"

	"github.com/Acorx/ion/internal/parser"
)

type G struct{ pkg string }

func New(pkg string) *G { return &G{pkg: pkg} }

func (g *G) Gen(p *parser.Prog) map[string]string {
	files := map[string]string{}
	files["MainActivity.kt"] = g.mainActivity(p)
	for _, s := range p.Screens {
		files[s.Name+"Activity.kt"] = g.screenActivity(s)
		files["activity_"+snake(s.Name)+".xml"] = g.layout(s)
	}
	files["AndroidManifest.xml"] = g.manifest(p)
	files["build.gradle"] = g.gradle()
	return files
}

// B wraps strings.Builder with helper methods
type B struct{ strings.Builder }

func (b *B) f(f string, a ...interface{}) { b.Builder.WriteString(fmt.Sprintf(f, a...)) }
func (b *B) s(ss string)                  { b.Builder.WriteString(ss) }

func (g *G) mainActivity(p *parser.Prog) string {
	var b B
	b.f("package %s\n\n", g.pkg)
	b.s("import android.content.Intent\nimport android.os.Bundle\n")
	b.s("import android.widget.*\nimport androidx.appcompat.app.AppCompatActivity\n\n")
	b.s("class MainActivity : AppCompatActivity() {\n")
	b.s("    override fun onCreate(savedInstanceState: Bundle?) {\n")
	b.s("        super.onCreate(savedInstanceState)\n")
	if len(p.Screens) > 0 {
		sc := p.Screens[0]
		b.f("        setContentView(R.layout.activity_%s)\n", snake(sc.Name))
		b.s(g.setupCode(sc, "        "))
	}
	b.s("    }\n}\n")
	return b.String()
}

func (g *G) screenActivity(s parser.Screen) string {
	var b B
	b.f("package %s\n\n", g.pkg)
	b.s("import android.content.Intent\nimport android.os.Bundle\n")
	b.s("import android.widget.*\nimport androidx.appcompat.app.AppCompatActivity\n\n")
	b.f("class %sActivity : AppCompatActivity() {\n", s.Name)
	b.s("    override fun onCreate(savedInstanceState: Bundle?) {\n")
	b.s("        super.onCreate(savedInstanceState)\n")
	b.f("        setContentView(R.layout.activity_%s)\n\n", snake(s.Name))
	b.s(g.setupCode(s, "        "))
	b.s("    }\n}\n")
	return b.String()
}

// setupCode generates Kotlin code for component event handlers
func (g *G) setupCode(s parser.Screen, ind string) string {
	var b B
	id := 0
	for _, st := range s.Body {
		id++
		uid := fmt.Sprintf("ion_%d", id)
		if es, ok := st.(parser.ExprStmt); ok {
			if ce, ok := es.E.(parser.CallExpr); ok {
				b.s(g.setupComponent(ce, uid, ind))
				continue
			}
		}
		// Regular statement
		b.s(g.stmt(st, ind))
	}
	return b.String()
}

func (g *G) setupComponent(ce parser.CallExpr, uid, ind string) string {
	var b B
	switch ce.Fn {
	case "__text":
		if len(ce.Args) > 0 {
			b.f("%sfindViewById<TextView>(R.id.%s).text = %s\n", ind, uid, g.expr(ce.Args[0]))
		}

	case "__button":
		label := "\"Button\""
		if len(ce.Args) > 0 {
			label = g.expr(ce.Args[0])
		}
		// If there's a handler arg (index 1), wire it up
		if len(ce.Args) > 1 {
			if blk, ok := ce.Args[1].(*parser.BlockExpr); ok {
				b.f("%sfindViewById<Button>(R.id.%s).setOnClickListener {\n", ind, uid)
				for _, s := range blk.Body {
					b.s(g.stmt(s, ind+"    "))
				}
				b.f("%s}\n", ind)
			}
		} else {
			b.f("%sfindViewById<Button>(R.id.%s).setOnClickListener {\n", ind, uid)
			b.f("%s    Toast.makeText(this, %s, Toast.LENGTH_SHORT).show()\n", ind, label)
			b.f("%s}\n", ind)
		}

	case "__input":
		// nothing to wire by default
	case "__switch":
		if len(ce.Args) > 1 {
			if blk, ok := ce.Args[1].(*parser.BlockExpr); ok {
				b.f("%sfindViewById<Switch>(R.id.%s).setOnCheckedChangeListener { _, isChecked ->\n", ind, uid)
				for _, s := range blk.Body {
					b.s(g.stmt(s, ind+"    "))
				}
				b.f("%s}\n", ind)
			}
		}
	}
	return b.String()
}

func (g *G) stmt(s parser.Stmt, ind string) string {
	var b B
	switch st := s.(type) {
	case parser.AssignStmt:
		b.f("%sval %s = %s\n", ind, st.Name, g.expr(st.Val))
	case parser.NavStmt:
		b.f("%sstartActivity(Intent(this, %sActivity::class.java))\n", ind, st.Target)
	case parser.BackStmt:
		b.f("%sfinish()\n", ind)
	case parser.ToastStmt:
		b.f("%sToast.makeText(this, %s, Toast.LENGTH_SHORT).show()\n", ind, g.expr(st.Msg))
	case parser.VibStmt:
		b.f("%svibrate(200)\n", ind)
	case parser.NotifStmt:
		b.f("%s// notify: %s, %s\n", ind, g.expr(st.Title), g.expr(st.Msg))
	case parser.IfStmt:
		b.f("%sif (%s) {\n", ind, g.expr(st.Cond))
		for _, s2 := range st.Then {
			b.s(g.stmt(s2, ind+"    "))
		}
		if len(st.Else) > 0 {
			b.f("%s} else {\n", ind)
			for _, s2 := range st.Else {
				b.s(g.stmt(s2, ind+"    "))
			}
		}
		b.f("%s}\n", ind)
	case parser.ForStmt:
		b.f("%sfor (%s in %s) {\n", ind, st.Var, g.expr(st.Iter))
		for _, s2 := range st.Body {
			b.s(g.stmt(s2, ind+"    "))
		}
		b.f("%s}\n", ind)
	case parser.WhileStmt:
		b.f("%swhile (%s) {\n", ind, g.expr(st.Cond))
		for _, s2 := range st.Body {
			b.s(g.stmt(s2, ind+"    "))
		}
		b.f("%s}\n", ind)
	case parser.RetStmt:
		b.f("%sreturn %s\n", ind, g.expr(st.Val))
	case parser.AwaitStmt:
		b.f("%s// await %s\n", ind, g.expr(st.Call))
	case parser.BgStmt:
		b.f("%sThread {\n", ind)
		for _, s2 := range st.Body {
			b.s(g.stmt(s2, ind+"    "))
		}
		b.f("%s}.start()\n", ind)
	case parser.NativeStmt:
		b.f("%s%s\n", ind, st.Code)
	case parser.ExprStmt:
		b.f("%s%s\n", ind, g.expr(st.E))
	}
	return b.String()
}

func (g *G) expr(e parser.Expr) string {
	if e == nil {
		return "null"
	}
	switch ex := e.(type) {
	case parser.StrExpr:
		return fmt.Sprintf("\"%s\"", ex.V)
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
		return fmt.Sprintf("(%s %s %s)", g.expr(ex.L), ex.Op, g.expr(ex.R))
	case parser.UnExpr:
		return fmt.Sprintf("%s%s", ex.Op, g.expr(ex.R))
	case parser.CallExpr:
		args := make([]string, len(ex.Args))
		for i, a := range ex.Args {
			args[i] = g.expr(a)
		}
		return fmt.Sprintf("%s(%s)", ex.Fn, strings.Join(args, ", "))
	case parser.MethodExpr:
		args := make([]string, len(ex.Args))
		for i, a := range ex.Args {
			args[i] = g.expr(a)
		}
		return fmt.Sprintf("%s.%s(%s)", g.expr(ex.Obj), ex.Method, strings.Join(args, ", "))
	case parser.FieldExpr:
		return fmt.Sprintf("%s.%s", g.expr(ex.Obj), ex.Field)
	case parser.IdxExpr:
		return fmt.Sprintf("%s[%s]", g.expr(ex.Obj), g.expr(ex.Idx))
	case parser.ArrExpr:
		elems := make([]string, len(ex.Elems))
		for i, e2 := range ex.Elems {
			elems[i] = g.expr(e2)
		}
		return fmt.Sprintf("listOf(%s)", strings.Join(elems, ", "))
	case parser.BlockExpr:
		// Block as expression — generate lambda
		var b B
		b.s("{\n")
		for _, s := range ex.Body {
			b.s(g.stmt(s, "    "))
		}
		b.s("}")
		return b.String()
	default:
		return "null"
	}
}

func (g *G) layout(s parser.Screen) string {
	var b B
	b.s("<?xml version=\"1.0\" encoding=\"utf-8\"?>\n")
	b.s("<LinearLayout xmlns:android=\"http://schemas.android.com/apk/res/android\"\n")
	b.s("    android:layout_width=\"match_parent\"\n    android:layout_height=\"match_parent\"\n")
	b.s("    android:orientation=\"vertical\" android:padding=\"16dp\">\n")
	id := 0
	for _, st := range s.Body {
		id++
		if es, ok := st.(parser.ExprStmt); ok {
			g.layoutNode(&b, es.E, id)
		}
	}
	b.s("</LinearLayout>\n")
	return b.String()
}

func (g *G) layoutNode(b *B, e parser.Expr, id int) {
	uid := fmt.Sprintf("ion_%d", id)
	if ce, ok := e.(parser.CallExpr); ok {
		switch ce.Fn {
		case "__text":
			b.f("    <TextView android:id=\"@+id/%s\"\n", uid)
			b.s("        android:layout_width=\"wrap_content\"\n        android:layout_height=\"wrap_content\"\n")
			b.s("        android:textSize=\"18sp\" />\n\n")
		case "__button":
			b.f("    <Button android:id=\"@+id/%s\"\n", uid)
			b.s("        android:layout_width=\"wrap_content\"\n        android:layout_height=\"wrap_content\" />\n\n")
		case "__input":
			b.f("    <EditText android:id=\"@+id/%s\"\n", uid)
			b.s("        android:layout_width=\"match_parent\"\n        android:layout_height=\"wrap_content\" />\n\n")
		case "__switch":
			b.f("    <Switch android:id=\"@+id/%s\"\n", uid)
			b.s("        android:layout_width=\"wrap_content\"\n        android:layout_height=\"wrap_content\" />\n\n")
		case "__image":
			b.f("    <ImageView android:id=\"@+id/%s\"\n", uid)
			b.s("        android:layout_width=\"wrap_content\"\n        android:layout_height=\"wrap_content\" />\n\n")
		case "__progress":
			b.f("    <ProgressBar android:id=\"@+id/%s\"\n", uid)
			b.s("        android:layout_width=\"match_parent\"\n        android:layout_height=\"wrap_content\"\n")
			b.s("        style=\"?android:attr/progressBarStyleHorizontal\" android:max=\"100\" />\n\n")
		}
	}
}

func (g *G) manifest(p *parser.Prog) string {
	var b B
	b.s("<?xml version=\"1.0\" encoding=\"utf-8\"?>\n")
	b.f("<manifest xmlns:android=\"http://schemas.android.com/apk/res/android\" package=\"%s\">\n", g.pkg)
	b.f("    <application android:label=\"%s\"\n", p.Name)
	b.s("        android:theme=\"@style/Theme.AppCompat.Light\">\n")
	b.s("        <activity android:name=\".MainActivity\" android:exported=\"true\">\n")
	b.s("            <intent-filter>\n")
	b.s("                <action android:name=\"android.intent.action.MAIN\" />\n")
	b.s("                <category android:name=\"android.intent.category.LAUNCHER\" />\n")
	b.s("            </intent-filter>\n        </activity>\n")
	for _, s := range p.Screens[1:] {
		b.f("        <activity android:name=\".%sActivity\" />\n", s.Name)
	}
	b.s("    </application>\n</manifest>\n")
	return b.String()
}

func (g *G) gradle() string {
	return fmt.Sprintf(`plugins {
    id 'com.android.application'
    id 'org.jetbrains.kotlin.android'
}
android {
    namespace '%s'
    compileSdk 34
    defaultConfig {
        applicationId "%s"
        minSdk 24
        targetSdk 34
        versionCode 1
        versionName "1.0"
    }
}
dependencies {
    implementation 'androidx.appcompat:appcompat:1.6.1'
    implementation 'androidx.core:core-ktx:1.12.0'
}
`, g.pkg, g.pkg)
}

func snake(s string) string {
	var r []byte
	for i, c := range s {
		if c >= 'A' && c <= 'Z' {
			if i > 0 {
				r = append(r, '_')
			}
			r = append(r, byte(c+32))
		} else {
			r = append(r, byte(c))
		}
	}
	return string(r)
}
