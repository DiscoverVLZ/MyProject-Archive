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
fmt.Printf("Ошибка запуска: %v\n", err)
os.Exit(1)
}
}

func initConfig() {
wd, _ := os.Getwd()
cfg = Config{
WatchDir:      wd,
Extensions:    []string{".php", ".html", ".js", ".css", ".txt", ".json"},
LogDir:        "docs/changelog",
SummaryFile:   "docs/project_state.md",
IsRunning:     false,
CustomFolders: []string{wd},
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

// === ЛЕВАЯ ПАНЕЛЬ (Команды) ===
leftPane = tview.NewList().
ShowSecondaryText(false).
SetHighlightFullLine(true)

leftPane.SetBorder(true).
SetTitle(" [::b]📁 КОМАНДЫ[::-] ").
SetTitleAlign(tview.AlignLeft)

updateLeftPane()

// === ПРАВАЯ ПАНЕЛЬ (Папки и расширения) ===
rightPane = tview.NewList().
ShowSecondaryText(false).
SetHighlightFullLine(true)

rightPane.SetBorder(true).
SetTitle(" [::b]📊 СОДЕРЖИМОЕ[::-] ").
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

if panel != nil {
panel.SetText(text)
}
}

func updateLeftPane() {
leftPane.Clear()

// Команды управления
leftPane.AddItem("📁 УПРАВЛЕНИЕ ПАПКАМИ", "Добавить/удалить папки для мониторинга", 'P', func() {
showFolderManager()
})

leftPane.AddItem("⚙ УПРАВЛЕНИЕ РАСШИРЕНИЯМИ", "Настроить отслеживаемые расширения", 'E', func() {
showExtensionManager()
})

if cfg.IsRunning {
leftPane.AddItem("⏹ ОСТАНОВИТЬ МОНИТОРИНГ", "Остановить отслеживание файлов", 'S', func() {
toggleMonitoring()
})
} else {
leftPane.AddItem("▶ ЗАПУСТИТЬ МОНИТОРИНГ", "Начать отслеживание файлов", 'S', func() {
toggleMonitoring()
})
}

leftPane.AddItem("🔍 БЫСТРОЕ СКАНИРОВАНИЕ", "Просканировать все папки сейчас", 'F', func() {
quickScan()
})

leftPane.AddItem("📊 ПРОСМОТР ЛОГОВ", "Посмотреть историю изменений", 'L', func() {
showLogViewer()
})

leftPane.AddItem("📈 СТАТИСТИКА", "Показать статистику проекта", 'T', func() {
showStatistics()
})

leftPane.AddItem("🛠 НАСТРОЙКИ", "Настройки программы", 'N', func() {
showSettings()
})

leftPane.AddItem("❌ ВЫХОД", "Завершить работу программы", 'Q', func() {
app.Stop()
})
}

func updateRightPane() {
rightPane.Clear()

// Отображаем список папок для мониторинга
rightPane.AddItem("[::b]📂 ОТСЛЕЖИВАЕМЫЕ ПАПКИ[::-]", "", 0, nil)

for i, folder := range cfg.CustomFolders {
icon := "  "
if folder == cfg.WatchDir {
icon = "★ "
}

folderName := folder
if len(folderName) > 40 {
folderName = "..." + folderName[len(folderName)-37:]
}

count := countFilesInFolder(folder)
text := fmt.Sprintf("%s[yellow]%s[-] ([cyan]%d[-] файлов)", icon, folderName, count)

idx := i
rightPane.AddItem(text, "", rune('1'+i), func() {
manageFolder(idx)
})
}

rightPane.AddItem("", "", 0, nil)
rightPane.AddItem("[::b]⚙ ОТСЛЕЖИВАЕМЫЕ РАСШИРЕНИЯ[::-]", "", 0, nil)

for i, ext := range cfg.Extensions {
rightPane.AddItem(fmt.Sprintf("  %s", ext), "", rune('a'+i), nil)
}
}

func updateStatusBar() {
timeStr := time.Now().Format("15:04:05")

var helpText string
if currentPanel == "left" {
helpText = "[F1]Помощь [F2]Добавить [F3]Удалить [F4]Изменить [F5]Запуск [F6]Остановка [F10]Выход"
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
toggleMonitoring()
return nil
case tcell.KeyF6:
toggleMonitoring()
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
addNewFolder()
return nil
case tcell.KeyDelete:
deleteSelectedFolder()
return nil
case tcell.KeyCtrlQ:
app.Stop()
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
leftPane.SetCurrentItem(idx)
}
} else {
idx := rightPane.GetCurrentItem()
if idx >= 0 {
rightPane.SetCurrentItem(idx)
}
}
}

func showHelp() {
modal := tview.NewModal().
SetText("[::b]ГОРЯЧИЕ КЛАВИШИ AILAN ARCHIVIST[::-]\n\n" +
"[yellow]F1[::-] - Эта справка\n" +
"[yellow]F2/Ins[::-] - Добавить новую папку\n" +
"[yellow]F3/Del[::-] - Удалить выбранную папку\n" +
"[yellow]F4[::-] - Редактировать папку\n" +
"[yellow]F5[::-] - Запустить мониторинг\n" +
"[yellow]F6[::-] - Остановить мониторинг\n" +
"[yellow]F9[::-] - Настройки\n" +
"[yellow]F10/Ctrl+Q[::-] - Выход\n" +
"[yellow]Tab[::-] - Переключение между панелями\n" +
"[yellow]Enter[::-] - Выбрать элемент\n" +
"[yellow]Ctrl+L[::-] - Просмотр логов\n" +
"[yellow]Ctrl+S[::-] - Быстрое сканирование").
AddButtons([]string{"OK"}).
SetDoneFunc(func(buttonIndex int, buttonLabel string) {
pages.RemovePage("help")
})

modal.SetTitle(" ❓ СПРАВКА ")
pages.AddPage("help", modal, false, true)
}

func showFolderManager() {
form := tview.NewForm().
AddInputField("Добавить папку:", "", 40, nil, nil).
AddButton("Добавить", func() {
field := form.GetFormItem(0).(*tview.InputField)
newFolder := strings.TrimSpace(field.GetText())
if newFolder != "" {
addFolder(newFolder)
}
pages.RemovePage("folderManager")
}).
AddButton("Отмена", func() {
pages.RemovePage("folderManager")
})

form.SetBorder(true).SetTitle(" 📁 УПРАВЛЕНИЕ ПАПКАМИ ")

flex := tview.NewFlex().
AddItem(nil, 0, 1, false).
AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
AddItem(nil, 0, 1, false).
AddItem(form, 10, 1, true).
AddItem(nil, 0, 1, false), 60, 1, true).
AddItem(nil, 0, 1, false)

pages.AddPage("folderManager", flex, false, true)
}

func showExtensionManager() {
// Создаем модальное окно для управления расширениями
modal := tview.NewModal().
SetText("Управление расширениями файлов\n\nВыберите действие:").
AddButtons([]string{"Добавить", "Удалить", "Сбросить", "Отмена"}).
SetDoneFunc(func(buttonIndex int, buttonLabel string) {
switch buttonLabel {
case "Добавить":
showAddExtensionDialog()
case "Удалить":
showRemoveExtensionDialog()
case "Сбросить":
resetExtensions()
}
pages.RemovePage("extensionManager")
})

modal.SetTitle(" ⚙ УПРАВЛЕНИЕ РАСШИРЕНИЯМИ ")
pages.AddPage("extensionManager", modal, false, true)
}

func showAddExtensionDialog() {
modal := tview.NewModal().
SetText("Выберите расширение для добавления:").
AddButtons([]string{".py", ".java", ".cpp", ".xml", ".yml", ".md", ".sql", "Другое", "Отмена"}).
SetDoneFunc(func(buttonIndex int, buttonLabel string) {
if buttonLabel == "Другое" {
showCustomExtensionDialog()
} else if buttonLabel != "Отмена" && buttonLabel != "" {
addExtension(buttonLabel)
}
pages.RemovePage("addExtension")
})

modal.SetTitle(" ➕ ДОБАВИТЬ РАСШИРЕНИЕ ")
pages.AddPage("addExtension", modal, false, true)
}

func showCustomExtensionDialog() {
form := tview.NewForm().
AddInputField("Расширение (начинается с точки):", ".", 20, nil, nil).
AddButton("Добавить", func() {
field := form.GetFormItem(0).(*tview.InputField)
ext := strings.TrimSpace(field.GetText())
if ext != "" && strings.HasPrefix(ext, ".") {
addExtension(ext)
}
pages.RemovePage("customExtension")
}).
AddButton("Отмена", func() {
pages.RemovePage("customExtension")
})

form.SetBorder(true).SetTitle(" ✏ ВВЕДИТЕ РАСШИРЕНИЕ ")

flex := tview.NewFlex().
AddItem(nil, 0, 1, false).
AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
AddItem(nil, 0, 1, false).
AddItem(form, 10, 1, true).
AddItem(nil, 0, 1, false), 50, 1, true).
AddItem(nil, 0, 1, false)

pages.AddPage("customExtension", flex, false, true)
}

func showRemoveExtensionDialog() {
if len(cfg.Extensions) <= 1 {
showMessage("Нельзя удалить все расширения!", "error")
return
}

modal := tview.NewModal().
SetText("Выберите расширение для удаления:").
AddButtons(append(cfg.Extensions, "Отмена")).
SetDoneFunc(func(buttonIndex int, buttonLabel string) {
if buttonLabel != "Отмена" && buttonLabel != "" {
removeExtension(buttonLabel)
}
pages.RemovePage("removeExtension")
})

modal.SetTitle(" ➖ УДАЛИТЬ РАСШИРЕНИЕ ")
pages.AddPage("removeExtension", modal, false, true)
}

func addExtension(ext string) {
// Проверяем, нет ли уже такого расширения
for _, existing := range cfg.Extensions {
if existing == ext {
showMessage("Это расширение уже есть в списке", "warning")
return
}
}

cfg.Extensions = append(cfg.Extensions, ext)
addLogEntry(fmt.Sprintf("Добавлено расширение: %s", ext))
updateRightPane()
updateTopPanel(nil)
showMessage(fmt.Sprintf("Добавлено расширение: %s", ext), "success")
}

func removeExtension(ext string) {
newExtensions := []string{}
for _, existing := range cfg.Extensions {
if existing != ext {
newExtensions = append(newExtensions, existing)
}
}

cfg.Extensions = newExtensions
addLogEntry(fmt.Sprintf("Удалено расширение: %s", ext))
updateRightPane()
updateTopPanel(nil)
showMessage(fmt.Sprintf("Удалено расширение: %s", ext), "success")
}

func resetExtensions() {
cfg.Extensions = []string{".php", ".html", ".js", ".css", ".txt", ".json"}
addLogEntry("Расширения сброшены к стандартным")
updateRightPane()
updateTopPanel(nil)
showMessage("Расширения сброшены к стандартным", "success")
}

func toggleMonitoring() {
cfg.IsRunning = !cfg.IsRunning

if cfg.IsRunning {
addLogEntry("Мониторинг запущен")
showMessage("Мониторинг запущен", "success")
} else {
addLogEntry("Мониторинг остановлен")
showMessage("Мониторинг остановлен", "success")
}

// Обновляем UI
updateUI()
}

func quickScan() {
go func() {
count := countAllTrackedFiles()
msg := fmt.Sprintf("Быстрое сканирование: найдено %d файлов", count)
addLogEntry(msg)

app.QueueUpdateDraw(func() {
showMessage(msg, "info")
updateTopPanel(nil)
})
}()
}

func showLogViewer() {
// Показываем простой просмотр логов
modal := tview.NewModal().
SetText("Для просмотра полных логов откройте папку:\n" + cfg.LogDir + "\n\nИли используйте журнал событий внизу окна.").
AddButtons([]string{"Открыть папку", "Закрыть"}).
SetDoneFunc(func(buttonIndex int, buttonLabel string) {
if buttonLabel == "Открыть папку" {
openLogsFolder()
}
pages.RemovePage("logViewer")
})

modal.SetTitle(" 📊 ПРОСМОТР ЛОГОВ ")
pages.AddPage("logViewer", modal, false, true)
}

func showStatistics() {
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

modal := tview.NewModal().
SetText(text).
AddButtons([]string{"OK", "Экспорт", "Обновить"}).
SetDoneFunc(func(buttonIndex int, buttonLabel string) {
if buttonLabel == "Обновить" {
pages.RemovePage("statistics")
showStatistics()
} else if buttonLabel == "Экспорт" {
exportStatistics()
} else {
pages.RemovePage("statistics")
}
})

modal.SetTitle(" 📈 СТАТИСТИКА ")
pages.AddPage("statistics", modal, false, true)
}

func showSettings() {
form := tview.NewForm().
AddInputField("Папка для логов:", cfg.LogDir, 40, nil, nil).
AddInputField("Сводный файл:", cfg.SummaryFile, 40, nil, nil).
AddCheckbox("Автозапуск мониторинга", false, nil).
AddCheckbox("Подсветка изменений", true, nil).
AddButton("Сохранить", func() {
// Сохраняем настройки
logDirField := form.GetFormItem(0).(*tview.InputField)
summaryField := form.GetFormItem(1).(*tview.InputField)

cfg.LogDir = logDirField.GetText()
cfg.SummaryFile = summaryField.GetText()

os.MkdirAll(cfg.LogDir, 0755)
addLogEntry("Настройки сохранены")
showMessage("Настройки сохранены", "success")
pages.RemovePage("settings")
}).
AddButton("Сброс", func() {
resetSettings()
pages.RemovePage("settings")
}).
AddButton("Отмена", func() {
pages.RemovePage("settings")
})

form.SetBorder(true).SetTitle(" 🛠 НАСТРОЙКИ ")

flex := tview.NewFlex().
AddItem(nil, 0, 1, false).
AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
AddItem(nil, 0, 1, false).
AddItem(form, 15, 1, true).
AddItem(nil, 0, 1, false), 60, 1, true).
AddItem(nil, 0, 1, false)

pages.AddPage("settings", flex, false, true)
}

// === ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ ===
func addNewFolder() {
showFolderManager()
}

func deleteSelectedFolder() {
if currentPanel == "right" {
idx := rightPane.GetCurrentItem()
if idx >= 1 && idx-1 < len(cfg.CustomFolders) {
deleteFolder(idx - 1)
}
}
}

func editSelectedFolder() {
if currentPanel == "right" {
idx := rightPane.GetCurrentItem()
if idx >= 1 && idx-1 < len(cfg.CustomFolders) {
editFolder(idx - 1)
}
}
}

func addFolder(newFolder string) {
// Проверяем существование папки
if info, err := os.Stat(newFolder); err != nil || !info.IsDir() {
showMessage("Ошибка: папка не существует или недоступна", "error")
return
}

// Проверяем, нет ли уже такой папки
for _, folder := range cfg.CustomFolders {
if folder == newFolder {
showMessage("Эта папка уже в списке", "warning")
return
}
}

cfg.CustomFolders = append(cfg.CustomFolders, newFolder)
addLogEntry(fmt.Sprintf("Добавлена папка: %s", newFolder))
updateRightPane()
updateTopPanel(nil)
showMessage(fmt.Sprintf("Добавлена папка: %s", newFolder), "success")
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
showMessage(fmt.Sprintf("Основная папка: %s", folder), "success")
case "Просмотреть файлы":
showFolderFiles(folder)
case "Сканировать":
count := countFilesInFolder(folder)
msg := fmt.Sprintf("Сканирование папки %s: %d файлов", folder, count)
addLogEntry(msg)
showMessage(msg, "info")
case "Изменить":
editFolder(index)
case "Удалить":
deleteFolder(index)
}
pages.RemovePage("manageFolder")
})

modal.SetTitle(" 📂 УПРАВЛЕНИЕ ПАПКОЙ ")
pages.AddPage("manageFolder", modal, false, true)
}
}

func showFolderFiles(folder string) {
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

content := fmt.Sprintf("[yellow]Папка: %s[-]\n", folder)
content += fmt.Sprintf("[yellow]Всего файлов: %d[-]\n\n", len(files))

for _, file := range files {
content += file + "\n"
}

modal := tview.NewModal().
SetText(content).
AddButtons([]string{"Закрыть"}).
SetDoneFunc(func(buttonIndex int, buttonLabel string) {
pages.RemovePage("folderFiles")
})

modal.SetTitle(fmt.Sprintf(" 📁 ФАЙЛЫ: %s ", folder))
pages.AddPage("folderFiles", modal, false, true)
}

func editFolder(index int) {
folder := cfg.CustomFolders[index]

form := tview.NewForm().
AddInputField("Новый путь:", folder, 50, nil, nil).
AddButton("Сохранить", func() {
field := form.GetFormItem(0).(*tview.InputField)
newPath := strings.TrimSpace(field.GetText())

if newPath != "" && newPath != folder {
// Проверяем существование
if info, err := os.Stat(newPath); err != nil || !info.IsDir() {
showMessage("Ошибка: папка не существует", "error")
} else {
cfg.CustomFolders[index] = newPath
addLogEntry(fmt.Sprintf("Папка изменена: %s → %s", folder, newPath))
updateRightPane()
showMessage(fmt.Sprintf("Папка изменена: %s", newPath), "success")
}
}
pages.RemovePage("editFolder")
}).
AddButton("Отмена", func() {
pages.RemovePage("editFolder")
})

form.SetBorder(true).SetTitle(" ✏ РЕДАКТИРОВАНИЕ ПАПКИ ")

flex := tview.NewFlex().
AddItem(nil, 0, 1, false).
AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
AddItem(nil, 0, 1, false).
AddItem(form, 10, 1, true).
AddItem(nil, 0, 1, false), 60, 1, true).
AddItem(nil, 0, 1, false)

pages.AddPage("editFolder", flex, false, true)
}

func deleteFolder(index int) {
if len(cfg.CustomFolders) <= 1 {
showMessage("Нельзя удалить все папки!", "error")
return
}

folder := cfg.CustomFolders[index]

modal := tview.NewModal().
SetText(fmt.Sprintf("Удалить папку:\n%s?", folder)).
AddButtons([]string{"Да", "Нет"}).
SetDoneFunc(func(buttonIndex int, buttonLabel string) {
if buttonLabel == "Да" {
cfg.CustomFolders = append(cfg.CustomFolders[:index], cfg.CustomFolders[index+1:]...)
addLogEntry(fmt.Sprintf("Удалена папка: %s", folder))
updateRightPane()
updateTopPanel(nil)
showMessage(fmt.Sprintf("Удалена папка: %s", folder), "success")
}
pages.RemovePage("deleteFolder")
})

modal.SetTitle(" 🗑 УДАЛЕНИЕ ПАПКИ ")
pages.AddPage("deleteFolder", modal, false, true)
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

func showMessage(message string, msgType string) {
var prefix, color string

switch msgType {
case "success":
prefix = "✅ "
color = "[green]"
case "error":
prefix = "❌ "
color = "[red]"
case "warning":
prefix = "⚠  "
color = "[yellow]"
case "info":
prefix = "ℹ  "
color = "[blue]"
default:
prefix = "• "
color = "[white]"
}

addLogEntry(prefix + message)
}

func updateUI() {
updateLeftPane()
updateRightPane()
updateTopPanel(nil)
updateStatusBar()
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
showMessage(fmt.Sprintf("Ошибка экспорта статистики: %v", err), "error")
} else {
showMessage(fmt.Sprintf("Статистика экспортирована в %s", filename), "success")
}
}

func openLogsFolder() {
logPath, _ := filepath.Abs(cfg.LogDir)
os.StartProcess("explorer.exe", []string{logPath}, &os.ProcAttr{})
showMessage("Проводник открыт с папкой логов", "success")
}

func resetSettings() {
cfg.LogDir = "docs/changelog"
cfg.SummaryFile = "docs/project_state.md"
addLogEntry("Настройки сброшены к значениям по умолчанию")
showMessage("Настройки сброшены", "success")
}
