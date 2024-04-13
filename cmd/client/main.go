package main

import (
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	Id       int
	Name     string
	input    textinput.Model
	messages []message
}

type message struct {
	Name    string
	Message string
}

func newClient() *model {
	model := &model{
		input: textinput.New(),
	}
	model.input.Placeholder = "Your text"
	model.input.CharLimit = 256
	model.input.Width = 30

	model.input.Focus()

	return model
}

func main() {
	newProgram := tea.NewProgram(newClient())
	if _, err := newProgram.Run(); err != nil {
		log.Fatal("Server is down")
	}
}

func (m *model) Init() tea.Cmd {
	return textinput.Blink
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch key := msg.(type) {
	case tea.KeyMsg:
		switch key.Type {
		case tea.KeyEnter:
			newMsg := m.input.Value()

			if newMsg == "" {
				return m, nil
			}
			m.messages = append(m.messages, message{"User", newMsg})

			m.input.SetValue("")
		case tea.KeyCtrlC:
			return m, tea.Quit
		}
	}

	m.input, cmd = m.input.Update(msg)

	return m, cmd
}

func (m model) View() string {
	ui := strings.Builder{}

	ui.WriteString("\n")

	for _, message := range m.messages {
		ui.WriteString(message.Name + " " + message.Message + "\n")
	}

	ui.WriteString("\n<------------------------------------->\n")

	ui.WriteString(m.input.View())

	return ui.String()
}
