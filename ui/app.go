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

const safeHeaderSize = 512

type AppState struct {
	filePath     string
	originalData []byte
	currentData  []byte
	hexTable     *tview.Table
	app          *tview.Application
	pages        *tview.Pages
	statusLine   *tview.TextView
}

func (s *AppState) loadFile(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	s.filePath = path
	s.originalData = append([]byte{}, data...)
	s.currentData = append([]byte{}, data...)
	return nil
}

func RunTUI(initialFilePath string) error {
	state := &AppState{
		app:   tview.NewApplication(),
		pages: tview.NewPages(),
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

	if state.filePath != "" {
		state.setStatus("App loaded successfully. Press 'Enter' on a byte to edit, 'Del' to remove row, 'R' to replace row.")
	} else {
		state.setStatus("Welcome to Byte Crusher! Please click 'Open File' to begin.")
	}

	sidePanel := tview.NewFlex().SetDirection(tview.FlexRow)
	sidePanel.SetBorder(true).SetTitle(" Operations Panel ")

	// --- Replace Form ---
	replaceForm := tview.NewForm()
	replaceForm.AddInputField("Find (Hex)", "", 10, nil, nil)
	replaceForm.AddInputField("Replace (Hex)", "", 10, nil, nil)
	replaceForm.AddButton("Replace", func() {
		if state.filePath == "" {
			state.setStatus("[red]No file loaded!")
			return
		}
		fromStr := replaceForm.GetFormItemByLabel("Find (Hex)").(*tview.InputField).GetText()
		toStr := replaceForm.GetFormItemByLabel("Replace (Hex)").(*tview.InputField).GetText()

		fromStr = strings.ReplaceAll(fromStr, " ", "")
		toStr = strings.ReplaceAll(toStr, " ", "")

		from, err1 := hex.DecodeString(fromStr)
		to, err2 := hex.DecodeString(toStr)

		if err1 != nil || err2 != nil {
			state.setStatus("[red]Invalid hex string in Find/Replace")
			return
		}

		if len(from) == 0 {
			state.setStatus("[red]Find string cannot be empty")
			return
		}

		state.currentData = crusher.ReplacePattern(state.currentData, safeHeaderSize, from, to)
		state.updateHexTable()
		state.setStatus(fmt.Sprintf("[green]Replaced %s with %s", fromStr, toStr))
	})
	replaceForm.SetTitle(" Replace bytes ").SetBorder(true)
	sidePanel.AddItem(replaceForm, 9, 1, false)

	// --- Corrupt Form ---
	corruptForm := tview.NewForm()
	corruptForm.AddInputField("Severity (0.01-1.0)", "0.05", 10, nil, nil)
	corruptForm.AddButton("Corrupt", func() {
		if state.filePath == "" {
			state.setStatus("[red]No file loaded!")
			return
		}
		sevStr := corruptForm.GetFormItemByLabel("Severity (0.01-1.0)").(*tview.InputField).GetText()
		sev, err := strconv.ParseFloat(sevStr, 64)
		if err != nil {
			state.setStatus("[red]Invalid severity value")
			return
		}

		crusher.Corrupt(state.currentData, safeHeaderSize, sev)
		state.updateHexTable()
		state.setStatus(fmt.Sprintf("[green]Corrupted %.2f%% of data", sev*100))
	})
	corruptForm.SetTitle(" Byte Corruption ").SetBorder(true)
	sidePanel.AddItem(corruptForm, 7, 1, false)

	// --- Edit Actions Form ---
	editForm := tview.NewForm()
	editForm.AddButton("Jump (Hex)", func() {
		if state.filePath == "" {
			return
		}
		state.showInputModal("Jump to Offset", "Target Hex (e.g. 1a0):", "", func(newVal string) {
			newVal = strings.TrimSpace(newVal)
			newVal = strings.TrimPrefix(newVal, "0x")
			addr, err := strconv.ParseInt(newVal, 16, 64)
			if err != nil {
				state.setStatus("[red]Invalid hex address format")
				return
			}
			row := int(addr / 16)
			if row < 0 {
				row = 0
			}
			if row >= state.hexTable.GetRowCount() {
				row = state.hexTable.GetRowCount() - 1
			}
			if row < 0 {
				row = 0
			}
			state.hexTable.Select(row, 1)
		})
	})
	editForm.AddButton("Reset Mod", func() {
		if state.filePath == "" {
			return
		}
		state.currentData = append([]byte{}, state.originalData...)
		state.updateHexTable()
		state.hexTable.ScrollToBeginning()
		state.hexTable.Select(0, 1)
		state.setStatus("[yellow]Reset to original file")
	})
	editForm.SetTitle(" Navigation ").SetBorder(true)
	sidePanel.AddItem(editForm, 5, 1, false)

	// --- File Actions Form ---
	fileForm := tview.NewForm()
	fileForm.AddButton("Open File", func() {
		state.showInputModal("Open New File", "File Path:", state.filePath, func(newVal string) {
			newVal = strings.TrimSpace(newVal)
			if newVal == "" {
				return
			}
			err := state.loadFile(newVal)
			if err != nil {
				state.setStatus(fmt.Sprintf("[red]Failed to load: %v", err))
			} else {
				state.hexTable.SetTitle(fmt.Sprintf(" Hex Table - %s ", state.filePath))
				state.updateHexTable()
				state.hexTable.ScrollToBeginning()
				state.hexTable.Select(0, 1)
				state.setStatus(fmt.Sprintf("[green]Successfully loaded %s", state.filePath))
			}
		})
	})
	fileForm.AddButton("Save File", func() {
		if state.filePath == "" {
			state.setStatus("[red]No file loaded to save!")
			return
		}
		err := ioutil.WriteFile(state.filePath, state.currentData, 0644)
		if err != nil {
			state.setStatus(fmt.Sprintf("[red]Save failed: %v", err))
		} else {
			state.setStatus("[green]File saved successfully!")
			state.originalData = append([]byte{}, state.currentData...) // update original
		}
	})
	fileForm.AddButton("Quit", func() {
		state.app.Stop()
	})
	fileForm.SetTitle(" Main Actions ").SetBorder(true)
	sidePanel.AddItem(fileForm, 0, 1, false)

	// --- Layout Setup ---
	mainFlex := tview.NewFlex().
		AddItem(state.hexTable, 0, 3, true).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(sidePanel, 0, 8, false).
			AddItem(state.statusLine, 3, 1, false), 0, 1, false)

	state.pages.AddPage("main", mainFlex, true, true)
	state.app.SetRoot(state.pages, true).EnableMouse(true)

	// App-level global shortcuts (Tab switching)
	state.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlQ {
			state.app.Stop()
			return nil
		}
		if event.Key() == tcell.KeyTab {
			if state.hexTable.HasFocus() {
				state.app.SetFocus(sidePanel)
			} else {
				state.app.SetFocus(state.hexTable)
			}
			return nil
		}
		return event
	})

	// Table-specific interactive shortcuts
	state.hexTable.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if state.filePath == "" {
			return event // do nothing if no file
		}

		row, col := state.hexTable.GetSelection()
		if col < 1 || col > 16 {
			// Jump column if they are on borders
			if event.Key() == tcell.KeyLeft && col > 1 {
				state.hexTable.Select(row, 16)
				return nil
			} else if event.Key() == tcell.KeyRight && col < 17 {
				state.hexTable.Select(row, 1)
				return nil
			}
			return event // pass through if not on a byte column
		}

		byteIndex := row*16 + (col - 1)

		isProtected := byteIndex < safeHeaderSize
		wantsModify := event.Key() == tcell.KeyEnter || event.Key() == tcell.KeyBackspace || event.Key() == tcell.KeyDelete || event.Key() == tcell.KeyBackspace2 || event.Rune() == 'R' || event.Rune() == 'r'

		if isProtected && wantsModify {
			state.setStatus("[red]Cannot modify header bytes manually! Security lock is active.")
			return nil // block
		}

		if byteIndex >= len(state.currentData) {
			return event
		}

		// Edit single byte
		if event.Key() == tcell.KeyEnter {
			currentByteStr := fmt.Sprintf("%02x", state.currentData[byteIndex])
			state.showInputModal(fmt.Sprintf("Edit Byte at %08x", byteIndex), "New Hex Value:", currentByteStr, func(newVal string) {
				newVal = strings.TrimSpace(newVal)
				b, err := hex.DecodeString(newVal)
				if err != nil || len(b) != 1 {
					state.setStatus("[red]Invalid hex format. Must be exactly 1 byte (e.g. FF)")
					return
				}
				state.currentData[byteIndex] = b[0]
				state.updateHexTable()
				state.hexTable.Select(row, col)
				state.setStatus(fmt.Sprintf("[green]Updated byte to %02x", b[0]))
			})
			return nil
		}

		// Delete Row
		if event.Key() == tcell.KeyBackspace || event.Key() == tcell.KeyDelete || event.Key() == tcell.KeyBackspace2 {
			startIdx := row * 16
			if startIdx < safeHeaderSize {
				state.setStatus("[red]Cannot delete row in protected header territory!")
				return nil
			}
			endIdx := startIdx + 16
			if endIdx > len(state.currentData) {
				endIdx = len(state.currentData)
			}

			state.currentData = append(state.currentData[:startIdx], state.currentData[endIdx:]...)
			state.updateHexTable()

			// Adjust cursor if row deleted
			if row >= state.hexTable.GetRowCount() {
				row = state.hexTable.GetRowCount() - 1
			}
			if row < 0 {
				row = 0
			}
			state.hexTable.Select(row, col)
			state.setStatus(fmt.Sprintf("[yellow]Deleted row at address %08x", startIdx))
			return nil
		}

		// Replace Row
		if event.Rune() == 'R' || event.Rune() == 'r' {
			startIdx := row * 16
			endIdx := startIdx + 16
			if endIdx > len(state.currentData) {
				endIdx = len(state.currentData)
			}

			existingHex := hex.EncodeToString(state.currentData[startIdx:endIdx])

			state.showInputModal(fmt.Sprintf("Replace %d bytes starting %08x", endIdx-startIdx, startIdx), "New Sequence (Hex):", existingHex, func(newVal string) {
				newVal = strings.ReplaceAll(newVal, " ", "")
				b, err := hex.DecodeString(newVal)
				if err != nil {
					state.setStatus("[red]Invalid hex string provided.")
					return
				}

				var newData []byte
				newData = append(newData, state.currentData[:startIdx]...)
				newData = append(newData, b...)
				newData = append(newData, state.currentData[endIdx:]...)

				state.currentData = newData
				state.updateHexTable()
				state.hexTable.Select(row, col)
				state.setStatus(fmt.Sprintf("[green]Replaced row data successfully! Row now has %d bytes.", len(b)))
			})
			return nil
		}

		return event
	})

	state.updateHexTable()

	// If launched with no args, show the open file modal immediately
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

func (s *AppState) showInputModal(title, label, initialValue string, onOk func(text string)) {
	form := tview.NewForm().
		AddInputField(label, initialValue, 40, nil, nil) // adjusted input to 40 spaces

	form.AddButton("OK", func() {
		val := form.GetFormItemByLabel(label).(*tview.InputField).GetText()
		s.pages.RemovePage("modal")
		onOk(val)
		s.app.SetFocus(s.hexTable) // move focus AFTER ok callback builds the table
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
			AddItem(nil, 0, 1, false), 90, 1, true). // widened modal box to 90 chars
		AddItem(nil, 0, 1, false)

	s.pages.AddPage("modal", modal, true, true)
	s.app.SetFocus(form)
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

		// Col 0: Address
		addrCell := tview.NewTableCell(fmt.Sprintf("%08x ", start)).
			SetTextColor(tcell.ColorYellow).
			SetSelectable(false).
			SetAlign(tview.AlignRight)
		s.hexTable.SetCell(line, 0, addrCell)

		// Col 1-16: Bytes
		for i := start; i < start+bytesPerLine; i++ {
			c := i - start + 1
			if i < len(s.currentData) {
				color := tcell.ColorGreen
				if i < safeHeaderSize {
					color = tcell.ColorGray
				}
				cell := tview.NewTableCell(fmt.Sprintf("%02x", s.currentData[i])).
					SetTextColor(color).
					SetSelectable(true).
					SetAlign(tview.AlignCenter)
				s.hexTable.SetCell(line, c, cell)
			} else {
				cell := tview.NewTableCell("  ").SetSelectable(false)
				s.hexTable.SetCell(line, c, cell)
			}
		}

		// Col 17: ASCII
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
}

func (s *AppState) setStatus(msg string) {
	s.statusLine.SetText(fmt.Sprintf(" Status: %s\n Header Protection: %d bytes.", msg, safeHeaderSize))
}
