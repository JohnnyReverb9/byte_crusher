# Byte Crusher 🗜️

A fast, interactive, terminal-based (TUI) Hex Editor specifically built for **glitch art** and direct binary data manipulation ("byte crushing"). 

Written in Go using the `tview` library, Byte Crusher allows you to safely corrupt, mutate, and manipulate file data in real-time right from your terminal.

![Build Status](https://img.shields.io/badge/build-passing-brightgreen)
![Go Version](https://img.shields.io/badge/go-%3E%3D1.20-blue)

## Features

- **Interactive Hex Grid:** Navigate freely through your file using arrow keys. Both hex addresses and ASCII representations are rendered seamlessly.
- **Header Protection (`SafeHeaderSize`):** Unlike basic hex editors, Byte Crusher locks you from modifying the first 512 bytes of a file (highlighted in grey). This guarantees that essential file headers (magic bytes, metadata) aren't destroyed, ensuring that image viewers and media players can still successfully parse and display the glitched files!
- **Surgical Editing:** 
  - <kbd>Enter</kbd>: Modify an exact byte on the spot.
  - <kbd>R</kbd>: Replace an entire 16-byte row.
  - <kbd>Del</kbd> / <kbd>Backspace</kbd>: Permanently delete a raw data sequence, shrinking the file.
- **Mass Corruption Algorithms:** Instantly randomize between 1% to 100% of the active data via the built-in Corruption severity menu.
- **Find and Replace:** Sweep the entire file instantly to replace specific hex sequences with another pattern.
- **Jump Navigation:** Effortlessly snap to a specific memory offset using the `Jump` dialog.

## Installation

Ensure you have [Go](https://go.dev/dl/) correctly installed.

Clone the repository and build the binary:

```bash
git clone https://github.com/yourusername/byte_crusher.git
cd byte_crusher
go build -o byte_crusher main.go
```

## Usage

You can launch Byte Crusher with or without a target file. 

```bash
# Launch directly into a file
./byte_crusher ./assets/sample_image.bmp

# Launch empty and open via the UI
./byte_crusher
```

### Hotkeys & Navigation

- <kbd>↑</kbd> <kbd>↓</kbd> <kbd>←</kbd> <kbd>→</kbd> : Move the cursor across the byte cells.
- <kbd>Tab</kbd> : Cycle focus between the main Hex Table and the Operations panel.
- <kbd>Enter</kbd> (in Hex Table) : Opens a quick-edit modal to override the hovered byte.
- <kbd>R</kbd> (in Hex Table) : Opens a full 16-byte replacement modal for the current row.
- <kbd>Backspace</kbd> / <kbd>Del</kbd> : Deletes the hovered row entirely.

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
