package main

import (
"fmt"
"os"


"github.com/gdamore/tcell/v2"
"github.com/rivo/tview"
)

var (
app      *tview.Application
leftPane *tview.List
rightPane *tview.List
)

func main() {
app = tview.NewApplication()

// Левая панель - команды
leftPane = tview.NewList().
ShowSecondaryText(true).
SetHighlightFullLine(true)

leftPane.SetBorder(true).
SetTitle(" [yellow]🖥 КОМАНДЫ[-] ").
SetTitleAlign(tview.AlignLeft)

leftPane.AddItem("[green]▶ Запустить мониторинг[-]", "Нажмите Enter или F5", 'S', func() {
showMessage("Мониторинг запущен")
})

leftPane.AddItem("[cyan]📁 Добавить папку[-]", "Нажмите Enter или F4", 'A', func() {
showAddFolderDialog()
})

leftPane.AddItem("[red]❌ Выход[-]", "Нажмите Enter или F10", 'X', func() {
app.Stop()
})

// Правая панель - мониторинг
rightPane = tview.NewList().
ShowSecondaryText(true).
SetHighlightFullLine(true)

rightPane.SetBorder(true).
SetTitle(" [yellow]📁 МОНИТОРИНГ[-] ").
SetTitleAlign(tview.AlignLeft)

rightPane.AddItem("📁 [yellow]Текущая папка[-]", "[gray]Нажмите Enter для управления[-]", 0, func() {
showMessage("Управление папкой")
})

// Статус бар
statusBar := tview.NewTextView().
SetDynamicColors(true)

statusBar.SetText("[white]F1 Справка | F4 Добавить | F5 Мониторинг | F10 Выход[-]")

// Основной layout
flex := tview.NewFlex().SetDirection(tview.FlexRow)

contentFlex := tview.NewFlex().SetDirection(tview.FlexColumn)
contentFlex.AddItem(leftPane, 0, 1, true)
contentFlex.AddItem(rightPane, 0, 1, false)

flex.AddItem(contentFlex, 0, 3, true)
flex.AddItem(statusBar, 1, 0, false)

// Настройка горячих клавиш
app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
switch event.Key() {
case tcell.KeyF1:
showHelp()
return nil
case tcell.KeyF4:
showAddFolderDialog()
return nil
case tcell.KeyF5:
showMessage("Мониторинг вкл/выкл")
return nil
case tcell.KeyF10:
app.Stop()
return nil
}
return event
})

if err := app.SetRoot(flex, true).EnableMouse(true).Run(); err != nil {
fmt.Printf("Ошибка: %v\n", err)
os.Exit(1)
}
}

func showAddFolderDialog() {
modal := tview.NewModal().
SetText("Добавление папки для мониторинга").
AddButtons([]string{"Добавить", "Отмена"}).
SetDoneFunc(func(buttonIndex int, buttonLabel string) {
if buttonLabel == "Добавить" {
showMessage("Папка добавлена")
}
app.SetRoot(leftPane, true)
})

app.SetRoot(modal, true)
}

func showHelp() {
modal := tview.NewModal().
SetText("AILAN Archivist - Total Commander Style\n\nГорячие клавиши:\nF1 - Справка\nF4 - Добавить папку\nF5 - Мониторинг\nF10 - Выход").
AddButtons([]string{"Закрыть"}).
SetDoneFunc(func(buttonIndex int, buttonLabel string) {
app.SetRoot(leftPane, true)
})

app.SetRoot(modal, true)
}

func showMessage(msg string) {
modal := tview.NewModal().
SetText(msg).
AddButtons([]string{"OK"}).
SetDoneFunc(func(buttonIndex int, buttonLabel string) {
app.SetRoot(leftPane, true)
})

app.SetRoot(modal, true)
}
