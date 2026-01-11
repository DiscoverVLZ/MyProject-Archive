package main

import (
"fmt"
"os"
"path/filepath"
"time"
"github.com/fsnotify/fsnotify"
"github.com/gdamore/tcell/v2"
"github.com/rivo/tview"
)

// Конфигурация
type Config struct {
Folders      []string
Extensions   []string
LogDir       string
SummaryFile  string
IsWatching   bool
Watcher      *fsnotify.Watcher
TotalFiles   int
CurrentDir   string
}

var (
cfg          Config
app          *tview.Application
leftPane     *tview.List
rightPane    *tview.List
statusBar    *tview.TextView
logView      *tview.TextView
mainFlex     *tview.Flex
currentPanel = "left"
)

func main() {
initConfig()
initUI()

if err := app.SetRoot(mainFlex, true).EnableMouse(true).Run(); err != nil {
fmt.Printf("Ошибка: %v\n", err)
os.Exit(1)
}
}

func initConfig() {
dir, _ := os.Getwd()
cfg = Config{
Folders:     []string{dir},
Extensions:  []string{".php", ".html", ".js", ".css", ".txt", ".json", ".md", ".go"},
LogDir:      filepath.Join(dir, "docs", "changelog"),
SummaryFile: filepath.Join(dir, "docs", "project_state.md"),
IsWatching:  false,
TotalFiles:  0,
CurrentDir:  dir,
}

os.MkdirAll(cfg.LogDir, 0755)
os.MkdirAll(filepath.Dir(cfg.SummaryFile), 0755)
}

func initUI() {
app = tview.NewApplication()
tview.Styles.PrimitiveBackgroundColor = tcell.ColorBlack
tview.Styles.BorderColor = tcell.ColorWhite
tview.Styles.TitleColor = tcell.ColorYellow
tview.Styles.PrimaryTextColor = tcell.ColorWhite

createMainUI()
setupHotkeys()
}

func createMainUI() {
// Верхняя панель
topPanel := tview.NewTextView().
SetDynamicColors(true).
SetTextAlign(tview.AlignCenter)
updateTopPanel(topPanel)

// Левая панель
leftPane = tview.NewList().
ShowSecondaryText(true).
SetHighlightFullLine(true)
leftPane.SetBorder(true).
SetTitle(" [yellow]КОМАНДЫ[-] ").
SetBackgroundColor(tcell.ColorBlack)
updateLeftPane()

// Правая панель
rightPane = tview.NewList().
ShowSecondaryText(true).
SetHighlightFullLine(true)
rightPane.SetBorder(true).
SetTitle(" [yellow]МОНИТОРИНГ[-] ").
SetBackgroundColor(tcell.ColorBlack)
updateRightPane()

// Логи
logView = tview.NewTextView().
SetDynamicColors(true).
SetScrollable(true)
logView.SetBorder(true).
SetTitle(" [yellow]ЖУРНАЛ[-] ").
SetBackgroundColor(tcell.ColorBlack)
addLog("Готов к работе")

// Статус бар
statusBar = tview.NewTextView().
SetDynamicColors(true)
updateStatusBar()

// Основной layout
mainFlex = tview.NewFlex().SetDirection(tview.FlexRow)
mainFlex.AddItem(topPanel, 1, 0, false)

panels := tview.NewFlex().SetDirection(tview.FlexColumn)
panels.AddItem(leftPane, 0, 1, true)
panels.AddItem(rightPane, 0, 1, false)

mainFlex.AddItem(panels, 0, 3, true)
mainFlex.AddItem(logView, 10, 1, false)
mainFlex.AddItem(statusBar, 1, 0, false)
}

func updateTopPanel(panel *tview.TextView) {
status := "[red]ВЫКЛ"
if cfg.IsWatching {
status = "[green]ВКЛ"
}

text := fmt.Sprintf("AILAN ARCHIVIST | Мониторинг: %s | Файлов: %d | Папок: %d", 
status, cfg.TotalFiles, len(cfg.Folders))
panel.SetText(text)
panel.SetBackgroundColor(tcell.ColorDarkBlue)
}

func updateLeftPane() {
leftPane.Clear()

if cfg.IsWatching {
leftPane.AddItem("[green]⏸ Остановить мониторинг[-]", "F5", 'S', func() {
stopWatching()
})
} else {
leftPane.AddItem("[green]▶ Запустить мониторинг[-]", "F5", 'S', func() {
startWatching()
})
}

leftPane.AddItem("[cyan]📁 Добавить папку[-]", "F4", 'A', func() {
addFolder(cfg.CurrentDir)
})

leftPane.AddItem("[cyan]🗑 Удалить папку[-]", "F8", 'D', func() {
if len(cfg.Folders) > 0 {
cfg.Folders = cfg.Folders[:len(cfg.Folders)-1]
updateRightPane()
updateTopPanel(nil)
addLog("[yellow]Папка удалена[-]")
}
})

leftPane.AddItem("[cyan]🔍 Быстрое сканирование[-]", "F7", 'Q', func() {
quickScan()
})

leftPane.AddItem("[red]❌ Выход[-]", "F10", 'X', func() {
stopWatching()
app.Stop()
})
}

func updateRightPane() {
rightPane.Clear()

for _, folder := range cfg.Folders {
folderName := filepath.Base(folder)
if len(folderName) > 30 {
folderName = "..." + folderName[len(folderName)-27:]
}
rightPane.AddItem(fmt.Sprintf("📁 %s", folderName), folder, 0, nil)
}
}

func updateStatusBar() {
helpText := "[F1]Помощь [F4]Добавить [F5]Мониторинг [F7]Сканировать [F8]Удалить [F10]Выход"
statusBar.SetText(fmt.Sprintf("[white]%s | %s[-]", time.Now().Format("15:04:05"), helpText))
statusBar.SetBackgroundColor(tcell.ColorDarkBlue)
}

func setupHotkeys() {
app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
switch event.Key() {
case tcell.KeyF1:
showHelp()
return nil
case tcell.KeyF4:
addFolder(cfg.CurrentDir)
return nil
case tcell.KeyF5:
toggleWatching()
return nil
case tcell.KeyF7:
quickScan()
return nil
case tcell.KeyF8:
if len(cfg.Folders) > 0 {
cfg.Folders = cfg.Folders[:len(cfg.Folders)-1]
updateRightPane()
updateTopPanel(nil)
addLog("[yellow]Папка удалена[-]")
}
return nil
case tcell.KeyF10:
stopWatching()
app.Stop()
return nil
case tcell.KeyTab:
togglePanel()
return nil
}
return event
})
}

func togglePanel() {
if currentPanel == "left" {
currentPanel = "right"
app.SetFocus(rightPane)
} else {
currentPanel = "left"
app.SetFocus(leftPane)
}
}

func showHelp() {
modal := tview.NewModal().
SetText("AILAN ARCHIVIST\n\nГорячие клавиши:\nF4 - Добавить папку\nF5 - Мониторинг\nF7 - Сканирование\nF8 - Удалить\nF10 - Выход\nTab - Переключение панелей").
AddButtons([]string{"Закрыть"}).
SetDoneFunc(func(buttonIndex int, buttonLabel string) {
app.SetRoot(mainFlex, true)
})
app.SetRoot(modal, true)
}

func addFolder(path string) {
cfg.Folders = append(cfg.Folders, path)
count := countFilesInFolder(path)
cfg.TotalFiles += count
addLog(fmt.Sprintf("[green]Добавлена папка: %s (%d файлов)[-]", path, count))
updateRightPane()
updateTopPanel(nil)
}

func toggleWatching() {
if cfg.IsWatching {
stopWatching()
} else {
startWatching()
}
}

func startWatching() {
if cfg.IsWatching {
return
}

var err error
cfg.Watcher, err = fsnotify.NewWatcher()
if err != nil {
addLog(fmt.Sprintf("[red]Ошибка: %v[-]", err))
return
}

cfg.IsWatching = true
addLog("[green]Мониторинг запущен[-]")

// Запуск мониторинга в отдельном потоке
go func() {
for cfg.IsWatching {
select {
case event := <-cfg.Watcher.Events:
processEvent(event)
case err := <-cfg.Watcher.Errors:
if err != nil {
addLog(fmt.Sprintf("[red]Ошибка: %v[-]", err))
}
}
}
}()

updateLeftPane()
updateTopPanel(nil)
}

func stopWatching() {
if !cfg.IsWatching {
return
}

if cfg.Watcher != nil {
cfg.Watcher.Close()
}

cfg.IsWatching = false
addLog("[yellow]Мониторинг остановлен[-]")
updateLeftPane()
updateTopPanel(nil)
}

func processEvent(event fsnotify.Event) {
ext := filepath.Ext(event.Name)
shouldTrack := false
for _, trackedExt := range cfg.Extensions {
if ext == trackedExt {
shouldTrack = true
break
}
}

if !shouldTrack {
return
}

eventType := "изменен"
if event.Op&fsnotify.Create == fsnotify.Create {
eventType = "создан"
cfg.TotalFiles++
} else if event.Op&fsnotify.Remove == fsnotify.Remove {
eventType = "удален"
if cfg.TotalFiles > 0 {
cfg.TotalFiles--
}
}

filename := filepath.Base(event.Name)
msg := fmt.Sprintf("[cyan]Файл %s %s[-]", filename, eventType)

app.QueueUpdateDraw(func() {
addLog(msg)
updateTopPanel(nil)
})
}

func quickScan() {
go func() {
addLog("[yellow]Сканирование...[-]")
time.Sleep(2 * time.Second)
total := 0
for _, folder := range cfg.Folders {
total += countFilesInFolder(folder)
}
cfg.TotalFiles = total
app.QueueUpdateDraw(func() {
updateTopPanel(nil)
addLog(fmt.Sprintf("[green]Найдено: %d файлов[-]", total))
})
}()
}

func countFilesInFolder(folder string) int {
count := 0
filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
if err != nil || info.IsDir() {
return nil
}

ext := filepath.Ext(path)
for _, trackedExt := range cfg.Extensions {
if ext == trackedExt {
count++
break
}
}
return nil
})
return count
}

func addLog(message string) {
timestamp := time.Now().Format("15:04:05")
currentText := logView.GetText(true)
logView.SetText(currentText + fmt.Sprintf("[gray]%s[-] %s\n", timestamp, message))
logView.ScrollToEnd()
}
