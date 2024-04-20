package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type clientModel struct {
	name   string
	input  textinput.Model
	status string
}

type paste struct {
	Name      string `json:"name"`
	PasteBody string `json:"pasteBody"`
}

func newClient() *clientModel {
	model := &clientModel{
		input: textinput.New(),
	}
	model.input.Placeholder = "Your text"
	model.input.CharLimit = 256
	model.input.Width = 30

	model.input.Focus()

	model.status = "Please register"

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

			m.status = ""

			if newMsg == "" {
				m.status = "Invalid message"
				return m, nil
			}

			if m.name == "" {
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
					m.status = "Registered successfully, welcome " + newMsg
					m.name = newMsg
				} else {
					m.status = "Could not register, name in use"
				}
				res.Body.Close()
			} else {
				data, err := json.Marshal(paste{m.name, newMsg})
				if err != nil {
					log.Fatal("Could not parse name to json")
				}

				req, err := http.NewRequest(http.MethodPost, "http://localhost:8080/paste", bytes.NewBuffer(data))
				if err != nil {
					log.Fatal("Could not form paste request")
				}

				res, err := http.DefaultClient.Do(req)
				if err != nil {
					log.Fatal("Could not send paste request", err)
				}

				if res.StatusCode == http.StatusAccepted {
					body, err := io.ReadAll(res.Body)
					if err != nil {
						log.Fatal("Could not parse response body")
					}
					m.status = "Pasted successfully at: " + string(body)
				} else {
					m.status = "Could not paste"
				}
				res.Body.Close()
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

	ui.WriteString("\n" + m.status + "\n\n")

	if m.name == "" {
		ui.WriteString("Enter username: \n")
	} else {
		ui.WriteString("Enter message to paste: \n")
	}

	ui.WriteString("\n<------------------------------------->\n")

	ui.WriteString(m.input.View())

	return ui.String()
}
