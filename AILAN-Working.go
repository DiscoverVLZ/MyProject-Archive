package main

import (
"fmt"
"os"
"path/filepath"

"github.com/gdamore/tcell/v2"
"github.com/rivo/tview"
)

var (
app          *tview.Application
pages        *tview.Pages
leftPane     *tview.List
rightPane    *tview.List
statusBar    *tview.TextView
logView      *tview.TextView
folders      []string
currentPanel string = "left"
)

func main() {
app = tview.NewApplication()
pages = tview.NewPages()

// Инициализация
folders = []string{getCurrentDir()}

// Настройка стилей
tview.Styles.PrimitiveBackgroundColor = tcell.ColorBlack
tview.Styles.BorderColor = tcell.ColorWhite
tview.Styles.TitleColor = tcell.ColorYellow
tview.Styles.PrimaryTextColor = tcell.ColorWhite

createUI()

if err := app.SetRoot(pages, true).EnableMouse(true).Run(); err != nil {
fmt.Printf("Ошибка: %v\n", err)
os.Exit(1)
}
}

func getCurrentDir() string {
dir, err := os.Getwd()
if err != nil {
return "."
}
return dir
}

func createUI() {
// Верхняя панель
topPanel := tview.NewTextView().
SetDynamicColors(true).
SetTextAlign(tview.AlignCenter)

topPanel.SetText("[white]🚀 AILAN ARCHIVIST - Total Commander Style[-]")
topPanel.SetBackgroundColor(tcell.ColorDarkBlue)

// Левая панель (Команды)
leftPane = tview.NewList().
ShowSecondaryText(true).
SetHighlightFullLine(true)

leftPane.SetBorder(true).
SetTitle(" [yellow]🖥 КОМАНДЫ[-] ").
SetBackgroundColor(tcell.ColorBlack)

updateLeftPane()

// Правая панель (Папки)
rightPane = tview.NewList().
ShowSecondaryText(true).
SetHighlightFullLine(true)

rightPane.SetBorder(true).
SetTitle(" [yellow]📁 ПАПКИ[-] ").
SetBackgroundColor(tcell.ColorBlack)

updateRightPane()

// Лог
logView = tview.NewTextView().
SetDynamicColors(true).
SetScrollable(true)

logView.SetBorder(true).
SetTitle(" [yellow]📝 ЛОГ[-] ").
SetBackgroundColor(tcell.ColorBlack)

logView.SetText("[gray]Готов к работе...[-]")

// Статус бар
statusBar = tview.NewTextView().
SetDynamicColors(true)

updateStatusBar()

// Основной layout
mainFlex := tview.NewFlex().SetDirection(tview.FlexRow)

// Верхняя панель
mainFlex.AddItem(topPanel, 1, 0, false)

// Панели
panelsFlex := tview.NewFlex().SetDirection(tview.FlexColumn)
panelsFlex.AddItem(leftPane, 0, 1, true)
panelsFlex.AddItem(rightPane, 0, 1, false)

mainFlex.AddItem(panelsFlex, 0, 3, true)
mainFlex.AddItem(logView, 8, 1, false)
mainFlex.AddItem(statusBar, 1, 0, false)

pages.AddPage("main", mainFlex, true, true)

// Горячие клавиши
setupHotkeys()
}

func updateLeftPane() {
leftPane.Clear()

leftPane.AddItem("[green]▶ Запустить мониторинг[-]", "F5", 'S', func() {
addLog("[green]Мониторинг запущен[-]")
})

leftPane.AddItem("[cyan]📁 Добавить папку[-]", "F4", 'A', func() {
showAddFolder()
})

leftPane.AddItem("[cyan]🗑 Удалить папку[-]", "F8", 'D', func() {
deleteSelectedFolder()
})

leftPane.AddItem("[cyan]🔍 Быстрое сканирование[-]", "F7", 'Q', func() {
addLog("[yellow]Сканирование...[-]")
})

leftPane.AddItem("", "", 0, nil)

leftPane.AddItem("[white]📊 Статистика[-]", "F2", 'T', func() {
showStats()
})

leftPane.AddItem("[white]📝 Логи[-]", "F3", 'L', func() {
addLog("[cyan]Просмотр логов[-]")
})

leftPane.AddItem("", "", 0, nil)

leftPane.AddItem("[red]❌ Выход[-]", "F10", 'X', func() {
app.Stop()
})
}

func updateRightPane() {
rightPane.Clear()

if len(folders) == 0 {
rightPane.AddItem("[gray]Нет папок[-]", "Нажмите F4", 0, func() {
showAddFolder()
})
} else {
for i, folder := range folders {
folderName := folder
if len(folderName) > 30 {
folderName = "..." + folderName[len(folderName)-27:]
}

idx := i
rightPane.AddItem(fmt.Sprintf("📁 [yellow]%s[-]", folderName), 
"[gray]Нажмите Enter[-]", 0, func() {
showFolderMenu(idx)
})
}
}
}

func updateStatusBar() {
helpText := "[F1]Помощь [F4]Добавить [F5]Мониторинг [F7]Сканировать [F8]Удалить [F10]Выход"
statusBar.SetText(fmt.Sprintf("[white]%s[-]", helpText))
}

func setupHotkeys() {
app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
switch event.Key() {
case tcell.KeyF1:
showHelp()
return nil
case tcell.KeyF2:
showStats()
return nil
case tcell.KeyF3:
addLog("[cyan]Просмотр логов[-]")
return nil
case tcell.KeyF4:
showAddFolder()
return nil
case tcell.KeyF5:
addLog("[green]Мониторинг запущен[-]")
return nil
case tcell.KeyF7:
addLog("[yellow]Сканирование...[-]")
return nil
case tcell.KeyF8:
deleteSelectedFolder()
return nil
case tcell.KeyF10:
app.Stop()
return nil
case tcell.KeyTab:
togglePanel()
return nil
case tcell.KeyEnter:
executeSelected()
return nil
}
return event
})
}

func togglePanel() {
if currentPanel == "left" {
currentPanel = "right"
app.SetFocus(rightPane)
rightPane.SetTitle(" [yellow]📁 ПАПКИ[-] [green]◄[-] ")
leftPane.SetTitle(" [yellow]🖥 КОМАНДЫ[-] ")
} else {
currentPanel = "left"
app.SetFocus(leftPane)
leftPane.SetTitle(" [yellow]🖥 КОМАНДЫ[-] [green]◄[-] ")
rightPane.SetTitle(" [yellow]📁 ПАПКИ[-] ")
}
updateStatusBar()
}

func executeSelected() {
if currentPanel == "left" {
idx := leftPane.GetCurrentItem()
if idx >= 0 {
leftPane.SetCurrentItem(idx)
}
} else {
idx := rightPane.GetCurrentItem()
if idx >= 0 {
rightPane.SetCurrentItem(idx)
}
}
}

func showAddFolder() {
modal := tview.NewModal().
SetText("Добавить папку для мониторинга\n\nВведите путь:").
AddButtons([]string{"Добавить", "Отмена"}).
SetDoneFunc(func(buttonIndex int, buttonLabel string) {
if buttonLabel == "Добавить" {
// В реальности здесь был бы ввод пути
newFolder := getCurrentDir()
folders = append(folders, newFolder)
updateRightPane()
addLog(fmt.Sprintf("[green]Добавлена папка: %s[-]", newFolder))
}
pages.ShowPage("main")
})

modal.SetBorder(true).
SetTitle(" [yellow]📁 ДОБАВИТЬ ПАПКУ[-] ")

pages.AddPage("addFolder", modal, true, true)
pages.ShowPage("addFolder")
}

func deleteSelectedFolder() {
if len(folders) == 0 {
addLog("[red]Нет папок для удаления[-]")
return
}

modal := tview.NewModal().
SetText("Удалить выбранную папку?").
AddButtons([]string{"Да", "Нет"}).
SetDoneFunc(func(buttonIndex int, buttonLabel string) {
if buttonLabel == "Да" && len(folders) > 0 {
removed := folders[0]
folders = folders[1:]
updateRightPane()
addLog(fmt.Sprintf("[yellow]Удалена папка: %s[-]", removed))
}
pages.ShowPage("main")
})

modal.SetBorder(true).
SetTitle(" [yellow]🗑 УДАЛЕНИЕ[-] ")

pages.AddPage("deleteFolder", modal, true, true)
pages.ShowPage("deleteFolder")
}

func showFolderMenu(index int) {
if index < len(folders) {
folder := folders[index]

modal := tview.NewModal().
SetText(fmt.Sprintf("Папка: [yellow]%s[-]\n\nВыберите действие:", folder)).
AddButtons([]string{"Сканировать", "Сделать основной", "Удалить", "Отмена"}).
SetDoneFunc(func(buttonIndex int, buttonLabel string) {
switch buttonLabel {
case "Сканировать":
addLog(fmt.Sprintf("[cyan]Сканирование %s[-]", folder))
case "Сделать основной":
addLog(fmt.Sprintf("[green]Основная папка: %s[-]", folder))
case "Удалить":
if index < len(folders) {
removed := folders[index]
folders = append(folders[:index], folders[index+1:]...)
updateRightPane()
addLog(fmt.Sprintf("[yellow]Удалена: %s[-]", removed))
}
}
pages.ShowPage("main")
})

modal.SetBorder(true).
SetTitle(" [yellow]📂 УПРАВЛЕНИЕ[-] ")

pages.AddPage("folderMenu", modal, true, true)
pages.ShowPage("folderMenu")
}
}

func showHelp() {
modal := tview.NewModal().
SetText("[yellow]🚀 AILAN ARCHIVIST[-]\n\nГорячие клавиши:\nF1 - Справка\nF4 - Добавить папку\nF5 - Мониторинг\nF7 - Сканирование\nF8 - Удалить\nF10 - Выход\n\nTab - Переключение панелей\nEnter - Выполнить команду").
AddButtons([]string{"Закрыть"}).
SetDoneFunc(func(buttonIndex int, buttonLabel string) {
pages.ShowPage("main")
})

modal.SetBorder(true).
SetTitle(" [yellow]❓ СПРАВКА[-] ")

pages.AddPage("help", modal, true, true)
pages.ShowPage("help")
}

func showStats() {
modal := tview.NewModal().
SetText(fmt.Sprintf("[yellow]📊 СТАТИСТИКА[-]\n\nПапок: [cyan]%d[-]\n\nГотов к работе!", len(folders))).
AddButtons([]string{"Закрыть", "Экспорт"}).
SetDoneFunc(func(buttonIndex int, buttonLabel string) {
if buttonLabel == "Экспорт" {
addLog("[green]Статистика экспортирована[-]")
}
pages.ShowPage("main")
})

modal.SetBorder(true).
SetTitle(" [yellow]📈 СТАТИСТИКА[-] ")

pages.AddPage("stats", modal, true, true)
pages.ShowPage("stats")
}

func addLog(message string) {
currentText := logView.GetText(true)
logView.SetText(currentText + message + "\n")
logView.ScrollToEnd()
}
