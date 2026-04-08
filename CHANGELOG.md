# Changelog

All notable changes to this project will be documented in this file.

## [0.3.0] - 2025-04-08

### Added
- **Input binding** — `input "hint" -> varName` creates a property getter for the input value
  - Generates `private lateinit var ion_NInput: EditText`
  - Generates `private val varName: String get() = ion_NInput.text.toString()`
  - Access input values directly in functions: `toast("Hello " + userName)`
- Improved codegen architecture for bound variables

### Fixed
- Correct argument indexing for input binding (Args[1] instead of Args[2])
- Removed unused `collectBindVars` function

## [0.2.0] - 2025-04-08

### Added
- Complete CLI with colored output (`fatih/color`)
  - `ion build <file.ion> [--out dir]`
  - `ion transpile <file.ion>`
  - `ion format <file.ion>`
  - `ion check <file.ion>`
  - `ion watch <file.ion>`
- Unit tests for lexer, parser, and codegen
- Comprehensive README with badges and architecture diagram
- Precise error messages with line/column numbers
- Formatter for pretty-printing .ion files

## [0.1.0] - 2025-04-07

### Added
- Initial release
- Lexer with UTF-8 arrow support (`→`)
- Parser for app/screen/fn/if/http constructs
- Code generator for MainActivity, layouts, and AndroidManifest
- Basic UI components: text, button, input, switch, progress