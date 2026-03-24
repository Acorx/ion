# ion ⚡

> Python readability. Kotlin performance. Android native.

A minimalist programming language that transpiles to Android Kotlin. Write less, ship more.

## Example

```ion
app HelloApp

screen Main {
    text "Welcome to ion! ⚡"
    button "Say Hello" -> toast("Hello!")
    button "Settings" -> navigate(Settings)
}

screen Settings {
    text "Settings Page"
    switch "Dark Mode" -> toggle_theme()
    button "Back" -> back()
}
```

Compiles to native Android Kotlin + XML layouts. Zero runtime. Zero dependencies.

## Install

```bash
go install github.com/Acorx/ion/cmd@latest
```

## Usage

```bash
ion app.ion                    # Build → output/
ion app.ion --out ./myapp      # Build to custom directory
ion transpile app.ion          # Preview generated Kotlin
```

## Features

- **6x less code** than Kotlin for the same Android app
- **Brace-delimited** syntax (no indentation hell)
- **Inline event handlers** — `button "X" -> toast("Hi")`
- **Native Android output** — Kotlin + XML + Manifest + Gradle
- **Single binary** — no dependencies, no framework

## Syntax

```ion
app Name                       // App declaration

screen Name {                  // Screen (Activity)
    text "value"               // TextView
    button "label" -> action   // Button with click handler
    input "hint"               // EditText
    switch "label" -> action   // Switch
    image "src"                // ImageView
    progress                   // ProgressBar
}

fn name(params) { ... }        // Function
native -> "kotlin code"        // Escape hatch

// Statements
x = 42                         // Assignment
navigate(Screen)               // Navigate to screen
back()                         // Go back
toast("msg")                   // Show toast
vibrate()                      // Vibrate device
notify("title", "msg")         // Notification
if expr { ... } else { ... }
for x in list { ... }
while expr { ... }
return expr
await expr
background { ... }
```

## Architecture

```
.ion file
  → Lexer (tokens)
  → Parser (AST)
  → Code Generator (Kotlin + XML)
  → Android project
```

**Total: ~1200 lines of Go.**

## Philosophy

> "La contrainte est la liberté."

- 1 token = 1 meaning
- 1 node = 1 line of output
- Zero ambiguity
- Parser fits in your head

## License

MIT
