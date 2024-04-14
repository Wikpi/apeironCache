package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type clientModel struct {
	Id       int
	Name     string
	input    textinput.Model
	messages []message
}

type message struct {
	Name    string
	Message string
}

func newClient() *clientModel {
	model := &clientModel{
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
		log.Fatal("Could not start client")
	}
}

func (m *clientModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m *clientModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch key := msg.(type) {
	case tea.KeyMsg:
		switch key.Type {
		case tea.KeyEnter:
			newMsg := m.input.Value()
			m.input.SetValue("")

			if newMsg == "" {
				return m, nil
			}

			if m.Name == "" {
				data, err := json.Marshal(newMsg)
				if err != nil {
					log.Fatal("Could not parse name to json")
				}

				req, err := http.NewRequest(http.MethodPost, "http://localhost:8080/register", bytes.NewBuffer(data))
				if err != nil {
					log.Fatal("Could not form name request")
				}

				res, err := http.DefaultClient.Do(req)
				if err != nil {
					log.Fatal("Could not send name request", err)
				}

				if res.StatusCode == http.StatusAccepted {
					m.Name = newMsg
				}
				return m, nil
			}
		case tea.KeyCtrlC:
			return m, tea.Quit
		}
	}

	m.input, cmd = m.input.Update(msg)

	return m, cmd
}

func (m clientModel) View() string {
	ui := strings.Builder{}

	ui.WriteString("Enter username: \n")
	ui.WriteString(m.Name + "\n")

	ui.WriteString("\n<------------------------------------->\n")

	ui.WriteString(m.input.View())

	return ui.String()
}
