package main

import (
"fmt"
"io/fs"
"os"
"path/filepath"
"sort"
"strings"
"time"

"github.com/gdamore/tcell/v2"
"github.com/rivo/tview"
)

// === КОНФИГУРАЦИЯ ===
type Config struct {
WatchDir      string
Extensions    []string
LogDir        string
SummaryFile   string
IsRunning     bool
CustomFolders []string
}

var (
cfg      Config
app      *tview.Application
pages    *tview.Pages
leftPane *tview.List
rightPane *tview.List
statusBar *tview.TextView
logView   *tview.TextView
mainFlex  *tview.Flex
currentPanel string // "left" или "right"
)

// === ОСНОВНАЯ ФУНКЦИЯ ===
func main() {
initConfig()
initUI()

// Запускаем приложение
if err := app.SetRoot(pages, true).EnableMouse(true).Run(); err != nil {
panic(err)
}
}

func initConfig() {
cfg = Config{
WatchDir:      ".",
Extensions:    []string{".php", ".html", ".js", ".css", ".txt", ".json"},
LogDir:        "docs/changelog",
SummaryFile:   "docs/project_state.md",
IsRunning:     false,
CustomFolders: []string{"."},
}

os.MkdirAll(cfg.LogDir, 0755)
os.MkdirAll(filepath.Dir(cfg.SummaryFile), 0755)

currentPanel = "left"
}

func initUI() {
app = tview.NewApplication()
pages = tview.NewPages()

// Создаем главный интерфейс как в Total Commander
createMainUI()

// Добавляем главную страницу
pages.AddPage("main", mainFlex, true, true)

// Устанавливаем горячие клавиши
setupHotkeys()
}

func createMainUI() {
// === ВЕРХНЯЯ ПАНЕЛЬ ===
topPanel := tview.NewTextView().
SetDynamicColors(true).
SetRegions(true).
SetTextAlign(tview.AlignCenter)

topPanel.SetBorder(true).
SetTitle(" 🚀 AILAN ARCHIVIST - TOTAL COMMANDER STYLE ")

updateTopPanel(topPanel)

// === ЛЕВАЯ ПАНЕЛЬ (Файлы/Папки) ===
leftPane = tview.NewList().
ShowSecondaryText(false).
SetHighlightFullLine(true)

leftPane.SetBorder(true).
SetTitle(" [::b]📁 ЛЕВАЯ ПАНЕЛЬ[::-] ").
SetTitleAlign(tview.AlignLeft)

updateLeftPane()

// === ПРАВАЯ ПАНЕЛЬ (Логи/Статус) ===
rightPane = tview.NewList().
ShowSecondaryText(false).
SetHighlightFullLine(true)

rightPane.SetBorder(true).
SetTitle(" [::b]📊 ПРАВАЯ ПАНЕЛЬ[::-] ").
SetTitleAlign(tview.AlignLeft)

updateRightPane()

// === ПАНЕЛЬ СТАТУСА ===
statusBar = tview.NewTextView().
SetDynamicColors(true).
SetRegions(true)

statusBar.SetBorder(false)
updateStatusBar()

// === ПРОСМОТР ЛОГОВ ===
logView = tview.NewTextView().
SetDynamicColors(true).
SetScrollable(true).
SetChangedFunc(func() {
app.Draw()
})

logView.SetBorder(true).
SetTitle(" [::b]📝 ЖУРНАЛ СОБЫТИЙ[::-] ")

// === ОСНОВНОЙ LAYOUT ===
mainFlex = tview.NewFlex().SetDirection(tview.FlexRow)

// Верхняя панель (10%)
mainFlex.AddItem(topPanel, 3, 1, false)

// Основная область (80%)
contentFlex := tview.NewFlex().SetDirection(tview.FlexColumn)
contentFlex.AddItem(leftPane, 0, 1, true)   // Левая панель 50%
contentFlex.AddItem(rightPane, 0, 1, false) // Правая панель 50%

mainFlex.AddItem(contentFlex, 0, 4, true)

// Область логов (20%)
mainFlex.AddItem(logView, 10, 1, false)

// Панель статуса (5%)
mainFlex.AddItem(statusBar, 1, 1, false)
}

func updateTopPanel(panel *tview.TextView) {
status := "[red]🔴 ВЫКЛ"
if cfg.IsRunning {
status = "[green]🟢 ВКЛ"
}

text := fmt.Sprintf(`[white]Мониторинг: %s | Папок: [yellow]%d[-] | Файлов: [yellow]%d[-] | Расширений: [yellow]%d[-]`, 
status, len(cfg.CustomFolders), countAllTrackedFiles(), len(cfg.Extensions))

panel.SetText(text)
}

func updateLeftPane() {
leftPane.Clear()

// Заголовок
leftPane.AddItem("📁 УПРАВЛЕНИЕ ПАПКАМИ", "", 0, func() {
showFolderManager()
})

leftPane.AddItem("⚙ УПРАВЛЕНИЕ РАСШИРЕНИЯМИ", "", 0, func() {
showExtensionManager()
})

leftPane.AddItem("▶ ЗАПУСТИТЬ МОНИТОРИНГ", "", 0, func() {
toggleMonitoring()
})

leftPane.AddItem("⏹ ОСТАНОВИТЬ МОНИТОРИНГ", "", 0, func() {
toggleMonitoring()
})

leftPane.AddItem("🔍 БЫСТРОЕ СКАНИРОВАНИЕ", "", 0, func() {
quickScan()
})

leftPane.AddItem("📊 ПРОСМОТР ЛОГОВ", "", 0, func() {
showLogViewer()
})

leftPane.AddItem("📈 СТАТИСТИКА", "", 0, func() {
showStatistics()
})

leftPane.AddItem("🛠 НАСТРОЙКИ", "", 0, func() {
showSettings()
})

leftPane.AddItem("❌ ВЫХОД", "", 0, func() {
app.Stop()
})
}

func updateRightPane() {
rightPane.Clear()

// Отображаем список папок для мониторинга
rightPane.AddItem("[::b]📂 ОТСЛЕЖИВАЕМЫЕ ПАПКИ[::-]", "", 0, nil)
rightPane.AddItem("", "", 0, nil)

for i, folder := range cfg.CustomFolders {
icon := "  "
if folder == cfg.WatchDir {
icon = "★ "
}

folderName := folder
if len(folderName) > 30 {
folderName = "..." + folderName[len(folderName)-27:]
}

count := countFilesInFolder(folder)
text := fmt.Sprintf("%s[yellow]%s[-] ([cyan]%d[-] файлов)", icon, folderName, count)

idx := i // Capture for closure
rightPane.AddItem(text, "", 0, func() {
manageFolder(idx)
})
}

rightPane.AddItem("", "", 0, nil)
rightPane.AddItem("[::b]⚙ ОТСЛЕЖИВАЕМЫЕ РАСШИРЕНИЯ[::-]", "", 0, nil)
rightPane.AddItem("", "", 0, nil)

for _, ext := range cfg.Extensions {
rightPane.AddItem(fmt.Sprintf("  %s", ext), "", 0, nil)
}
}

func updateStatusBar() {
timeStr := time.Now().Format("15:04:05")

var helpText string
if currentPanel == "left" {
helpText = "[F1]Помощь [F2]Новая папка [F3]Удалить [F4]Редактировать [F5]Копировать [F6]Переместить [F7]Создать [F8]Удалить [F10]Выход"
} else {
helpText = "[Tab]Переключить панель [Enter]Выбрать [Ins]Добавить [Del]Удалить [F9]Настройки [Ctrl+Q]Выход"
}

statusText := fmt.Sprintf("[white]%s | %s", timeStr, helpText)
statusBar.SetText(statusText)
}

func setupHotkeys() {
app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
switch event.Key() {
case tcell.KeyF1:
showHelp()
return nil
case tcell.KeyF2:
addNewFolder()
return nil
case tcell.KeyF3:
deleteSelectedFolder()
return nil
case tcell.KeyF4:
editSelectedFolder()
return nil
case tcell.KeyF5:
copyFolder()
return nil
case tcell.KeyF6:
moveFolder()
return nil
case tcell.KeyF7:
createNewFolder()
return nil
case tcell.KeyF8:
deleteSelectedFolder()
return nil
case tcell.KeyF9:
showSettings()
return nil
case tcell.KeyF10:
app.Stop()
return nil
case tcell.KeyTab:
togglePanel()
return nil
case tcell.KeyEnter:
selectCurrentItem()
return nil
case tcell.KeyInsert:
addNewItem()
return nil
case tcell.KeyDelete:
deleteCurrentItem()
return nil
case tcell.KeyCtrlQ:
app.Stop()
return nil
case tcell.KeyCtrlM:
toggleMonitoring()
return nil
case tcell.KeyCtrlL:
showLogViewer()
return nil
case tcell.KeyCtrlS:
quickScan()
return nil
}

return event
})
}

// === ФУНКЦИИ ИНТЕРФЕЙСА ===
func togglePanel() {
if currentPanel == "left" {
currentPanel = "right"
app.SetFocus(rightPane)
} else {
currentPanel = "left"
app.SetFocus(leftPane)
}
updateStatusBar()
}

func selectCurrentItem() {
if currentPanel == "left" {
idx := leftPane.GetCurrentItem()
if idx >= 0 {
// Выполняем действие для выбранного элемента
leftPane.GetItem(idx).(*tview.List).SetSelectedFunc(idx)
}
} else {
idx := rightPane.GetCurrentItem()
if idx >= 0 {
rightPane.GetItem(idx).(*tview.List).SetSelectedFunc(idx)
}
}
}

func showHelp() {
modal := tview.NewModal().
SetText("[::b]ГОРЯЧИЕ КЛАВИШИ AILAN ARCHIVIST[::-]\n\n" +
"[yellow]F1[::-] - Эта справка\n" +
"[yellow]F2[::-] - Добавить новую папку\n" +
"[yellow]F3/F8[::-] - Удалить выбранную папку\n" +
"[yellow]F4[::-] - Редактировать папку\n" +
"[yellow]F5[::-] - Копировать папку\n" +
"[yellow]F6[::-] - Переместить папку\n" +
"[yellow]F7[::-] - Создать новую папку\n" +
"[yellow]F9[::-] - Настройки\n" +
"[yellow]F10/Ctrl+Q[::-] - Выход\n" +
"[yellow]Tab[::-] - Переключение между панелями\n" +
"[yellow]Enter[::-] - Выбрать элемент\n" +
"[yellow]Ins[::-] - Добавить элемент\n" +
"[yellow]Del[::-] - Удалить элемент\n" +
"[yellow]Ctrl+M[::-] - Вкл/Выкл мониторинг\n" +
"[yellow]Ctrl+L[::-] - Просмотр логов\n" +
"[yellow]Ctrl+S[::-] - Быстрое сканирование").
AddButtons([]string{"OK"}).
SetDoneFunc(func(buttonIndex int, buttonLabel string) {
pages.HidePage("help")
})

pages.AddPage("help", modal, true, true)
}

func showFolderManager() {
form := tview.NewForm().
AddInputField("Добавить папку:", "", 40, nil, nil).
AddButton("Добавить", func() {
// Получаем значение из поля
// Здесь будет логика добавления папки
pages.HidePage("folderManager")
}).
AddButton("Отмена", func() {
pages.HidePage("folderManager")
})

form.SetBorder(true).SetTitle(" 📁 УПРАВЛЕНИЕ ПАПКАМИ ")
pages.AddPage("folderManager", tview.NewCenter(form, 50, 10), true, true)
}

func showExtensionManager() {
// Создаем список расширений с чекбоксами
list := tview.NewList()

for _, ext := range cfg.Extensions {
list.AddItem(ext, "", 0, nil)
}

list.SetBorder(true).SetTitle(" ⚙ УПРАВЛЕНИЕ РАСШИРЕНИЯМИ ")

// Кнопки управления
flex := tview.NewFlex().SetDirection(tview.FlexRow)
flex.AddItem(list, 0, 1, true)

buttonFlex := tview.NewFlex().SetDirection(tview.FlexColumn)
buttonFlex.AddItem(tview.NewButton("Добавить").SetSelectedFunc(func() {
showAddExtensionDialog()
}), 0, 1, false)

buttonFlex.AddItem(tview.NewButton("Удалить").SetSelectedFunc(func() {
idx := list.GetCurrentItem()
if idx >= 0 {
// Удаляем расширение
cfg.Extensions = append(cfg.Extensions[:idx], cfg.Extensions[idx+1:]...)
showExtensionManager() // Обновляем
}
}), 0, 1, false)

buttonFlex.AddItem(tview.NewButton("Закрыть").SetSelectedFunc(func() {
pages.HidePage("extensionManager")
}), 0, 1, false)

flex.AddItem(buttonFlex, 1, 1, false)

pages.AddPage("extensionManager", tview.NewCenter(flex, 50, 20), true, true)
}

func showAddExtensionDialog() {
modal := tview.NewModal().
SetText("Введите новое расширение (начинается с точки):").
AddButtons([]string{".php", ".html", ".js", ".css", ".txt", ".json", ".py", ".java", "Другое", "Отмена"}).
SetDoneFunc(func(buttonIndex int, buttonLabel string) {
if buttonLabel == "Другое" {
showCustomExtensionDialog()
} else if buttonLabel != "Отмена" && buttonLabel != "" {
// Добавляем выбранное расширение
addExtension(buttonLabel)
pages.HidePage("addExtension")
showExtensionManager() // Обновляем
} else {
pages.HidePage("addExtension")
}
})

modal.SetTitle(" ➕ ДОБАВИТЬ РАСШИРЕНИЕ ")
pages.AddPage("addExtension", modal, true, true)
}

func showCustomExtensionDialog() {
form := tview.NewForm().
AddInputField("Расширение (начинается с точки):", ".", 20, nil, nil).
AddButton("Добавить", func() {
// Получаем значение
pages.HidePage("customExtension")
showExtensionManager() // Обновляем
}).
AddButton("Отмена", func() {
pages.HidePage("customExtension")
})

form.SetBorder(true).SetTitle(" ✏ ВВЕДИТЕ РАСШИРЕНИЕ ")
pages.AddPage("customExtension", tview.NewCenter(form, 50, 10), true, true)
}

func addExtension(ext string) {
// Проверяем, нет ли уже такого расширения
for _, existing := range cfg.Extensions {
if existing == ext {
return
}
}

cfg.Extensions = append(cfg.Extensions, ext)
addLogEntry(fmt.Sprintf("Добавлено расширение: %s", ext))
updateRightPane()
}

func toggleMonitoring() {
cfg.IsRunning = !cfg.IsRunning

if cfg.IsRunning {
addLogEntry("Мониторинг запущен")
} else {
addLogEntry("Мониторинг остановлен")
}

// Обновляем UI
updateUI()
}

func quickScan() {
go func() {
count := countAllTrackedFiles()
addLogEntry(fmt.Sprintf("Быстрое сканирование: найдено %d файлов", count))

app.QueueUpdateDraw(func() {
updateTopPanel(nil) // Обновляем верхнюю панель
})
}()
}

func showLogViewer() {
// Создаем модальное окно с логами
textView := tview.NewTextView().
SetDynamicColors(true).
SetScrollable(true)

textView.SetBorder(true).SetTitle(" 📝 ЖУРНАЛ СОБЫТИЙ ")

// Загружаем последние логи
logContent := loadRecentLogs(50)
textView.SetText(logContent)

// Кнопки
modal := tview.NewFlex().SetDirection(tview.FlexRow)
modal.AddItem(textView, 0, 1, true)

buttonFlex := tview.NewFlex().SetDirection(tview.FlexColumn)
buttonFlex.AddItem(tview.NewButton("Обновить").SetSelectedFunc(func() {
logContent := loadRecentLogs(50)
textView.SetText(logContent)
}), 0, 1, false)

buttonFlex.AddItem(tview.NewButton("Очистить").SetSelectedFunc(func() {
textView.SetText("")
}), 0, 1, false)

buttonFlex.AddItem(tview.NewButton("Закрыть").SetSelectedFunc(func() {
pages.HidePage("logViewer")
}), 0, 1, false)

modal.AddItem(buttonFlex, 1, 1, false)

center := tview.NewCenter(modal, 80, 20)
pages.AddPage("logViewer", center, true, true)
}

func showStatistics() {
modal := tview.NewModal()

totalFiles := countAllTrackedFiles()
stats := gatherStatistics()

text := fmt.Sprintf("[::b]📊 СТАТИСТИКА ПРОЕКТА[::-]\n\n"+
"[yellow]Всего отслеживаемых файлов:[-] %d\n"+
"[yellow]Папок для мониторинга:[-] %d\n"+
"[yellow]Отслеживаемых расширений:[-] %d\n\n"+
"[yellow]СТАТУС МОНИТОРИНГА:[-] ", 
totalFiles, len(cfg.CustomFolders), len(cfg.Extensions))

if cfg.IsRunning {
text += "[green]🟢 АКТИВЕН[-]"
} else {
text += "[red]🔴 ОСТАНОВЛЕН[-]"
}

text += "\n\n[yellow]РАСПРЕДЕЛЕНИЕ ПО РАСШИРЕНИЯМ:[-]\n"

for ext, count := range stats.ByExtension {
percentage := 0
if totalFiles > 0 {
percentage = (count * 100) / totalFiles
}
text += fmt.Sprintf("  %s: %d файлов (%d%%)\n", ext, count, percentage)
}

modal.SetText(text).
AddButtons([]string{"OK", "Экспорт", "Обновить"}).
SetDoneFunc(func(buttonIndex int, buttonLabel string) {
if buttonLabel == "Обновить" {
showStatistics() // Обновляем статистику
} else if buttonLabel == "Экспорт" {
exportStatistics()
} else {
pages.HidePage("statistics")
}
})

modal.SetTitle(" 📈 СТАТИСТИКА ")
pages.AddPage("statistics", modal, true, true)
}

func showSettings() {
form := tview.NewForm().
AddInputField("Папка для логов:", cfg.LogDir, 40, nil, nil).
AddInputField("Сводный файл:", cfg.SummaryFile, 40, nil, nil).
AddCheckbox("Автозапуск мониторинга", false, nil).
AddCheckbox("Звуковые уведомления", false, nil).
AddCheckbox("Подсветка изменений", true, nil).
AddButton("Сохранить", func() {
// Сохраняем настройки
pages.HidePage("settings")
}).
AddButton("Сброс", func() {
// Сбрасываем настройки
pages.HidePage("settings")
}).
AddButton("Отмена", func() {
pages.HidePage("settings")
})

form.SetBorder(true).SetTitle(" 🛠 НАСТРОЙКИ ")
pages.AddPage("settings", tview.NewCenter(form, 60, 20), true, true)
}

// === ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ ===
func addNewFolder() {
form := tview.NewForm().
AddInputField("Путь к новой папке:", "", 50, nil, nil).
AddButton("Добавить", func() {
// Логика добавления папки
pages.HidePage("newFolder")
updateRightPane()
}).
AddButton("Отмена", func() {
pages.HidePage("newFolder")
})

form.SetBorder(true).SetTitle(" 📁 ДОБАВИТЬ ПАПКУ ")
pages.AddPage("newFolder", tview.NewCenter(form, 60, 10), true, true)
}

func deleteSelectedFolder() {
idx := rightPane.GetCurrentItem()
if idx >= 2 && idx-2 < len(cfg.CustomFolders) {
// Показываем подтверждение
modal := tview.NewModal().
SetText(fmt.Sprintf("Удалить папку:\n%s?", cfg.CustomFolders[idx-2])).
AddButtons([]string{"Да", "Нет"}).
SetDoneFunc(func(buttonIndex int, buttonLabel string) {
if buttonLabel == "Да" {
// Удаляем папку
cfg.CustomFolders = append(cfg.CustomFolders[:idx-2], cfg.CustomFolders[idx-1:]...)
addLogEntry(fmt.Sprintf("Удалена папка: %s", cfg.CustomFolders[idx-2]))
updateRightPane()
}
pages.HidePage("confirmDelete")
})

modal.SetTitle(" 🗑 УДАЛЕНИЕ ПАПКИ ")
pages.AddPage("confirmDelete", modal, true, true)
}
}

func manageFolder(index int) {
if index < len(cfg.CustomFolders) {
folder := cfg.CustomFolders[index]

modal := tview.NewModal().
SetText(fmt.Sprintf("Папка: %s\n\nВыберите действие:", folder)).
AddButtons([]string{"Сделать основной", "Просмотреть файлы", "Сканировать", "Изменить", "Удалить", "Отмена"}).
SetDoneFunc(func(buttonIndex int, buttonLabel string) {
switch buttonLabel {
case "Сделать основной":
cfg.WatchDir = folder
addLogEntry(fmt.Sprintf("Основная папка изменена: %s", folder))
updateRightPane()
case "Просмотреть файлы":
showFolderFiles(folder)
case "Сканировать":
count := countFilesInFolder(folder)
addLogEntry(fmt.Sprintf("Сканирование папки %s: %d файлов", folder, count))
case "Изменить":
editFolder(index)
case "Удалить":
deleteFolder(index)
}
pages.HidePage("manageFolder")
})

modal.SetTitle(" 📂 УПРАВЛЕНИЕ ПАПКОЙ ")
pages.AddPage("manageFolder", modal, true, true)
}
}

func showFolderFiles(folder string) {
textView := tview.NewTextView().
SetDynamicColors(true).
SetScrollable(true)

textView.SetBorder(true).SetTitle(fmt.Sprintf(" 📁 ФАЙЛЫ: %s ", folder))

// Собираем список файлов
var files []string
filepath.WalkDir(folder, func(path string, d fs.DirEntry, err error) error {
if err != nil || d.IsDir() {
return nil
}

ext := filepath.Ext(path)
for _, tracked := range cfg.Extensions {
if ext == tracked {
relPath, _ := filepath.Rel(folder, path)
files = append(files, relPath)
break
}
}
return nil
})

sort.Strings(files)

content := fmt.Sprintf("[yellow]Всего файлов: %d[-]\n\n", len(files))
for _, file := range files {
content += file + "\n"
}

textView.SetText(content)

modal := tview.NewFlex().SetDirection(tview.FlexRow)
modal.AddItem(textView, 0, 1, true)

buttonFlex := tview.NewFlex().SetDirection(tview.FlexColumn)
buttonFlex.AddItem(tview.NewButton("Закрыть").SetSelectedFunc(func() {
pages.HidePage("folderFiles")
}), 0, 1, false)

modal.AddItem(buttonFlex, 1, 1, false)

center := tview.NewCenter(modal, 70, 20)
pages.AddPage("folderFiles", center, true, true)
}

func editFolder(index int) {
form := tview.NewForm().
AddInputField("Новый путь:", cfg.CustomFolders[index], 50, nil, nil).
AddButton("Сохранить", func() {
// Сохраняем изменения
pages.HidePage("editFolder")
updateRightPane()
}).
AddButton("Отмена", func() {
pages.HidePage("editFolder")
})

form.SetBorder(true).SetTitle(" ✏ РЕДАКТИРОВАНИЕ ПАПКИ ")
pages.AddPage("editFolder", tview.NewCenter(form, 60, 10), true, true)
}

func deleteFolder(index int) {
modal := tview.NewModal().
SetText(fmt.Sprintf("Удалить папку:\n%s?", cfg.CustomFolders[index])).
AddButtons([]string{"Да", "Нет"}).
SetDoneFunc(func(buttonIndex int, buttonLabel string) {
if buttonLabel == "Да" {
removed := cfg.CustomFolders[index]
cfg.CustomFolders = append(cfg.CustomFolders[:index], cfg.CustomFolders[index+1:]...)
addLogEntry(fmt.Sprintf("Удалена папка: %s", removed))
updateRightPane()
}
pages.HidePage("deleteFolder")
})

modal.SetTitle(" 🗑 УДАЛЕНИЕ ПАПКИ ")
pages.AddPage("deleteFolder", modal, true, true)
}

func addNewItem() {
if currentPanel == "right" {
addNewFolder()
}
}

func deleteCurrentItem() {
if currentPanel == "right" {
deleteSelectedFolder()
}
}

func addLogEntry(message string) {
timestamp := time.Now().Format("15:04:05")
logEntry := fmt.Sprintf("[gray]%s[-] %s\n", timestamp, message)

currentText := logView.GetText(true)
logView.SetText(currentText + logEntry)

// Прокручиваем вниз
logView.ScrollToEnd()

// Также записываем в файл
logToFile(message)
}

func logToFile(message string) {
dateStr := time.Now().Format("2006-01-02")
logFile := filepath.Join(cfg.LogDir, dateStr+"_changes.md")

entry := fmt.Sprintf("### %s\n", time.Now().Format("15:04:05"))
entry += fmt.Sprintf("- **Событие:** %s\n", message)
entry += fmt.Sprintf("- **Время:** %s\n\n", time.Now().Format("2006-01-02 15:04:05"))

content, err := os.ReadFile(logFile)
if err != nil {
header := fmt.Sprintf("# Изменения за %s\n\n", dateStr)
entry = header + entry
} else {
entry = string(content) + "\n" + entry
}

os.WriteFile(logFile, []byte(entry), 0644)
}

func loadRecentLogs(count int) string {
today := time.Now().Format("2006-01-02")
logFile := filepath.Join(cfg.LogDir, today+"_changes.md")

content, err := os.ReadFile(logFile)
if err != nil {
return "Логи отсутствуют"
}

lines := strings.Split(string(content), "\n")

// Берем последние count строк
start := len(lines) - count
if start < 0 {
start = 0
}

// Добавляем цвета для лучшей читаемости
var result strings.Builder
for i := start; i < len(lines); i++ {
line := lines[i]

// Добавляем цвета в зависимости от типа сообщения
if strings.Contains(line, "запущен") || strings.Contains(line, "добавлен") {
line = strings.ReplaceAll(line, "**Событие:**", "[green]**Событие:**[-]")
} else if strings.Contains(line, "остановлен") || strings.Contains(line, "удален") {
line = strings.ReplaceAll(line, "**Событие:**", "[red]**Событие:**[-]")
} else if strings.Contains(line, "сканирование") {
line = strings.ReplaceAll(line, "**Событие:**", "[yellow]**Событие:**[-]")
}

result.WriteString(line + "\n")
}

return result.String()
}

func updateUI() {
updateTopPanel(nil)
updateRightPane()
updateStatusBar()

// Также обновляем заголовки панелей
leftPane.SetTitle(fmt.Sprintf(" [::b]📁 ЛЕВАЯ ПАНЕЛЬ[::-] %s", 
func() string {
if currentPanel == "left" {
return "[green]◄ АКТИВНА[-]"
}
return ""
}()))

rightPane.SetTitle(fmt.Sprintf(" [::b]📊 ПРАВАЯ ПАНЕЛЬ[::-] %s", 
func() string {
if currentPanel == "right" {
return "[green]◄ АКТИВНА[-]"
}
return ""
}()))
}

func countAllTrackedFiles() int {
total := 0
for _, folder := range cfg.CustomFolders {
total += countFilesInFolder(folder)
}
return total
}

func countFilesInFolder(folder string) int {
count := 0
filepath.WalkDir(folder, func(path string, d fs.DirEntry, err error) error {
if err != nil || d.IsDir() {
return nil
}

ext := filepath.Ext(path)
for _, tracked := range cfg.Extensions {
if ext == tracked {
count++
break
}
}
return nil
})
return count
}

type Statistics struct {
TotalFiles    int
ByExtension   map[string]int
ByFolder      map[string]int
}

func gatherStatistics() Statistics {
stats := Statistics{
ByExtension: make(map[string]int),
ByFolder:    make(map[string]int),
}

for _, folder := range cfg.CustomFolders {
folderCount := 0

filepath.WalkDir(folder, func(path string, d fs.DirEntry, err error) error {
if err != nil || d.IsDir() {
return nil
}

ext := filepath.Ext(path)
for _, tracked := range cfg.Extensions {
if ext == tracked {
stats.ByExtension[ext]++
folderCount++
stats.TotalFiles++
break
}
}
return nil
})

stats.ByFolder[folder] = folderCount
}

return stats
}

func exportStatistics() {
stats := gatherStatistics()

exportContent := fmt.Sprintf("AILAN Archivist - Статистика\n")
exportContent += fmt.Sprintf("Дата экспорта: %s\n\n", time.Now().Format("2006-01-02 15:04:05"))

exportContent += fmt.Sprintf("Всего файлов: %d\n", stats.TotalFiles)
exportContent += fmt.Sprintf("Папок для мониторинга: %d\n", len(cfg.CustomFolders))
exportContent += fmt.Sprintf("Отслеживаемых расширений: %d\n\n", len(cfg.Extensions))

exportContent += "По расширениям:\n"
for ext, count := range stats.ByExtension {
percentage := 0
if stats.TotalFiles > 0 {
percentage = (count * 100) / stats.TotalFiles
}
exportContent += fmt.Sprintf("  %s: %d файлов (%d%%)\n", ext, count, percentage)
}

exportContent += "\nПо папкам:\n"
for folder, count := range stats.ByFolder {
percentage := 0
if stats.TotalFiles > 0 {
percentage = (count * 100) / stats.TotalFiles
}
exportContent += fmt.Sprintf("  %s: %d файлов (%d%%)\n", folder, count, percentage)
}

// Сохраняем в файл
filename := fmt.Sprintf("statistics_%s.txt", time.Now().Format("20060102_150405"))
err := os.WriteFile(filename, []byte(exportContent), 0644)

if err != nil {
addLogEntry(fmt.Sprintf("Ошибка экспорта статистики: %v", err))
} else {
addLogEntry(fmt.Sprintf("Статистика экспортирована в %s", filename))
}
}
