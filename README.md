# Byte Crusher 🗜️

A fast, interactive, terminal-based (TUI) Hex Editor specifically built for **glitch art** and direct binary data manipulation ("byte crushing"). 

Written in Go using the `tview` library, Byte Crusher allows you to safely corrupt, mutate, and manipulate file data in real-time right from your terminal without the fear of destroying the file's structure.

![Build Status](https://img.shields.io/badge/build-passing-brightgreen)
![Go Version](https://img.shields.io/badge/go-%3E%3D1.20-blue)

## Features

- **Interactive Hex Grid:** Navigate freely through your file using arrow keys. Both hex addresses and ASCII representations are rendered seamlessly.
- **Configurable Header Protection:** Lock modifications to the first N bytes of a file (configurable on the fly, 512 bytes by default). This guarantees that essential file headers (magic bytes, metadata) aren't destroyed, ensuring image viewers and media players can still successfully parse the glitched files!
- **Surgical Editing & Data Bending:** Target specific bytes or execute massive file-wide operations.
- **Dynamic Selection Mode:** Press `V` to drop an Anchor and highlight a specific chunk of memory. All advanced operations will automatically bind exclusively to your selected range!
- **Infinite Undo / Redo:** Nothing is permanent until you save. Roll back any operation, from a single byte edit to a total file corruption, using the built-in state stack.

## Glitch Arsenal (Operations)

The right panel features a tabbed interface with three primary toolsets:

### 1. Mutate
- **Corrupt:** Instantly randomize between 1% to 100% of the active data.
- **Byte Shift:** Push the integer values of bytes up or down (e.g. `val + 10`). Incredible for shifting color palettes in BMP files.
- **Chunk Sorting:** Execute an ascending or descending sort over raw binary chunks to achieve classic "Pixel Sorting / Melting" effects.
- **Reverse:** Mirror a selected sequence of bytes backwards.
- **Pattern Overwrite:** Forcefully duplicate a custom hex pattern across a massive region.

### 2. Math & Find
- **Bitwise Engine:** Run `XOR`, `AND`, or `OR` operations across massive datasets using a single byte key. This generates beautiful, mathematically repeating geometric noise.
- **Search & Replace:** Sweep the file instantly to replace specific hex sequences with another pattern.

### 3. System
- **History & Nav:** `Undo`, `Redo`, total revert, and direct Address Jumping.
- **Header Guard:** Instantly unlock or expand the amount of protected header bytes.
- **File Management:** Open files on the fly and save modifications directly to disk.

## Installation

Ensure you have [Go](https://go.dev/dl/) correctly installed.

Clone the repository and build the binary:

```bash
git clone https://github.com/yourusername/byte_crusher.git
cd byte_crusher
go build -o byte_crusher main.go
```

## Usage

You can launch Byte Crusher with or without a target file:

```bash
# Launch directly into a file
./byte_crusher ./assets/sample_image.bmp

# Launch empty and open via the UI
./byte_crusher
```

### Hotkeys & Navigation

- <kbd>↑</kbd> <kbd>↓</kbd> <kbd>←</kbd> <kbd>→</kbd> : Move the cursor across the byte cells.
- <kbd>Tab</kbd> : Cycle focus between the main Hex Table and the Operations panel.
- <kbd>V</kbd> / <kbd>M</kbd> : Set anchor and begin Selection Mode.
- <kbd>Esc</kbd> : Cancel/clear current active selection.
- <kbd>Enter</kbd> (in Hex Table) : Opens a quick-edit modal to override the hovered byte.
- <kbd>R</kbd> (in Hex Table) : Opens a full 16-byte replacement modal for the current row.
- <kbd>Backspace</kbd> / <kbd>Del</kbd> : Deletes the hovered row entirely, shrinking the file.

### A Note on Formats for Glitch Art

Not all formats are equal when it comes to intentional data corruption:

- **👑 BMP / JPG / WAV:** Highly recommended. These formats handle corruption gracefully, resulting in stunning visual glitches (macroblock shifting, color banding) or audio screeches without completely breaking.
- **❌ PNG / ZIP:** Not recommended. PNG uses `zlib` compression with rigid CRC layer checks. Modifying even a single pixel byte will break the internal checksums, causing modern viewers to refuse to load the image entirely rather than showing a glitch.

## Built With

- [Go](https://golang.org/)
- [tview](https://github.com/rivo/tview) - Rich Interactive widgets for terminal UIs
- [tcell](https://github.com/gdamore/tcell) - Terminal handling

## License

This project is open-source and available under the MIT License.
