package main

// An example demonstrating an application with multiple views.
//
// Note that this example was produced before the Bubbles progress component
// was available (github.com/charmbracelet/bubbles/progress) and thus, we're
// implementing a progress bar from scratch here.

import (
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/reflow/indent"
	"github.com/muesli/termenv"
)

// General stuff for styling the view
var (
	term    = termenv.ColorProfile()
	keyword = makeFgStyle("211")
	subtle  = makeFgStyle("241")
	dot     = colorFg(" • ", "236")
)

func test() {

	// Set PAGER_LOG to a path to log to a file. For example:
	//
	//     export PAGER_LOG=debug.log
	//
	// This becomes handy when debugging stuff since you can't debug to stdout
	// because the UI is occupying it!
	path := os.Getenv("PAGER_LOG")
	if path != "" {
		f, err := tea.LogToFile(path, "pager")
		if err != nil {
			fmt.Printf("Could not open file %s: %v", path, err)
			os.Exit(1)
		}
		defer f.Close()
	}

	initialModel := model{0, false, 0, false, false}
	p := tea.NewProgram(initialModel)
	if err := p.Start(); err != nil {
		fmt.Println("could not start program:", err)
	}
}

type frameMsg struct{}

func frame() tea.Cmd {
	return tea.Tick(time.Second/60, func(time.Time) tea.Msg {
		return frameMsg{}
	})
}

type model struct {
	Choice   int
	Chosen   bool
	Frames   int
	Loaded   bool
	Quitting bool
}

func (m model) Init() tea.Cmd {
	return frame()
}

// Main update function.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Make sure these keys always quit
	if msg, ok := msg.(tea.KeyMsg); ok {
		k := msg.String()
		if k == "q" || k == "esc" || k == "ctrl+c" {
			m.Quitting = true
			return m, tea.Quit
		}
	}

	// Hand off the message and model to the appropriate update function for the
	// appropriate view based on the current state.
	if !m.Chosen {
		return updateChoices(msg, m)
	}
	return updateChosen(msg, m)
}

// The main view, which just calls the appropriate sub-view
func (m model) View() string {
	var s string
	if m.Quitting {
		return ""
		//return "\n  See you later!\n\n"
	}
	if !m.Chosen {
		s = choicesView(m)
	} else {
		s = chosenView(m)
	}
	return indent.String("\n"+s+"\n\n", 2)
}

// Sub-update functions

// Update loop for the first view where you're choosing a task.
func updateChoices(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		case "down":
			m.Choice += 1
			if m.Choice > 3 {
				m.Choice = 0
			}
		case "up":
			m.Choice -= 1
			if m.Choice < 0 {
				m.Choice = 3
			}
		case "enter":
			m.Chosen = true
			return m, frame()
		}
	}

	return m, nil
}

// Update loop for the second view after a choice has been made
func updateChosen(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	switch msg.(type) {

	case frameMsg:
		if !m.Loaded {
			m.Frames += 1
			return m, frame()
		}

	}

	return m, nil
}

// Sub-views

// The first view, where you're choosing a task
func choicesView(m model) string {
	c := m.Choice

	tpl := "What to do today?\n\n"
	tpl += "%s\n\n"
	tpl += subtle("j/k, up/down: select") + dot + subtle("enter: choose") + dot + subtle("q, esc: quit")

	choices := fmt.Sprintf(
		"%s\n%s\n%s\n%s",
		checkbox("Plant carrots", c == 0),
		checkbox("Go to the market", c == 1),
		checkbox("Read something", c == 2),
		checkbox("See friends", c == 3),
	)

	return fmt.Sprintf(tpl, choices)
}

// The second view, after a task has been chosen
func chosenView(m model) string {
	var msg string

	switch m.Choice {
	case 0:
		msg = fmt.Sprintf("Carrot planting?\n\nCool, we'll need %s and %s...", keyword("libgarden"), keyword("vegeutils"))
	case 1:
		msg = fmt.Sprintf("A trip to the market?\n\nOkay, then we should install %s and %s...", keyword("marketkit"), keyword("libshopping"))
	case 2:
		msg = fmt.Sprintf("Reading time?\n\nOkay, cool, then we’ll need a library. Yes, an %s.", keyword("actual library"))
	default:
		msg = fmt.Sprintf("It’s always good to see friends.\n\nFetching %s and %s...", keyword("social-skills"), keyword("conversationutils"))
	}

	label := "Downloading..."
	if m.Loaded {
		label = "Downloaded."
	}

	return msg + "\n\n" + label + "\n"
}

func checkbox(label string, checked bool) string {
	if checked {
		return colorFg("[x] "+label, "212")
	}
	return fmt.Sprintf("[ ] %s", label)
}

// Utils

// Color a string's foreground with the given value.
func colorFg(val, color string) string {
	return termenv.String(val).Foreground(term.Color(color)).String()
}

// Return a function that will colorize the foreground of a given string.
func makeFgStyle(color string) func(string) string {
	return termenv.Style{}.Foreground(term.Color(color)).Styled
}
