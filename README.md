# ion 🦊

[![Go Report Card](https://goreportcard.com/badge/github.com/Acorx/ion)](https://goreportcard.com/report/github.com/Acorx/ion)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Version](https://img.shields.io/badge/version-0.3.0-orange.svg)]()

**A minimalist DSL for Android app development.** Write UI and logic in a clean syntax, generate production-ready Kotlin.

## Features

- 🎨 **Declarative UI** — text, button, input, list, switch, and more
- 🔄 **Reactive state** — `state` declarations auto-update UI
- 🌐 **HTTP built-in** — `http get/post` with async handling
- 📱 **Navigation** — `navigate(Screen)` and `back()`
- 🔔 **Notifications** — toast, vibrate, system notify
- 📤 **Sharing** — `share()` and `open()` for deep links
- ⚡ **Native escape hatch** — `native -> "Kotlin code"` for platform APIs
- 🎯 **Type inference** — no boilerplate, clean syntax

## Example

```ion
app TodoApp

screen Main {
  text "📝 My Tasks"
  input "Add a task..." -> add_task()
  button "Load from API" -> load_tasks()
}

fn add_task() {
  toast("Task added!")
}

fn load_tasks() {
 http get "https://api.example.com/tasks" -> data
 toast("Loaded!")
}
```

### Input Binding (v0.3.0)

Bind inputs to variables and use them directly:

```ion
screen Form {
 input "Your name" -> userName
 input "Your email" -> userEmail
 button "Submit" -> submit()
}

fn submit() {
 toast("Thanks " + userName + "!")
 share(userEmail)
}
```

Generated Kotlin:
```kotlin
class MainActivity : AppCompatActivity() {
 private lateinit var ion_1Input: EditText
 private lateinit var ion_2Input: EditText
 private val userName: String get() = ion_1Input.text.toString()
 private val userEmail: String get() = ion_2Input.text.toString()
 
 // userName and userEmail are accessible in all functions
}
```

## Generated Kotlin

```kotlin
// MainActivity.kt
package com.example.todoapp

import android.os.Bundle
import android.widget.*
import androidx.appcompat.app.AppCompatActivity

class MainActivity : AppCompatActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_main)
        
        // Setup UI components
        val ion_1 = findViewById<EditText>(R.id.ion_1)
        val ion_2 = findViewById<Button>(R.id.ion_2)
        
        ion_2.setOnClickListener { load_tasks() }
    }
    
    private fun load_tasks() {
        Toast.makeText(this, "Loaded!", Toast.LENGTH_SHORT).show()
    }
}
```

## Architecture

```
ion/
├── cmd/           # CLI entry point
├── internal/
│   ├── lexer/     # Tokenizer (line/col tracking)
│   ├── parser/    # AST builder
│   ├── codegen/   # Kotlin/XML generator
│   └── formatter/ # Source formatter
└── examples/      # Sample .ion files
```

**Compilation pipeline:**
```
.ion → lexer (tokens) → parser (AST) → codegen (Kotlin/XML)
```

## Installation

```bash
git clone https://github.com/Acorx/ion.git
cd ion
go build -o ion ./cmd
```

## Usage

```bash
# Build an .ion file
ion build app.ion --out ./android-project

# Show generated Kotlin
ion transpile app.ion

# Format source
ion format app.ion

# Validate without building
ion check app.ion

# Watch mode (rebuild on change)
ion watch app.ion --out ./output
```

## Roadmap

- [ ] Type system (int, string, list, map)
- [ ] Error recovery with suggestions
- [ ] LSP support (VS Code extension)
- [ ] Hot reload for development
- [ ] iOS target (Swift generation)
- [ ] Component libraries (Material, Compose)

## Philosophy

ion is **not** trying to replace Flutter/React Native. It's for:
- Simple CRUD apps
- Prototyping Android ideas
- Teaching UI programming
- Backend devs who need a mobile front-end

## License

MIT © AcorX Lab

## Contributing

PRs welcome! Areas of interest:
- More UI components (slider, checkbox, progress)
- Better error messages
- Code formatting improvements
- Test coverage
