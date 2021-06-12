package main

// An example demonstrating an application with multiple views.
//
// Note that this example was produced before the Bubbles progress component
// was available (github.com/charmbracelet/bubbles/progress) and thus, we're
// implementing a progress bar from scratch here.

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	_ "embed"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/muesli/reflow/indent"
	"github.com/muesli/termenv"
)

// General stuff for styling the view
var (
	term   = termenv.ColorProfile()
	subtle = makeFgStyle("241")
	dot    = colorFg(" • ", "236")
	//go:embed users.txt
	usersf string
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

	// https://www.name-generator.org.uk/quick/
	p := tea.NewProgram(initialModel())
	if err := p.Start(); err != nil {
		fmt.Println("could not start program:", err)
	}
}

type model struct {
	Choice    int
	Chosen    bool
	Quitting  bool
	textInput textinput.Model
	choices   []string
	filtered  []string
	nvis      int
	Shift     int
	lastValue string
	async     bool
}

func initialModel() tea.Model {

	ti := textinput.NewModel()
	ti.Placeholder = "<Select User>"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20

	list := strings.Split(usersf, "\n")
	list = nil

	initialModel := model{
		Choice:    -1,
		Chosen:    false,
		Quitting:  false,
		textInput: ti,
		choices:   list,
		nvis:      8,
		Shift:     0,
		async:     true,
	}
	initialModel.filtered = initialModel.choices
	if (initialModel.choices == nil || len(initialModel.choices) == 0) && !initialModel.async {
		log.Fatalf("Empty initial list means async should be set")
	}
	return &initialModel
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

// Main update function.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Make sure these keys always quit
	if msg, ok := msg.(tea.KeyMsg); ok {
		k := msg.String()
		if (k == "esc" && m.textInput.Value() == "") || k == "ctrl+c" {
			m.Quitting = true
			return m, tea.Quit
		}
	}

	// Hand off the message and model to the appropriate update function for the
	// appropriate view based on the current state.
	if !m.Chosen {
		return updateChoices(msg, m)
	}
	m.Quitting = true
	return m, tea.Quit
}

// The main view, which just calls the appropriate sub-view
func (m model) View() string {
	var s string
	if m.Quitting {
		return ""
	}
	if !m.Chosen {
		s = choicesView(&m)
	} else {
		return ""
	}
	return indent.String("\n"+s+"\n\n", 2)
}

func (m *model) getNVisible() int {
	l := len(m.filtered)
	if l < m.nvis {
		return l
	}
	return m.nvis
}

// Sub-update functions

// Update loop for the first view where you're choosing a task.
func updateChoices(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	lfiltered := len(m.filtered)
	esc := false
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		case "down", "tab":
			m.Choice += 1
			if m.Choice >= lfiltered {
				m.Choice = -1
				m.updateText(m.lastValue)
			} else {
				m.updateText(m.filtered[m.Choice])
			}
		case "up", "shift+tab":
			m.Choice -= 1
			if m.Choice <= -1 {
				m.Choice = lfiltered - 1
			}
			if m.Choice >= 0 {
				m.updateText(m.filtered[m.Choice])
			}
		case "enter":
			m.Chosen = true
			return m, nil
		case "esc":
			if m.Choice >= 0 {
				m.Choice = -1
				m.updateText(m.lastValue)
				esc = true
			} else if m.textInput.Value() != "" {
				m.filtered = m.choices
				m.updateText("")
				m.lastValue = ""
			}
		}
	}

	nvis := m.getNVisible()
	shift := m.Choice - nvis + 1
	if shift > m.Shift {
		m.Shift = shift
	}
	if m.Choice >= 0 && m.Choice+nvis < m.Shift+nvis {
		m.Shift = m.Choice
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	v := m.textInput.Value()
	//v = "T"
	if v != "" && (v != m.lastValue || esc) && (m.Choice < 0 || v != m.filtered[m.Choice]) {
		m.filtered = m.filter(v)
		m.Choice = -1
		m.Shift = 0
		m.lastValue = v
	}
	if v == "" {
		m.Choice = -1
		m.Shift = 0
		m.lastValue = ""
		m.filtered = m.choices
	}
	return m, cmd
}

func (m *model) updateText(text string) {
	m.textInput.SetValue(text)
	m.textInput.SetCursor(len(text))
}

func (m model) filter(v string) []string {
	source := v
	matches := fuzzy.RankFindNormalizedFold(source, m.choices)
	sort.Sort(matches) // [{whl wheel 2 2} {whl cartwheel 6 0}]
	res := make([]string, 0)
	//dbg := fmt.Sprintf("matches for source '%s': ", source)
	for _, match := range matches {
		//dbg = dbg + fmt.Sprintf(" (%s %d <= %d)", match.Target, match.Distance, len(m.choices[match.OriginalIndex]))
		if match.Distance < 0 {
			continue
		}
		if match.Distance > len(m.choices[match.OriginalIndex]) {
			continue
		}
		res = append(res, match.Target)
	}
	//log.Println(dbg)
	return res
}

// Sub-views

// The first view, where you're choosing a task
func choicesView(m *model) string {

	tpl := "What to do today?\n\n"
	tpl += "%s\n\n"
	tpl += subtle("<esc>: clean/exit, up/down: select") + dot + subtle("enter: choose")

	choice := m.Choice
	choices := m.textInput.View()
	nvis := m.getNVisible()
	shift := m.Shift
	debug := fmt.Sprintf("\nlen(filtered): %d ~ choice: %d ~ nvis: %d ~ m.shift: %d", len(m.filtered), choice, nvis, shift)
	choices = choices + debug
	for i := 0; i < m.nvis; i++ {
		if i < nvis {
			choices = choices + "\n" + checkbox(m.filtered[i+shift], i+shift == choice)
		}
	}
	return fmt.Sprintf(tpl, choices)
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
