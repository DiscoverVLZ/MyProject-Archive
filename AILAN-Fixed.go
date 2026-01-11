package main

import (
"fmt"
"os"
"path/filepath"
"sync"
"time"

"github.com/fsnotify/fsnotify"
"github.com/gdamore/tcell/v2"
"github.com/rivo/tview"
)

type Config struct {
Folders     []string
Extensions  []string
LogDir      string
IsWatching  bool
Watcher     *fsnotify.Watcher
TotalFiles  int
CurrentDir  string
mutex       sync.RWMutex
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
Folders:    []string{dir},
Extensions: []string{".txt", ".json", ".md", ".go"},
LogDir:     filepath.Join(dir, "logs"),
IsWatching: false,
TotalFiles: 0,
CurrentDir: dir,
}

os.MkdirAll(cfg.LogDir, 0755)
}

func initUI() {
app = tview.NewApplication()
tview.Styles.PrimitiveBackgroundColor = tcell.ColorBlack
tview.Styles.BorderColor = tcell.ColorWhite
tview.Styles.TitleColor = tcell.ColorYellow
tview.Styles.PrimaryTextColor = tcell.ColorWhite

createMainUI()
setupHotkeys()

// Начальное сканирование в фоне
go func() {
time.Sleep(500 * time.Millisecond)
quickScan()
}()
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
addLog("🚀 AILAN Archivist готов")

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
cfg.mutex.RLock()
status := "[red]🔴 ВЫКЛ"
if cfg.IsWatching {
status = "[green]🟢 ВКЛ"
}
totalFiles := cfg.TotalFiles
folderCount := len(cfg.Folders)
cfg.mutex.RUnlock()

text := fmt.Sprintf("[white]🚀 AILAN ARCHIVIST | Мониторинг: %s | Файлов: [yellow]%d[-] | Папок: [yellow]%d[-]", 
status, totalFiles, folderCount)
panel.SetText(text)
panel.SetBackgroundColor(tcell.ColorDarkBlue)
}

func updateLeftPane() {
leftPane.Clear()

cfg.mutex.RLock()
isWatching := cfg.IsWatching
cfg.mutex.RUnlock()

if isWatching {
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
cfg.mutex.Lock()
if len(cfg.Folders) > 1 {
removed := cfg.Folders[len(cfg.Folders)-1]
cfg.Folders = cfg.Folders[:len(cfg.Folders)-1]
cfg.mutex.Unlock()

app.QueueUpdateDraw(func() {
updateRightPane()
updateTopPanel(nil)
addLog(fmt.Sprintf("[yellow]Удалена папка: %s[-]", filepath.Base(removed)))
})
} else {
cfg.mutex.Unlock()
addLog("[red]Нельзя удалить последнюю папку[-]")
}
})

leftPane.AddItem("[cyan]🔍 Быстрое сканирование[-]", "F7", 'Q', func() {
go quickScan()
})

leftPane.AddItem("[white]📊 Статистика[-]", "F2", 'T', func() {
showStatistics()
})

leftPane.AddItem("[red]❌ Выход[-]", "F10", 'X', func() {
stopWatching()
app.Stop()
})
}

func updateRightPane() {
cfg.mutex.RLock()
folders := make([]string, len(cfg.Folders))
copy(folders, cfg.Folders)
cfg.mutex.RUnlock()

rightPane.Clear()

for _, folder := range folders {
folderName := filepath.Base(folder)
if len(folderName) > 25 {
folderName = "..." + folderName[len(folderName)-22:]
}
rightPane.AddItem(fmt.Sprintf("📁 [yellow]%s[-]", folderName), 
fmt.Sprintf("[gray]%s[-]", folder), 0, nil)
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
case tcell.KeyF2:
showStatistics()
return nil
case tcell.KeyF4:
addFolder(cfg.CurrentDir)
return nil
case tcell.KeyF5:
toggleWatching()
return nil
case tcell.KeyF7:
go quickScan()
return nil
case tcell.KeyF8:
cfg.mutex.Lock()
if len(cfg.Folders) > 1 {
removed := cfg.Folders[len(cfg.Folders)-1]
cfg.Folders = cfg.Folders[:len(cfg.Folders)-1]
cfg.mutex.Unlock()

app.QueueUpdateDraw(func() {
updateRightPane()
updateTopPanel(nil)
addLog(fmt.Sprintf("[yellow]Удалена папка: %s[-]", filepath.Base(removed)))
})
} else {
cfg.mutex.Unlock()
addLog("[red]Нельзя удалить последнюю папку[-]")
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
SetText("[yellow]AILAN ARCHIVIST[-]\n\nГорячие клавиши:\nF4 - Добавить папку\nF5 - Мониторинг\nF7 - Сканирование\nF8 - Удалить\nF10 - Выход\nTab - Панели").
AddButtons([]string{"Закрыть"}).
SetDoneFunc(func(buttonIndex int, buttonLabel string) {
app.SetRoot(mainFlex, true)
})
app.SetRoot(modal, true)
}

func addFolder(path string) {
cfg.mutex.Lock()

// Проверка на дубликат
for _, folder := range cfg.Folders {
if folder == path {
cfg.mutex.Unlock()
app.QueueUpdateDraw(func() {
addLog("[yellow]Папка уже добавлена[-]")
})
return
}
}

cfg.Folders = append(cfg.Folders, path)
cfg.mutex.Unlock()

// Сканируем в фоне
go func() {
count := safeCountFiles(path)
cfg.mutex.Lock()
cfg.TotalFiles += count
cfg.mutex.Unlock()

app.QueueUpdateDraw(func() {
updateRightPane()
updateTopPanel(nil)
addLog(fmt.Sprintf("[green]Добавлена папка: %s (%d файлов)[-]", 
filepath.Base(path), count))
})
}()
}

func showStatistics() {
cfg.mutex.RLock()
totalFiles := cfg.TotalFiles
folderCount := len(cfg.Folders)
isWatching := cfg.IsWatching
cfg.mutex.RUnlock()

status := "[red]🔴 ВЫКЛ"
if isWatching {
status = "[green]🟢 ВКЛ"
}

modal := tview.NewModal().
SetText(fmt.Sprintf("[yellow]📊 СТАТИСТИКА[-]\n\nПапок: [cyan]%d[-]\nФайлов: [cyan]%d[-]\nМониторинг: %s", 
folderCount, totalFiles, status)).
AddButtons([]string{"Закрыть"}).
SetDoneFunc(func(buttonIndex int, buttonLabel string) {
app.SetRoot(mainFlex, true)
})
app.SetRoot(modal, true)
}

func toggleWatching() {
cfg.mutex.Lock()
if cfg.IsWatching {
cfg.mutex.Unlock()
stopWatching()
} else {
cfg.mutex.Unlock()
startWatching()
}
}

func startWatching() {
cfg.mutex.Lock()
if cfg.IsWatching {
cfg.mutex.Unlock()
return
}

var err error
cfg.Watcher, err = fsnotify.NewWatcher()
if err != nil {
cfg.mutex.Unlock()
app.QueueUpdateDraw(func() {
addLog(fmt.Sprintf("[red]Ошибка мониторинга: %v[-]", err))
})
return
}

cfg.IsWatching = true
cfg.mutex.Unlock()

app.QueueUpdateDraw(func() {
addLog("[green]🚀 Мониторинг запущен[-]")
updateLeftPane()
updateTopPanel(nil)
})

// Запуск в фоне
go func() {
for {
cfg.mutex.RLock()
if !cfg.IsWatching || cfg.Watcher == nil {
cfg.mutex.RUnlock()
break
}
watcher := cfg.Watcher
cfg.mutex.RUnlock()

select {
case event, ok := <-watcher.Events:
if !ok {
return
}
processEvent(event)
case err, ok := <-watcher.Errors:
if !ok {
return
}
app.QueueUpdateDraw(func() {
addLog(fmt.Sprintf("[red]Ошибка: %v[-]", err))
})
}
}
}()
}

func stopWatching() {
cfg.mutex.Lock()
if !cfg.IsWatching {
cfg.mutex.Unlock()
return
}

if cfg.Watcher != nil {
cfg.Watcher.Close()
cfg.Watcher = nil
}

cfg.IsWatching = false
cfg.mutex.Unlock()

app.QueueUpdateDraw(func() {
addLog("[yellow]⏸ Мониторинг остановлен[-]")
updateLeftPane()
updateTopPanel(nil)
})
}

func processEvent(event fsnotify.Event) {
ext := filepath.Ext(event.Name)

cfg.mutex.RLock()
shouldTrack := false
for _, trackedExt := range cfg.Extensions {
if ext == trackedExt {
shouldTrack = true
break
}
}
cfg.mutex.RUnlock()

if !shouldTrack {
return
}

filename := filepath.Base(event.Name)
var msg string
var color string

if event.Op&fsnotify.Create == fsnotify.Create {
msg = fmt.Sprintf("Создан: %s", filename)
color = "green"
cfg.mutex.Lock()
cfg.TotalFiles++
cfg.mutex.Unlock()
} else if event.Op&fsnotify.Write == fsnotify.Write {
msg = fmt.Sprintf("Изменен: %s", filename)
color = "yellow"
} else if event.Op&fsnotify.Remove == fsnotify.Remove {
msg = fmt.Sprintf("Удален: %s", filename)
color = "red"
cfg.mutex.Lock()
if cfg.TotalFiles > 0 {
cfg.TotalFiles--
}
cfg.mutex.Unlock()
} else {
return
}

app.QueueUpdateDraw(func() {
addLog(fmt.Sprintf("[%s]%s[-]", color, msg))
updateTopPanel(nil)
})
}

func quickScan() {
app.QueueUpdateDraw(func() {
addLog("[yellow]🔍 Сканирование...[-]")
})

total := 0
cfg.mutex.RLock()
folders := make([]string, len(cfg.Folders))
copy(folders, cfg.Folders)
cfg.mutex.RUnlock()

for _, folder := range folders {
count := safeCountFiles(folder)
total += count

folderName := filepath.Base(folder)
app.QueueUpdateDraw(func() {
addLog(fmt.Sprintf("[gray]%s: %d файлов[-]", folderName, count))
})
time.Sleep(50 * time.Millisecond) // Пауза для UI
}

cfg.mutex.Lock()
cfg.TotalFiles = total
cfg.mutex.Unlock()

app.QueueUpdateDraw(func() {
updateTopPanel(nil)
addLog(fmt.Sprintf("[green]✓ Найдено: %d файлов[-]", total))
})
}

func safeCountFiles(folder string) int {
count := 0
filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
if err != nil || info.IsDir() {
return nil
}

ext := filepath.Ext(path)
cfg.mutex.RLock()
for _, trackedExt := range cfg.Extensions {
if ext == trackedExt {
count++
break
}
}
cfg.mutex.RUnlock()
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
