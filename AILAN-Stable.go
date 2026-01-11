package main

import (
"fmt"
"os"
"path/filepath"
"time"
"github.com/gdamore/tcell/v2"
"github.com/rivo/tview"
)

type Config struct {
Folders     []string
Extensions  []string
IsWatching  bool
TotalFiles  int
CurrentDir  string
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
stopChan     = make(chan bool, 1)
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
Folders:    []string{dir},
Extensions: []string{".txt", ".json", ".md"},
IsWatching: false,
TotalFiles: 0,
CurrentDir: dir,
}
}

func initUI() {
app = tview.NewApplication()

// Простые стили
tview.Styles.PrimitiveBackgroundColor = tcell.ColorBlack
tview.Styles.BorderColor = tcell.ColorWhite
tview.Styles.TitleColor = tcell.ColorYellow

createMainUI()
setupHotkeys()
}

func createMainUI() {
// Верхняя панель
topPanel := tview.NewTextView().
SetDynamicColors(true).
SetTextAlign(tview.AlignCenter)
updateTopPanel(topPanel)

// Левая панель - команды
leftPane = tview.NewList().
ShowSecondaryText(false).
SetHighlightFullLine(true)
leftPane.SetBorder(true).
SetTitle(" КОМАНДЫ ").
SetTitleColor(tcell.ColorYellow)

updateLeftPane()

// Правая панель - папки
rightPane = tview.NewList().
ShowSecondaryText(false).
SetHighlightFullLine(true)
rightPane.SetBorder(true).
SetTitle(" ПАПКИ ").
SetTitleColor(tcell.ColorYellow)

updateRightPane()

// Лог
logView = tview.NewTextView().
SetDynamicColors(true)
logView.SetBorder(true).
SetTitle(" ЛОГ ").
SetTitleColor(tcell.ColorYellow)

addLog("Программа запущена")

// Статус бар
statusBar = tview.NewTextView().
SetDynamicColors(true)
updateStatusBar()

// Layout
mainFlex = tview.NewFlex().SetDirection(tview.FlexRow)
mainFlex.AddItem(topPanel, 1, 0, false)

panels := tview.NewFlex().SetDirection(tview.FlexColumn)
panels.AddItem(leftPane, 0, 1, true)
panels.AddItem(rightPane, 0, 1, false)

mainFlex.AddItem(panels, 0, 3, true)
mainFlex.AddItem(logView, 8, 1, false)
mainFlex.AddItem(statusBar, 1, 0, false)
}

func updateTopPanel(panel *tview.TextView) {
status := "🔴 ВЫКЛ"
if cfg.IsWatching {
status = "🟢 ВКЛ"
}

text := fmt.Sprintf("🚀 AILAN ARCHIVIST | Мониторинг: %s | Файлов: %d", 
status, cfg.TotalFiles)
panel.SetText(text)
}

func updateLeftPane() {
leftPane.Clear()

// Всегда добавляем 4 основные команды
leftPane.AddItem("▶ Запустить мониторинг", "F5", '5', func() {
if !cfg.IsWatching {
cfg.IsWatching = true
addLog("Мониторинг запущен")
go startMonitoring()
updateLeftPane()
updateTopPanel(nil)
}
})

leftPane.AddItem("📁 Добавить папку", "F4", '4', func() {
dir, _ := os.Getwd()
cfg.Folders = append(cfg.Folders, dir)
addLog("Папка добавлена: " + filepath.Base(dir))
updateRightPane()
})

leftPane.AddItem("🗑 Удалить папку", "F8", '8', func() {
if len(cfg.Folders) > 1 {
cfg.Folders = cfg.Folders[:len(cfg.Folders)-1]
addLog("Папка удалена")
updateRightPane()
}
})

leftPane.AddItem("❌ Выход", "F10", '0', func() {
cfg.IsWatching = false
stopChan <- true
app.Stop()
})
}

func updateRightPane() {
rightPane.Clear()

for _, folder := range cfg.Folders {
name := filepath.Base(folder)
if len(name) > 20 {
name = name[:17] + "..."
}
rightPane.AddItem("📁 " + name, "", 0, nil)
}
}

func updateStatusBar() {
timeStr := time.Now().Format("15:04")
statusBar.SetText(fmt.Sprintf(" %s | F4:Добавить F5:Мониторинг F8:Удалить F10:Выход", timeStr))
}

func setupHotkeys() {
app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
switch event.Key() {
case tcell.KeyF4:
dir, _ := os.Getwd()
cfg.Folders = append(cfg.Folders, dir)
addLog("Папка добавлена: " + filepath.Base(dir))
updateRightPane()
return nil
case tcell.KeyF5:
if !cfg.IsWatching {
cfg.IsWatching = true
addLog("Мониторинг запущен")
go startMonitoring()
updateLeftPane()
updateTopPanel(nil)
}
return nil
case tcell.KeyF8:
if len(cfg.Folders) > 1 {
cfg.Folders = cfg.Folders[:len(cfg.Folders)-1]
addLog("Папка удалена")
updateRightPane()
}
return nil
case tcell.KeyF10:
cfg.IsWatching = false
stopChan <- true
app.Stop()
return nil
case tcell.KeyTab:
if currentPanel == "left" {
currentPanel = "right"
app.SetFocus(rightPane)
} else {
currentPanel = "left"
app.SetFocus(leftPane)
}
return nil
}
return event
})
}

func startMonitoring() {
addLog("Фоновый мониторинг начат")

ticker := time.NewTicker(5 * time.Second)
defer ticker.Stop()

for {
select {
case <-stopChan:
addLog("Мониторинг остановлен")
return
case <-ticker.C:
if cfg.IsWatching {
// Имитация обнаружения файлов
app.QueueUpdateDraw(func() {
cfg.TotalFiles += 1
addLog(fmt.Sprintf("Обнаружен новый файл (всего: %d)", cfg.TotalFiles))
updateTopPanel(nil)
})
}
}
}
}

func addLog(message string) {
timeStr := time.Now().Format("15:04:05")
logView.SetText(logView.GetText(false) + fmt.Sprintf("[%s] %s\n", timeStr, message))
logView.ScrollToEnd()
}
