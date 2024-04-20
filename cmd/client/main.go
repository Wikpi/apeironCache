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
	status status
}

type status struct {
	name    string
	message strings.Builder
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

	model.status.message.WriteString(getContinueMessage(model))

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

			m.status.message.Reset()

			if newMsg == "" {
				m.status.message.WriteString(getContinueMessage(m))
				return m, nil
			}

			command := strings.Split(newMsg, " ")[0]
			newMsg = strings.Join(strings.Split(newMsg, " ")[1:], " ")

			switch command {
			case "register":
				m.registerClient(newMsg)
			case "login":
				m.loginClient(newMsg)
			case "paste":
				m.pasteClientRequest(newMsg)
			case "get":
				m.getClientRequest(newMsg)
			}
			m.status.message.WriteString("\n\nPress Enter to continue")

			/*timer := *time.NewTimer(time.Minute)

			go func() {
				<-timer.C
				timer.Stop()

				m.status.message.Reset()
				m.status.message.WriteString(getContinueMessage(m))
			}()*/

		case tea.KeyCtrlC:
			return m, tea.Quit
		}
	}

	m.input, cmd = m.input.Update(msg)

	return m, cmd
}

func getContinueMessage(m *clientModel) string {
	message := strings.Builder{}

	message.WriteString("Use ")

	if m.name == "" {
		message.WriteString("register, login, ")
	}
	message.WriteString("paste or get")

	return message.String()
}

func (m *clientModel) registerClient(msg string) {
	m.status.name = "register"

	data, err := json.Marshal(msg)
	if err != nil {
		m.status.message.WriteString("Could not parse name to json")
		return
	}

	req, err := http.NewRequest(http.MethodPost, "http://localhost:8080/register", bytes.NewBuffer(data))
	if err != nil {
		m.status.message.WriteString("Could not form name request")
		return
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		m.status.message.WriteString("Could not send name request")
		return
	}

	if res.StatusCode == http.StatusAccepted {
		m.name = msg
		m.status.message.WriteString("Registered successfully, welcome " + m.name)
	} else {
		m.status.message.WriteString("Could not register, name in use")
	}
	res.Body.Close()
}

func (m *clientModel) loginClient(_ string) {
	m.status.name = "login"
}

func (m *clientModel) pasteClientRequest(msg string) {
	m.status.name = "paste"

	data, err := json.Marshal(paste{"name", msg})
	if err != nil {
		m.status.message.WriteString("Could not parse paste to json")
		return
	}

	req, err := http.NewRequest(http.MethodPost, "http://localhost:8080/paste", bytes.NewBuffer(data))
	if err != nil {
		m.status.message.WriteString("Could not form paste request")
		return
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		m.status.message.WriteString("Could not send paste request")
		return
	}

	if res.StatusCode == http.StatusAccepted {
		body, err := io.ReadAll(res.Body)
		if err != nil {
			m.status.message.WriteString("Could not parse response body")
			return
		}
		m.status.message.WriteString("Pasted successfully at: " + string(body))
	} else {
		m.status.message.WriteString("Could not paste")
	}
	res.Body.Close()
}

func (m *clientModel) getClientRequest(code string) {
	m.status.name = "get"

	data, err := json.Marshal(code)
	if err != nil {
		m.status.message.WriteString("Could not parse code to json")
		return
	}

	req, err := http.NewRequest(http.MethodPost, "http://localhost:8080/get", bytes.NewBuffer(data))
	if err != nil {
		m.status.message.WriteString("Could not form get request")
		return
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		m.status.message.WriteString("Could not send get request")
		return
	}

	if res.StatusCode == http.StatusAccepted {
		body, err := io.ReadAll(res.Body)
		if err != nil {
			m.status.message.WriteString("Could not read get response body")
			return
		}
		request := paste{}

		// requestMessage := formatText(request.PasteBody)

		if err := json.Unmarshal(body, &request); err != nil {
			m.status.message.WriteString("Could not parse get response body")
		}
		requestMessage := request.PasteBody

		m.status.message.WriteString("Paste request at " + code + ": \n\n" + requestMessage)
	} else {
		m.status.message.WriteString("Could not paste")
	}
	res.Body.Close()
}

func formatText(text string) string {
	formatedText := strings.Builder{}

	words := strings.Split(text, " ")

	for i := 0; i < len(words); i += 16 {
		formatedText.WriteString(strings.Join(words[i:i+15], " "))
		formatedText.WriteString("\n")
	}

	return formatedText.String()
}

func (m clientModel) View() string {
	ui := strings.Builder{}

	ui.WriteString("\n" + m.status.message.String() + "\n")

	ui.WriteString("\n<------------------------------------->\n")

	ui.WriteString(m.input.View())

	return ui.String()
}
