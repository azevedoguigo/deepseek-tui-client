package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type DeepSeekMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type DeepSeekRequest struct {
	Model    string            `json:"model"`
	Messages []DeepSeekMessage `json:"messages"`
	Stream   bool              `json:"stream"`
}

type DeepSeekResponse struct {
	Choices []struct {
		Message DeepSeekMessage `json:"message"`
	} `json:"choices"`
}

var (
	apiKey  = os.Getenv("DEEPSEEK_API_KEY")
	history []DeepSeekMessage
	loading bool
)

func sendMessageToDeepSeek(message string, chatView *tview.TextView, app *tview.Application) {
	if apiKey == "" {
		fmt.Fprintln(chatView, "\n[red]API key não configurada.")
		return
	}

	loading = true
	app.QueueUpdateDraw(func() {
		fmt.Fprintf(chatView, "\n[blue]DeepSeek está pensando...")
		chatView.ScrollToEnd()
	})

	history = append(history, DeepSeekMessage{Role: "user", Content: message})

	client := &http.Client{Timeout: 30 * time.Second}
	reqBody := DeepSeekRequest{
		Model:    "deepseek-chat",
		Messages: history,
		Stream:   false,
	}

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(
		"POST",
		"https://api.deepseek.com/v1/chat/completions",
		bytes.NewBuffer(jsonBody),
	)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	go func() {
		resp, err := client.Do(req)
		loading = false

		app.QueueUpdateDraw(func() {
			if err != nil {
				fmt.Fprintf(chatView, "\n[red]Erro na API: %v", err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				fmt.Fprintf(chatView, "\n[red]Erro HTTP %d: %s", resp.StatusCode, string(body))
				return
			}

			var apiResp DeepSeekResponse
			if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
				fmt.Fprintf(chatView, "\n[red]Erro ao decodificar a resposta: %v", err)
				return
			}

			if len(apiResp.Choices) > 0 {
				response := apiResp.Choices[0].Message.Content
				fmt.Fprintf(chatView, "\n[white]🤖 %s", response)
				history = append(history, DeepSeekMessage{Role: "assistant", Content: response})
			}
		})
	}()
}

func main() {
	app := tview.NewApplication()

	chats := []string{"chat1", "chat2", "chat3"}

	chatList := tview.NewList()
	for _, chat := range chats {
		chatList.AddItem(chat, "", ' ', nil)
	}
	chatList.SetBorder(true).SetTitle("Chats")

	chatView := tview.NewTextView().SetDynamicColors(true)
	chatView.SetBorder(true).SetTitle("Chat")
	chatView.SetText("[green]Bem-vindo ao DeepSeek TUI!\nSelecione um chat ou comece uma nova conversa.")

	inputField := tview.NewInputField().SetLabel("Mensagem: ")
	inputField.SetFieldBackgroundColor(tview.Styles.PrimitiveBackgroundColor)

	rightPanel := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(chatView, 0, 1, false).
		AddItem(inputField, 1, 1, true)

	grid := tview.NewGrid().
		SetColumns(30, 0).
		SetRows(0, 3).
		AddItem(chatList, 0, 0, 2, 1, 0, 0, false).
		AddItem(rightPanel, 0, 1, 1, 1, 0, 0, true)

	inputField.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter && !loading {
			msg := inputField.GetText()
			if msg != "" {
				chatView.ScrollToEnd()
				fmt.Fprintf(chatView, "\n[white]Você: %s", msg)
				inputField.SetText("")

				go sendMessageToDeepSeek(msg, chatView, app)
			}
		}
	})

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTAB:
			if app.GetFocus() == chatList {
				app.SetFocus(inputField)
			} else {
				app.SetFocus(chatList)
			}

			return nil
		}
		return event
	})

	if err := app.SetRoot(grid, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
