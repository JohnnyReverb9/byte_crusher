package ui

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"byte_crusher/crusher"
)

type AppState struct {
	filePath        string
	originalData    []byte
	currentData     []byte
	undoStack       [][]byte
	redoStack       [][]byte
	selectionAnchor int
	selectionActive bool
	safeHeaderSize  int
	hexTable        *tview.Table
	app             *tview.Application
	pages           *tview.Pages
	rightPages      *tview.Pages // for tabs
	statusLine      *tview.TextView
}

func (s *AppState) loadFile(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	s.filePath = path
	s.originalData = append([]byte{}, data...)
	s.currentData = append([]byte{}, data...)
	s.undoStack = nil
	s.redoStack = nil
	s.selectionActive = false
	return nil
}

func (s *AppState) pushUndo() {
	if len(s.currentData) > 0 {
		buf := make([]byte, len(s.currentData))
		copy(buf, s.currentData)
		s.undoStack = append(s.undoStack, buf)
		s.redoStack = nil
	}
}

func (s *AppState) undo() {
	if len(s.undoStack) == 0 {
		s.setStatus("[yellow]Nothing to undo!")
		return
	}
	buf := make([]byte, len(s.currentData))
	copy(buf, s.currentData)
	s.redoStack = append(s.redoStack, buf)

	last := s.undoStack[len(s.undoStack)-1]
	s.undoStack = s.undoStack[:len(s.undoStack)-1]
	s.currentData = last
	s.updateHexTable()
	s.setStatus("[green]Undo successful")
}

func (s *AppState) redo() {
	if len(s.redoStack) == 0 {
		s.setStatus("[yellow]Nothing to redo!")
		return
	}
	buf := make([]byte, len(s.currentData))
	copy(buf, s.currentData)
	s.undoStack = append(s.undoStack, buf)

	next := s.redoStack[len(s.redoStack)-1]
	s.redoStack = s.redoStack[:len(s.redoStack)-1]
	s.currentData = next
	s.updateHexTable()
	s.setStatus("[green]Redo successful")
}

func (s *AppState) getActiveRange() (int, int) {
	if !s.selectionActive || s.selectionAnchor == -1 {
		return s.safeHeaderSize, len(s.currentData)
	}

	row, col := s.hexTable.GetSelection()
	if col < 1 || col > 16 {
		return s.safeHeaderSize, len(s.currentData)
	}

	idx := row*16 + (col - 1)
	min := s.selectionAnchor
	max := idx
	if min > max {
		min, max = max, min
	}

	if min < s.safeHeaderSize {
		min = s.safeHeaderSize
	}
	if max >= len(s.currentData) {
		max = len(s.currentData) - 1
	}

	return min, max + 1
}

func RunTUI(initialFilePath string) error {
	state := &AppState{
		app:            tview.NewApplication(),
		pages:          tview.NewPages(),
		rightPages:     tview.NewPages(),
		safeHeaderSize: 512,
		selectionAnchor: -1,
	}

	if initialFilePath != "" {
		err := state.loadFile(initialFilePath)
		if err != nil {
			return err
		}
	}

	state.hexTable = tview.NewTable().
		SetBorders(false).
		SetSelectable(true, true).
		SetFixed(0, 0)

	title := " No File Loaded (Press Open) "
	if state.filePath != "" {
		title = fmt.Sprintf(" Hex Table - %s ", state.filePath)
	}
	state.hexTable.SetBorder(true).SetTitle(title)

	state.statusLine = tview.NewTextView().SetDynamicColors(true)
	state.statusLine.SetBorder(true)
	state.setStatus("App loaded successfully.")

	// Tab switcher Bar
	tabBar := tview.NewForm()
	tabBar.SetHorizontal(true)
	tabBar.AddButton("Mutate", func() { state.rightPages.SwitchToPage("mutate") })
	tabBar.AddButton("Math/Find", func() { state.rightPages.SwitchToPage("math") })
	tabBar.AddButton("System", func() { state.rightPages.SwitchToPage("system") })

	// --- MUTATE TAB ---
	mutateFlex := tview.NewFlex().SetDirection(tview.FlexRow)
	
	corruptForm := tview.NewForm()
	corruptForm.AddInputField("Severity (0.01-1.0)", "0.05", 10, nil, nil)
	corruptForm.AddButton("Corrupt", func() {
		if state.filePath == "" { return }
		start, end := state.getActiveRange()
		sevStr := corruptForm.GetFormItemByLabel("Severity (0.01-1.0)").(*tview.InputField).GetText()
		sev, _ := strconv.ParseFloat(sevStr, 64)
		state.pushUndo()
		crusher.Corrupt(state.currentData, start, end, sev)
		state.updateHexTable()
		state.setStatus(fmt.Sprintf("[green]Corrupted %.2f%% in range %08x-%08x", sev*100, start, end))
	})
	corruptForm.SetTitle(" Byte Corruption ").SetBorder(true)
	mutateFlex.AddItem(corruptForm, 7, 1, false)

	shiftForm := tview.NewForm()
	shiftForm.AddInputField("Shift (+/-)", "10", 10, nil, nil)
	shiftForm.AddButton("Shift", func() {
		if state.filePath == "" { return }
		start, end := state.getActiveRange()
		shStr := shiftForm.GetFormItemByLabel("Shift (+/-)").(*tview.InputField).GetText()
		sh, _ := strconv.ParseInt(shStr, 10, 8)
		state.pushUndo()
		crusher.ByteShift(state.currentData, start, end, int8(sh))
		state.updateHexTable()
		state.setStatus(fmt.Sprintf("[green]Shifted range %08x-%08x by %d", start, end, sh))
	})
	shiftForm.SetTitle(" Byte Shifting ").SetBorder(true)
	mutateFlex.AddItem(shiftForm, 7, 1, false)

	chunkForm := tview.NewForm()
	chunkForm.AddButton("Sort Asc", func() {
		if state.filePath == "" { return }
		start, end := state.getActiveRange()
		state.pushUndo()
		crusher.SortChunk(state.currentData, start, end, true)
		state.updateHexTable()
		state.setStatus("[green]Sorted chunk Ascending")
	})
	chunkForm.AddButton("Sort Desc", func() {
		if state.filePath == "" { return }
		start, end := state.getActiveRange()
		state.pushUndo()
		crusher.SortChunk(state.currentData, start, end, false)
		state.updateHexTable()
		state.setStatus("[green]Sorted chunk Descending")
	})
	chunkForm.AddButton("Reverse", func() {
		if state.filePath == "" { return }
		start, end := state.getActiveRange()
		state.pushUndo()
		crusher.ReverseChunk(state.currentData, start, end)
		state.updateHexTable()
		state.setStatus("[green]Reversed chunk")
	})
	chunkForm.SetTitle(" Block Operations ").SetBorder(true)
	mutateFlex.AddItem(chunkForm, 5, 1, false)

	pattForm := tview.NewForm()
	pattForm.AddInputField("Pattern (Hex)", "ff00", 10, nil, nil)
	pattForm.AddButton("Overwrite", func() {
		if state.filePath == "" { return }
		start, end := state.getActiveRange()
		patStr := pattForm.GetFormItemByLabel("Pattern (Hex)").(*tview.InputField).GetText()
		patStr = strings.ReplaceAll(patStr, " ", "")
		b, err := hex.DecodeString(patStr)
		if err == nil && len(b) > 0 {
			state.pushUndo()
			crusher.PatternOverwrite(state.currentData, start, end, b)
			state.updateHexTable()
			state.setStatus("[green]Overwrote with pattern")
		} else {
			state.setStatus("[red]Invalid pattern hex")
		}
	})
	pattForm.SetTitle(" Pattern Overwrite ").SetBorder(true)
	mutateFlex.AddItem(pattForm, 7, 1, false)

	state.rightPages.AddPage("mutate", mutateFlex, true, true)

	// --- MATH/FIND TAB ---
	mathFlex := tview.NewFlex().SetDirection(tview.FlexRow)
	
	bitForm := tview.NewForm()
	bitForm.AddInputField("Key (Hex, 1 byte)", "5a", 10, nil, nil)
	bitForm.AddButton("XOR", func() { state.applyBitwise("xor", bitForm) })
	bitForm.AddButton("AND", func() { state.applyBitwise("and", bitForm) })
	bitForm.AddButton("OR", func() { state.applyBitwise("or", bitForm) })
	bitForm.SetTitle(" Bitwise Operations ").SetBorder(true)
	mathFlex.AddItem(bitForm, 7, 1, false)

	replaceForm := tview.NewForm()
	replaceForm.AddInputField("Find (Hex)", "", 16, nil, nil)
	replaceForm.AddInputField("Replace (Hex)", "", 16, nil, nil)
	replaceForm.AddButton("Replace", func() {
		if state.filePath == "" { return }
		start, end := state.getActiveRange()
		fromStr := replaceForm.GetFormItemByLabel("Find (Hex)").(*tview.InputField).GetText()
		toStr := replaceForm.GetFormItemByLabel("Replace (Hex)").(*tview.InputField).GetText()
		fromStr = strings.ReplaceAll(fromStr, " ", "")
		toStr = strings.ReplaceAll(toStr, " ", "")
		from, err1 := hex.DecodeString(fromStr)
		to, err2 := hex.DecodeString(toStr)

		if err1 != nil || err2 != nil || len(from) == 0 {
			state.setStatus("[red]Invalid hex string in Find/Replace")
			return
		}
		state.pushUndo()
		state.currentData = crusher.ReplacePattern(state.currentData, start, end, from, to)
		state.updateHexTable()
		state.setStatus(fmt.Sprintf("[green]Replaced %s with %s", fromStr, toStr))
	})
	replaceForm.SetTitle(" Search & Replace ").SetBorder(true)
	mathFlex.AddItem(replaceForm, 9, 1, false)
	
	state.rightPages.AddPage("math", mathFlex, true, false)

	// --- SYSTEM TAB ---
	sysFlex := tview.NewFlex().SetDirection(tview.FlexRow)
	
	histForm := tview.NewForm()
	histForm.AddButton("Undo", func() { state.undo() })
	histForm.AddButton("Redo", func() { state.redo() })
	histForm.AddButton("Reset Fl", func() {
		if state.filePath == "" { return }
		state.pushUndo()
		state.currentData = append([]byte{}, state.originalData...)
		state.updateHexTable()
		state.hexTable.ScrollToBeginning()
		state.hexTable.Select(0, 1)
		state.setStatus("[yellow]Reset to original file")
	})
	histForm.AddButton("Jump", func() {
		if state.filePath == "" { return }
		state.showInputModal("Jump to Offset", "Target Hex (e.g. 1a0):", "", func(newVal string) {
			newVal = strings.TrimSpace(newVal)
			newVal = strings.TrimPrefix(newVal, "0x")
			addr, err := strconv.ParseInt(newVal, 16, 64)
			if err != nil { return }
			row := int(addr / 16)
			if row < 0 { row = 0 }
			if row >= state.hexTable.GetRowCount() { row = state.hexTable.GetRowCount() - 1 }
			state.hexTable.Select(row, 1)
		})
	})
	histForm.SetTitle(" History & Nav ").SetBorder(true)
	sysFlex.AddItem(histForm, 9, 1, false)

	hdrForm := tview.NewForm()
	hdrForm.AddInputField("Header Size (B)", fmt.Sprintf("%d", state.safeHeaderSize), 10, nil, nil)
	hdrForm.AddButton("Set", func() {
		val := hdrForm.GetFormItemByLabel("Header Size (B)").(*tview.InputField).GetText()
		h, err := strconv.Atoi(val)
		if err == nil && h >= 0 {
			state.safeHeaderSize = h
			state.updateHexTable()
			state.setStatus(fmt.Sprintf("[green]Header size locked to %d bytes", h))
		}
	})
	hdrForm.SetTitle(" Header Guard ").SetBorder(true)
	sysFlex.AddItem(hdrForm, 7, 1, false)

	fileForm := tview.NewForm()
	fileForm.AddButton("Open File", func() {
		state.showInputModal("Open New File", "File Path:", state.filePath, func(newVal string) {
			newVal = strings.TrimSpace(newVal)
			if newVal == "" { return }
			err := state.loadFile(newVal)
			if err != nil {
				state.setStatus(fmt.Sprintf("[red]Load failed: %v", err))
			} else {
				state.hexTable.SetTitle(fmt.Sprintf(" Hex Table - %s ", state.filePath))
				state.updateHexTable()
				state.hexTable.ScrollToBeginning()
				state.hexTable.Select(0, 1)
				state.setStatus(fmt.Sprintf("[green]Successfully loaded %s", state.filePath))
				hdrForm.GetFormItemByLabel("Header Size (B)").(*tview.InputField).SetText("512")
				state.safeHeaderSize = 512
			}
		})
	})
	fileForm.AddButton("Save File", func() {
		if state.filePath == "" { return }
		err := ioutil.WriteFile(state.filePath, state.currentData, 0644)
		if err != nil {
			state.setStatus(fmt.Sprintf("[red]Save failed: %v", err))
		} else {
			state.setStatus("[green]File saved successfully!")
			state.originalData = append([]byte{}, state.currentData...)
		}
	})
	fileForm.AddButton("Quit", func() { state.app.Stop() })
	fileForm.SetTitle(" File Actions ").SetBorder(true)
	sysFlex.AddItem(fileForm, 0, 1, false)

	state.rightPages.AddPage("system", sysFlex, true, false)

	// Combine components
	rightPanel := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(tabBar, 3, 1, false).
		AddItem(state.rightPages, 0, 1, false)

	mainFlex := tview.NewFlex().
		AddItem(state.hexTable, 0, 3, true).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(rightPanel, 0, 8, false).
			AddItem(state.statusLine, 3, 1, false), 0, 1, false)

	state.pages.AddPage("main", mainFlex, true, true)
	state.app.SetRoot(state.pages, true).EnableMouse(true)

	// App-level shortcuts
	state.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlQ {
			state.app.Stop()
			return nil
		}
		if event.Key() == tcell.KeyTab {
			if state.hexTable.HasFocus() {
				state.app.SetFocus(tabBar)
			} else {
				state.app.SetFocus(state.hexTable)
			}
			return nil
		}
		return event
	})

	// Table-specific interactive shortcuts
	state.hexTable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if state.filePath == "" { return event }

		if event.Rune() == 'v' || event.Rune() == 'V' || event.Rune() == 'm' || event.Rune() == 'M' {
			if !state.selectionActive {
				row, col := state.hexTable.GetSelection()
				if col >= 1 && col <= 16 {
					state.selectionActive = true
					state.selectionAnchor = row*16 + (col - 1)
					state.setStatus("[blue]Selection mode started! Use arrow keys. Hit Esc to cancel.")
					state.updateHexTable() // redraw colors
				}
			} else {
				state.selectionActive = false
				state.setStatus("[white]Selection cleared.")
				state.updateHexTable()
			}
			return nil
		}

		if event.Key() == tcell.KeyEsc {
			if state.selectionActive {
				state.selectionActive = false
				state.setStatus("[white]Selection cleared.")
				state.updateHexTable()
				return nil
			}
		}

		row, col := state.hexTable.GetSelection()

		// Allow redrawing selection visually when moving
		if state.selectionActive && (event.Key() == tcell.KeyUp || event.Key() == tcell.KeyDown || event.Key() == tcell.KeyLeft || event.Key() == tcell.KeyRight) {
			// Actually we need to let tview change selection, then redraw.
			// But input capture intercepts before. We can pass it through and redraw on 'SelectionChanged' callback instead.
			// However doing it manually or letting it run:
			// Just returning event is fine. But we need to schedule a redraw if selection changes.
			// Let's rely on Table's SelectionChanged event below!
		}

		if col < 1 || col > 16 {
			if event.Key() == tcell.KeyLeft && col > 1 {
				state.hexTable.Select(row, 16)
				return nil
			} else if event.Key() == tcell.KeyRight && col < 17 {
				state.hexTable.Select(row, 1)
				return nil
			}
			return event
		}

		byteIndex := row*16 + (col - 1)
		isProtected := byteIndex < state.safeHeaderSize
		wantsModify := event.Key() == tcell.KeyEnter || event.Key() == tcell.KeyBackspace || event.Key() == tcell.KeyDelete || event.Key() == tcell.KeyBackspace2 || event.Rune() == 'R' || event.Rune() == 'r'

		if isProtected && wantsModify {
			state.setStatus("[red]Cannot modify header bytes manually! Security lock is active.")
			return nil
		}

		if byteIndex >= len(state.currentData) { return event }

		if event.Key() == tcell.KeyEnter {
			currentByteStr := fmt.Sprintf("%02x", state.currentData[byteIndex])
			state.showInputModal(fmt.Sprintf("Edit Byte at %08x", byteIndex), "New Hex Value:", currentByteStr, func(newVal string) {
				newVal = strings.TrimSpace(newVal)
				b, err := hex.DecodeString(newVal)
				if err != nil || len(b) != 1 {
					state.setStatus("[red]Invalid hex format. Must be 1 byte (e.g. FF)")
					return
				}
				state.pushUndo()
				state.currentData[byteIndex] = b[0]
				state.updateHexTable()
				state.hexTable.Select(row, col)
				state.setStatus(fmt.Sprintf("[green]Updated byte to %02x", b[0]))
			})
			return nil
		}

		if event.Key() == tcell.KeyBackspace || event.Key() == tcell.KeyDelete || event.Key() == tcell.KeyBackspace2 {
			startIdx := row * 16
			if startIdx < state.safeHeaderSize {
				state.setStatus("[red]Cannot delete row in protected header territory!")
				return nil
			}
			endIdx := startIdx + 16
			if endIdx > len(state.currentData) { endIdx = len(state.currentData) }
			
			state.pushUndo()
			state.currentData = append(state.currentData[:startIdx], state.currentData[endIdx:]...)
			state.selectionActive = false // reset sel if deleted
			state.updateHexTable()
			
			if row >= state.hexTable.GetRowCount() { row = state.hexTable.GetRowCount() - 1 }
			if row < 0 { row = 0 }
			state.hexTable.Select(row, col)
			state.setStatus(fmt.Sprintf("[yellow]Deleted row at address %08x", startIdx))
			return nil
		}

		if event.Rune() == 'R' || event.Rune() == 'r' {
			startIdx := row * 16
			endIdx := startIdx + 16
			if endIdx > len(state.currentData) { endIdx = len(state.currentData) }
			existingHex := hex.EncodeToString(state.currentData[startIdx:endIdx])
			
			state.showInputModal(fmt.Sprintf("Replace %d bytes starting %08x", endIdx-startIdx, startIdx), "New Sequence (Hex):", existingHex, func(newVal string) {
				newVal = strings.ReplaceAll(newVal, " ", "")
				b, err := hex.DecodeString(newVal)
				if err != nil {
					state.setStatus("[red]Invalid hex string provided.")
					return
				}
				state.pushUndo()
				var newData []byte
				newData = append(newData, state.currentData[:startIdx]...)
				newData = append(newData, b...)
				newData = append(newData, state.currentData[endIdx:]...)
				state.currentData = newData
				state.updateHexTable()
				state.hexTable.Select(row, col)
				state.setStatus(fmt.Sprintf("[green]Replaced row data successfully!"))
			})
			return nil
		}

		return event
	})

	state.hexTable.SetSelectionChangedFunc(func(row, column int) {
		if state.selectionActive {
			state.updateHexTableColorsOnly()
		}
	})

	state.updateHexTable()

	if state.filePath == "" {
		state.showInputModal("Welcome to Byte Crusher", "Enter File Path to Open:", "", func(newVal string) {
			newVal = strings.TrimSpace(newVal)
			if newVal != "" {
				err := state.loadFile(newVal)
				if err != nil {
					state.setStatus(fmt.Sprintf("[red]Load failed: %v", err))
				} else {
					state.hexTable.SetTitle(fmt.Sprintf(" Hex Table - %s ", state.filePath))
					state.updateHexTable()
					state.hexTable.ScrollToBeginning()
					state.hexTable.Select(0, 1)
					state.setStatus(fmt.Sprintf("[green]Loaded %s", state.filePath))
				}
			}
		})
	}

	return state.app.Run()
}

func (s *AppState) applyBitwise(op string, form *tview.Form) {
	if s.filePath == "" { return }
	keyStr := form.GetFormItemByLabel("Key (Hex, 1 byte)").(*tview.InputField).GetText()
	keyStr = strings.TrimSpace(keyStr)
	b, err := hex.DecodeString(keyStr)
	if err != nil || len(b) != 1 {
		s.setStatus("[red]Invalid hex key. Must be exactly 1 byte (e.g. 5a)")
		return
	}
	start, end := s.getActiveRange()
	s.pushUndo()
	crusher.Bitwise(s.currentData, start, end, b[0], op)
	s.updateHexTable()
	s.setStatus(fmt.Sprintf("[green]Applied %s bitwise operation with key %02x", op, b[0]))
}

func (s *AppState) showInputModal(title, label, initialValue string, onOk func(text string)) {
	form := tview.NewForm().
		AddInputField(label, initialValue, 40, nil, nil)

	form.AddButton("OK", func() {
		val := form.GetFormItemByLabel(label).(*tview.InputField).GetText()
		s.pages.RemovePage("modal")
		onOk(val)
		s.app.SetFocus(s.hexTable)
	}).
		AddButton("Cancel", func() {
			s.pages.RemovePage("modal")
			s.app.SetFocus(s.hexTable)
		})

	form.SetBorder(true).SetTitle(" " + title + " ")

	modal := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(form, 10, 1, true).
			AddItem(nil, 0, 1, false), 90, 1, true).
		AddItem(nil, 0, 1, false)

	s.pages.AddPage("modal", modal, true, true)
	s.app.SetFocus(form)
}

func (s *AppState) updateHexTableColorsOnly() {
	if s.hexTable.GetRowCount() == 0 { return }
	
	// Fast path to only redraw colors for selection without recalculating sizes
	min, max := s.getActiveRange()
	
	for line := 0; line < s.hexTable.GetRowCount(); line++ {
		start := line * 16
		for c := 1; c <= 16; c++ {
			i := start + c - 1
			cell := s.hexTable.GetCell(line, c)
			if cell == nil || cell.Text == "  " { continue } // out of bounds
			
			// Setup color
			color := tcell.ColorGreen
			if i < s.safeHeaderSize {
				color = tcell.ColorGray
			}
			
			// Overwrite if inside selection mode
			if s.selectionActive && i >= min && i < max {
				// Selection highlights
				cell.SetBackgroundColor(tcell.ColorDarkCyan)
				cell.SetTextColor(tcell.ColorWhite)
			} else {
				cell.SetBackgroundColor(tcell.ColorBlack)
				cell.SetTextColor(color)
			}
		}
	}
}

func (s *AppState) updateHexTable() {
	s.hexTable.Clear()
	if s.filePath == "" {
		s.hexTable.SetCell(0, 0, tview.NewTableCell("No file opened. Use 'Open File' from Actions panel.").SetTextColor(tcell.ColorGray))
		return
	}

	maxLines := 2048

	bytesPerLine := 16
	lines := len(s.currentData) / bytesPerLine
	if len(s.currentData)%bytesPerLine != 0 {
		lines++
	}

	if lines > maxLines {
		lines = maxLines
	}

	for line := 0; line < lines; line++ {
		start := line * bytesPerLine
		end := start + bytesPerLine
		if end > len(s.currentData) {
			end = len(s.currentData)
		}

		addrCell := tview.NewTableCell(fmt.Sprintf("%08x ", start)).
			SetTextColor(tcell.ColorYellow).
			SetSelectable(false).
			SetAlign(tview.AlignRight)
		s.hexTable.SetCell(line, 0, addrCell)

		for i := start; i < start+bytesPerLine; i++ {
			c := i - start + 1
			if i < len(s.currentData) {
				cell := tview.NewTableCell(fmt.Sprintf("%02x", s.currentData[i])).
					SetSelectable(true).
					SetAlign(tview.AlignCenter)
				s.hexTable.SetCell(line, c, cell)
			} else {
				cell := tview.NewTableCell("  ").SetSelectable(false)
				s.hexTable.SetCell(line, c, cell)
			}
		}

		asciiStr := " | "
		for i := start; i < end; i++ {
			c := s.currentData[i]
			if c >= 32 && c <= 126 {
				asciiStr += string(c)
			} else {
				asciiStr += "."
			}
		}

		asciiCell := tview.NewTableCell(asciiStr).
			SetTextColor(tcell.ColorWhite).
			SetSelectable(false).
			SetAlign(tview.AlignLeft)
		s.hexTable.SetCell(line, 17, asciiCell)
	}
	
	s.updateHexTableColorsOnly() // apply initial colors and selection
}

func (s *AppState) setStatus(msg string) {
	s.statusLine.SetText(fmt.Sprintf(" Status: %s\n Header Protection: %d bytes.", msg, s.safeHeaderSize))
}
