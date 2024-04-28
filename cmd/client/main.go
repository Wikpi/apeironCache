package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

const textFlag = "-t"
const fileFlag = "-f"

type clientModel struct {
	name      string
	textInput textinput.Model
	textArea  textarea.Model
	status    status
}

type status struct {
	name    string
	message strings.Builder
}

type upload struct {
	User       string `json:"user"`
	Type       string `json:"type"`
	Size       string `json:"size"`
	UploadBody []byte `json:"uploadBody"`
}

func newClient() *clientModel {
	model := &clientModel{
		textInput: textinput.New(),
		textArea:  textarea.New(),
	}
	model.textInput.Placeholder = "Your text"
	model.textInput.CharLimit = 50
	model.textInput.Width = 50

	model.textArea.Placeholder = "Your text"
	model.textArea.CharLimit = 256
	model.textArea.Prompt = "| "
	model.textArea.SetWidth(30)

	model.textInput.Focus()

	model.status.name = "limbo"
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
			newMsg := m.textInput.Value()
			m.textInput.SetValue("")

			m.status.message.Reset()
			if newMsg == "" {
				m.status.message.WriteString("No command.\n")
				m.status.message.WriteString(getContinueMessage(m))
				return m, nil
			}

			words := strings.Split(newMsg, " ")

			command := words[0]

			switch command {
			case "r":
				m.registerClient(newMsg)
			case "l":
				m.loginClient(newMsg)
			case "p":
				if len(words) < 2 {
					m.status.message.WriteString("No flag.\n")
					break
				}
				flag := words[1]

				if len(words) < 3 {
					m.status.message.WriteString("No input.\n")
					break
				}
				input := strings.Join(words[2:], " ")

				m.uploadData(input, flag)
			case "g":
				if len(words) < 2 {
					m.status.message.WriteString("No code.\n")
					break
				}

				code := strings.Join(words[1:], " ")

				m.getClientRequest(code)
			default:
				m.status.message.WriteString("Incorrect command.\n")
			}
			m.status.message.WriteString("\n\nPress Enter to continue")

		case tea.KeyCtrlC:
			return m, tea.Quit
		}
	}

	m.textInput, cmd = m.textInput.Update(msg)

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

func (m *clientModel) uploadData(input string, flag string) error {
	// m.status.name = "paste"

	// Prepares JSON data to upload to the server
	data, err := m.prepareData(input, flag)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, "http://localhost:8080/paste", bytes.NewBuffer(data))
	if err != nil {
		m.status.message.WriteString("Could not form paste request")
		return err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		m.status.message.WriteString("Could not send paste request")
		return err
	}

	if res.StatusCode == http.StatusAccepted {
		body, err := io.ReadAll(res.Body)
		if err != nil {
			m.status.message.WriteString("Could not parse response body")
			return err
		}
		m.status.message.WriteString("Pasted successfully at: " + string(body))
	} else {
		m.status.message.WriteString("Could not paste")
	}
	res.Body.Close()

	return nil
}

func (m *clientModel) prepareData(input string, flag string) ([]byte, error) {
	newUpload := upload{}

	switch flag {
	case fileFlag:
		file, err := os.ReadFile(input)
		if err != nil {
			m.status.message.WriteString("Could not open file.\n")
			return nil, err
		}
		newUpload.UploadBody = file

	case textFlag:
		text, err := json.Marshal(input)
		if err != nil {
			m.status.message.WriteString("Could not parse upload text to json.\n")
			return nil, err
		}
		newUpload.UploadBody = text
	default:
		m.status.message.WriteString("Incorrect flag.\n")
		return nil, errors.New("Incorrect flag")
	}

	if m.name != "" {
		newUpload.User = m.name
	} else {
		newUpload.User = "Anonymous"
	}
	newUpload.Type = flag
	newUpload.Size = ""

	data, err := json.Marshal(newUpload)
	if err != nil {
		m.status.message.WriteString("Could not parse upload data to json")
		return nil, err
	}

	return data, nil
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
		request := upload{}

		if err := json.Unmarshal(body, &request); err != nil {
			m.status.message.WriteString("Could not parse get response body")
		}

		m.status.message.WriteString("Paste request at " + code + ": \n\n" + formatText(string(request.UploadBody)))
	} else {
		m.status.message.WriteString("Could not paste")
	}
	res.Body.Close()
}

func formatText(text string) string {
	formatedText := strings.Builder{}

	words := strings.Fields(text)

	for i := 0; i < len(words); i += 15 {
		lineEnd := i + 15
		if lineEnd > len(words) {
			lineEnd = len(words)
		}
		formatedText.WriteString(strings.Join(words[i:lineEnd], " "))

		formatedText.WriteString("\n")
	}

	return formatedText.String()
}

func (m clientModel) View() string {
	ui := strings.Builder{}

	ui.WriteString("\n" + m.status.message.String() + "\n")

	ui.WriteString("\n<------------------------------------->\n")

	ui.WriteString(m.textInput.View())

	return ui.String()
}
