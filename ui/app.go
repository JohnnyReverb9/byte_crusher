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
	hexView      *tview.TextView
	app          *tview.Application
	statusLine   *tview.TextView
}

func RunTUI(filePath string) error {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("could not read file: %w", err)
	}

	state := &AppState{
		filePath:     filePath,
		originalData: append([]byte{}, data...),
		currentData:  append([]byte{}, data...),
		app:          tview.NewApplication(),
	}

	state.hexView = tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWrap(false).
		SetScrollable(true)
	state.hexView.SetBorder(true).SetTitle(fmt.Sprintf(" Hex Dump - %s ", filePath))

	state.statusLine = tview.NewTextView().SetDynamicColors(true)
	state.statusLine.SetBorder(true)
	state.setStatus("App loaded successfully")

	sidePanel := tview.NewFlex().SetDirection(tview.FlexRow)
	sidePanel.SetBorder(true).SetTitle(" Operations Panel ")

	// --- Replace Form ---
	replaceForm := tview.NewForm()
	replaceForm.AddInputField("Find (Hex)", "", 10, nil, nil)
	replaceForm.AddInputField("Replace (Hex)", "", 10, nil, nil)
	replaceForm.AddButton("Replace", func() {
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
		state.updateHexView()
		state.setStatus(fmt.Sprintf("[green]Replaced %s with %s", fromStr, toStr))
	})
	replaceForm.SetTitle(" Replace bytes ").SetBorder(true)
	sidePanel.AddItem(replaceForm, 9, 1, false)

	// --- Corrupt Form ---
	corruptForm := tview.NewForm()
	corruptForm.AddInputField("Severity (0.01-1.0)", "0.05", 10, nil, nil)
	corruptForm.AddButton("Corrupt", func() {
		sevStr := corruptForm.GetFormItemByLabel("Severity (0.01-1.0)").(*tview.InputField).GetText()
		sev, err := strconv.ParseFloat(sevStr, 64)
		if err != nil {
			state.setStatus("[red]Invalid severity value")
			return
		}

		crusher.Corrupt(state.currentData, safeHeaderSize, sev)
		state.updateHexView()
		state.setStatus(fmt.Sprintf("[green]Corrupted %.2f%% of data", sev*100))
	})
	corruptForm.SetTitle(" Byte Corruption ").SetBorder(true)
	sidePanel.AddItem(corruptForm, 7, 1, false)

	// --- Global Actions Form ---
	globalForm := tview.NewForm()
	globalForm.AddButton("Reset", func() {
		state.currentData = append([]byte{}, state.originalData...)
		state.updateHexView()
		state.setStatus("[yellow]Reset to original file")
	})
	globalForm.AddButton("Save File", func() {
		err := ioutil.WriteFile(state.filePath, state.currentData, 0644)
		if err != nil {
			state.setStatus(fmt.Sprintf("[red]Save failed: %v", err))
		} else {
			state.setStatus("[green]File saved successfully!")
			state.originalData = append([]byte{}, state.currentData...) // update original
		}
	})
	globalForm.AddButton("Quit", func() {
		state.app.Stop()
	})
	globalForm.SetTitle(" Actions ").SetBorder(true)
	sidePanel.AddItem(globalForm, 0, 1, false)

	// --- Layout Setup ---
	flex := tview.NewFlex().
		AddItem(state.hexView, 0, 3, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(sidePanel, 0, 8, true).
			AddItem(state.statusLine, 3, 1, false), 0, 1, true)

	state.app.SetRoot(flex, true).EnableMouse(true)
	state.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlQ {
			state.app.Stop()
			return nil
		}
		if event.Key() == tcell.KeyTab {
			state.app.SetFocus(flex) // rudimentary cycle focus, tview handles form tabs naturally
		}
		return event
	})

	state.updateHexView()

	return state.app.Run()
}

func (s *AppState) updateHexView() {
	// Rendering a max of 2000 lines (32KB approx) to prevent UI lag on huge files
	hexStr := crusher.HexDumpColored(s.currentData, safeHeaderSize, 2000)
	s.hexView.SetText(hexStr)
	s.hexView.ScrollToBeginning()
}

func (s *AppState) setStatus(msg string) {
	s.statusLine.SetText(fmt.Sprintf(" Status: %s\n Header Protection: First %d bytes locked.", msg, safeHeaderSize))
}
