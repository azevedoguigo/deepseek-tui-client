package main

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

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
		if key == tcell.KeyEnter {
			msg := inputField.GetText()
			if msg != "" {
				chatView.ScrollToEnd()
				fmt.Fprintf(chatView, "\n[white]Você: %s", msg)
				inputField.SetText("")
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
