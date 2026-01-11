package main

import (
"fmt"
"os"
"path/filepath"

"time"

"github.com/gdamore/tcell/v2"
"github.com/rivo/tview"
)

// Структура для конфигурации
type Config struct {
Folders     []string
Extensions  []string
LogDir      string
TotalFiles  int
Monitoring  bool
CurrentDir  string
}

var (
app         *tview.Application
leftPane    *tview.List
rightPane   *tview.List
statusBar   *tview.TextView
logView     *tview.TextView
cfg         Config
currentPanel string = "left"
)

func main() {
app = tview.NewApplication()

// Инициализация конфигурации
initConfig()

// Настройка стилей Total Commander
tview.Styles.PrimitiveBackgroundColor = tcell.ColorBlack
tview.Styles.BorderColor = tcell.ColorWhite
tview.Styles.TitleColor = tcell.ColorYellow
tview.Styles.PrimaryTextColor = tcell.ColorWhite

createUI()

if err := app.SetRoot(createUI(), true).EnableMouse(true).Run(); err != nil {
fmt.Printf("Ошибка: %v\n", err)
os.Exit(1)
}
}

func initConfig() {
dir, _ := os.Getwd()
cfg = Config{
Folders:    []string{dir},
Extensions: []string{".php", ".html", ".js", ".css", ".txt", ".json", ".md", ".go"},
LogDir:     filepath.Join(dir, "docs", "changelog"),
TotalFiles: 0,
Monitoring: false,
CurrentDir: dir,
}

// Создаем папку для логов
os.MkdirAll(cfg.LogDir, 0755)
}

func createUI() tview.Primitive {
// === ВЕРХНЯЯ ПАНЕЛЬ ===
topPanel := tview.NewTextView().
SetDynamicColors(true).
SetTextAlign(tview.AlignCenter)

updateTopPanel(topPanel)

// === ЛЕВАЯ ПАНЕЛЬ (КОМАНДЫ) ===
leftPane = tview.NewList().
ShowSecondaryText(true).
SetHighlightFullLine(true)

leftPane.SetBorder(true).
SetTitle(" [yellow]🖥 КОМАНДЫ[-] ").
SetTitleAlign(tview.AlignLeft).
SetBackgroundColor(tcell.ColorBlack)

updateLeftPane()

// === ПРАВАЯ ПАНЕЛЬ (ПАПКИ) ===
rightPane = tview.NewList().
ShowSecondaryText(true).
SetHighlightFullLine(true)

rightPane.SetBorder(true).
SetTitle(" [yellow]📁 ПАПКИ МОНИТОРИНГА[-] ").
SetTitleAlign(tview.AlignLeft).
SetBackgroundColor(tcell.ColorBlack)

updateRightPane()

// === ПАНЕЛЬ ЛОГОВ ===
logView = tview.NewTextView().
SetDynamicColors(true).
SetScrollable(true)

logView.SetBorder(true).
SetTitle(" [yellow]📝 ЖУРНАЛ СОБЫТИЙ[-] ").
SetBackgroundColor(tcell.ColorBlack)

logView.SetText("[gray]Готов к работе...\nНажмите F4 для добавления папки[-]")

// === ПАНЕЛЬ СТАТУСА ===
statusBar = tview.NewTextView().
SetDynamicColors(true)

updateStatusBar()

// === ОСНОВНОЙ LAYOUT ===
mainFlex := tview.NewFlex().SetDirection(tview.FlexRow)

// Верхняя панель
mainFlex.AddItem(topPanel, 1, 0, false)

// Основные панели
panelsFlex := tview.NewFlex().SetDirection(tview.FlexColumn)
panelsFlex.AddItem(leftPane, 0, 1, true)
panelsFlex.AddItem(rightPane, 0, 1, false)

mainFlex.AddItem(panelsFlex, 0, 3, true)
mainFlex.AddItem(logView, 10, 1, false)
mainFlex.AddItem(statusBar, 1, 0, false)

// Настройка горячих клавиш
setupHotkeys()

return mainFlex
}

func updateTopPanel(panel *tview.TextView) {
status := "[red]🔴 ВЫКЛ"
if cfg.Monitoring {
status = "[green]🟢 ВКЛ"
}

text := fmt.Sprintf("[white]🚀 AILAN ARCHIVIST | Мониторинг: %s | Папок: [yellow]%d[-] | Файлов: [yellow]%d[-]", 
status, len(cfg.Folders), cfg.TotalFiles)
panel.SetText(text)
panel.SetBackgroundColor(tcell.ColorDarkBlue)
}

func updateLeftPane() {
leftPane.Clear()

// Заголовок
leftPane.AddItem("[yellow]🚀 ОСНОВНЫЕ КОМАНДЫ[-]", "", 0, nil)

// Команды управления
if cfg.Monitoring {
leftPane.AddItem("[green]⏸ Остановить мониторинг[-]", "Нажмите Enter или F5", 'S', func() {
toggleMonitoring()
})
} else {
leftPane.AddItem("[green]▶ Запустить мониторинг[-]", "Нажмите Enter или F5", 'S', func() {
toggleMonitoring()
})
}

leftPane.AddItem("[cyan]📁 Добавить папку[-]", "Нажмите Enter или F4", 'A', func() {
showAddFolderDialog()
})

leftPane.AddItem("[cyan]🗑 Удалить папку[-]", "Нажмите Enter или F8", 'D', func() {
deleteSelectedFolder()
})

leftPane.AddItem("[cyan]🔍 Быстрое сканирование[-]", "Нажмите Enter или F7", 'Q', func() {
quickScan()
})

// Разделитель
leftPane.AddItem("", "", 0, nil)
leftPane.AddItem("[yellow]📊 ИНФОРМАЦИЯ[-]", "", 0, nil)

leftPane.AddItem("[white]📈 Статистика[-]", "Нажмите Enter или F2", 'T', func() {
showStatistics()
})

leftPane.AddItem("[white]📝 Просмотр логов[-]", "Нажмите Enter или F3", 'L', func() {
showLogViewer()
})

leftPane.AddItem("[white]⚙ Настройки[-]", "Нажмите Enter или F9", 'N', func() {
showSettings()
})

// Разделитель
leftPane.AddItem("", "", 0, nil)
leftPane.AddItem("[yellow]🚪 ВЫХОД[-]", "", 0, nil)

leftPane.AddItem("[red]❌ Выход[-]", "Нажмите Enter или F10", 'X', func() {
app.Stop()
})
}

func updateRightPane() {
rightPane.Clear()

// Заголовок
rightPane.AddItem("[yellow]📂 ОТСЛЕЖИВАЕМЫЕ ПАПКИ[-]", "", 0, nil)

if len(cfg.Folders) == 0 {
rightPane.AddItem("[gray]Нет добавленных папок[-]", "Нажмите F4 чтобы добавить", 0, func() {
showAddFolderDialog()
})
} else {
for i, folder := range cfg.Folders {
// Обрезаем длинное имя папки
folderName := folder
if len(folderName) > 35 {
folderName = "..." + folderName[len(folderName)-32:]
}

// Определяем иконку
icon := "📁 "
if folder == cfg.CurrentDir {
icon = "⭐ "
}

// Считаем файлы в папке (упрощенно)
fileCount := countFilesInFolder(folder)

idx := i
rightPane.AddItem(
fmt.Sprintf("%s[yellow]%s[-]", icon, folderName),
fmt.Sprintf("[gray]Файлов: %d | Нажмите Enter[-]", fileCount),
0,
func() {
showFolderMenu(idx)
},
)
}
}

// Разделитель
rightPane.AddItem("", "", 0, nil)
rightPane.AddItem("[yellow]📄 ОТСЛЕЖИВАЕМЫЕ РАСШИРЕНИЯ[-]", "", 0, nil)

// Список расширений
for i, ext := range cfg.Extensions {
if i < 8 { // Показываем первые 8
rightPane.AddItem(fmt.Sprintf("  [cyan]%s[-]", ext), "", 0, nil)
} else if i == 8 {
rightPane.AddItem("  [gray]... и еще[-]", fmt.Sprintf("%d расширений", len(cfg.Extensions)-8), 0, nil)
break
}
}
}

func updateStatusBar() {
timeStr := time.Now().Format("15:04:05")
helpText := "[F1]Помощь [F4]Добавить [F5]Мониторинг [F7]Сканировать [F8]Удалить [F10]Выход"

if currentPanel == "right" {
helpText = "[Tab]Панели [Enter]Выбрать [F4]Добавить [F8]Удалить [F9]Настройки [Ctrl+Q]Выход"
}

statusBar.SetText(fmt.Sprintf("[white]%s | %s[-]", timeStr, helpText))
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
case tcell.KeyF3:
showLogViewer()
return nil
case tcell.KeyF4:
showAddFolderDialog()
return nil
case tcell.KeyF5:
toggleMonitoring()
return nil
case tcell.KeyF7:
quickScan()
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
executeSelected()
return nil
case tcell.KeyCtrlQ:
app.Stop()
return nil
}
return event
})
}

func togglePanel() {
if currentPanel == "left" {
currentPanel = "right"
app.SetFocus(rightPane)
rightPane.SetTitle(" [yellow]📁 ПАПКИ МОНИТОРИНГА[-] [green]◄[-] ")
leftPane.SetTitle(" [yellow]🖥 КОМАНДЫ[-] ")
} else {
currentPanel = "left"
app.SetFocus(leftPane)
leftPane.SetTitle(" [yellow]🖥 КОМАНДЫ[-] [green]◄[-] ")
rightPane.SetTitle(" [yellow]📁 ПАПКИ МОНИТОРИНГА[-] ")
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

// === ОСНОВНЫЕ ФУНКЦИИ ===

func toggleMonitoring() {
cfg.Monitoring = !cfg.Monitoring

if cfg.Monitoring {
addLog("[green]Мониторинг запущен[-]")
} else {
addLog("[yellow]Мониторинг остановлен[-]")
}

updateLeftPane()
updateTopPanel(nil)
}

func showAddFolderDialog() {
// Простой диалог добавления папки
modal := tview.NewModal().
SetText("Добавить папку для мониторинга\n\nИспользуется текущая папка").
AddButtons([]string{"Добавить текущую", "Добавить другую", "Отмена"}).
SetDoneFunc(func(buttonIndex int, buttonLabel string) {
switch buttonLabel {
case "Добавить текущую":
addFolder(cfg.CurrentDir)
case "Добавить другую":
addFolder("C:\\Example\\Path") // В реальности здесь был бы диалог выбора
addLog("[yellow]Для выбора папки требуется реализация диалога[-]")
}
})

modal.SetBorder(true).
SetTitle(" [yellow]📁 ДОБАВЛЕНИЕ ПАПКИ[-] ")

// Временно показываем модальное окно вместо полноценного диалога
app.SetRoot(modal, true)
}

func addFolder(path string) {
// Проверяем, нет ли уже такой папки
for _, folder := range cfg.Folders {
if folder == path {
addLog("[yellow]Папка уже добавлена[-]")
return
}
}

cfg.Folders = append(cfg.Folders, path)
updateRightPane()
updateTopPanel(nil)

addLog(fmt.Sprintf("[green]Добавлена папка: %s[-]", path))

// Возвращаемся к основному интерфейсу
app.SetRoot(createUI(), true)
}

func deleteSelectedFolder() {
if len(cfg.Folders) == 0 {
addLog("[red]Нет папок для удаления[-]")
return
}

modal := tview.NewModal().
SetText("Удалить выбранную папку из мониторинга?").
AddButtons([]string{"Да", "Нет"}).
SetDoneFunc(func(buttonIndex int, buttonLabel string) {
if buttonLabel == "Да" && len(cfg.Folders) > 0 {
removed := cfg.Folders[0]
cfg.Folders = cfg.Folders[1:]
updateRightPane()
updateTopPanel(nil)
addLog(fmt.Sprintf("[yellow]Удалена папка: %s[-]", removed))
}
app.SetRoot(createUI(), true)
})

modal.SetBorder(true).
SetTitle(" [yellow]🗑 УДАЛЕНИЕ ПАПКИ[-] ")

app.SetRoot(modal, true)
}

func showFolderMenu(index int) {
if index < len(cfg.Folders) {
folder := cfg.Folders[index]

modal := tview.NewModal().
SetText(fmt.Sprintf("Папка: [yellow]%s[-]\n\nВыберите действие:", folder)).
AddButtons([]string{"Сканировать", "Сделать основной", "Удалить", "Отмена"}).
SetDoneFunc(func(buttonIndex int, buttonLabel string) {
switch buttonLabel {
case "Сканировать":
count := countFilesInFolder(folder)
addLog(fmt.Sprintf("[cyan]Сканирование %s: %d файлов[-]", folder, count))
case "Сделать основной":
cfg.CurrentDir = folder
updateRightPane()
addLog(fmt.Sprintf("[green]Основная папка: %s[-]", folder))
case "Удалить":
if index < len(cfg.Folders) {
removed := cfg.Folders[index]
cfg.Folders = append(cfg.Folders[:index], cfg.Folders[index+1:]...)
updateRightPane()
updateTopPanel(nil)
addLog(fmt.Sprintf("[yellow]Удалена папка: %s[-]", removed))
}
}
app.SetRoot(createUI(), true)
})

modal.SetBorder(true).
SetTitle(" [yellow]📂 УПРАВЛЕНИЕ ПАПКОЙ[-] ")

app.SetRoot(modal, true)
}
}

func quickScan() {
addLog("[yellow]Запуск быстрого сканирования...[-]")

total := 0
for _, folder := range cfg.Folders {
count := countFilesInFolder(folder)
total += count
addLog(fmt.Sprintf("[gray]%s: %d файлов[-]", folder, count))
}

cfg.TotalFiles = total
updateTopPanel(nil)

addLog(fmt.Sprintf("[green]Сканирование завершено: %d файлов[-]", total))
}

func showStatistics() {
modal := tview.NewModal().
SetText(fmt.Sprintf("[yellow]📊 СТАТИСТИКА AILAN ARCHIVIST[-]\n\n"+
"[white]Отслеживаемых папок:[-] [cyan]%d[-]\n"+
"[white]Всего файлов:[-] [cyan]%d[-]\n"+
"[white]Отслеживаемых расширений:[-] [cyan]%d[-]\n\n"+
"[white]Мониторинг:[-] %s", 
len(cfg.Folders), cfg.TotalFiles, len(cfg.Extensions),
func() string {
if cfg.Monitoring {
return "[green]🟢 АКТИВЕН[-]"
}
return "[red]🔴 ОСТАНОВЛЕН[-]"
}())).
AddButtons([]string{"Закрыть", "Экспорт"}).
SetDoneFunc(func(buttonIndex int, buttonLabel string) {
if buttonLabel == "Экспорт" {
addLog("[green]Статистика экспортирована[-]")
}
app.SetRoot(createUI(), true)
})

modal.SetBorder(true).
SetTitle(" [yellow]📈 СТАТИСТИКА[-] ")

app.SetRoot(modal, true)
}

func showLogViewer() {
modal := tview.NewModal().
SetText("Просмотр логов\n\nЛоги сохраняются в папке docs/changelog").
AddButtons([]string{"Обновить", "Очистить", "Закрыть"}).
SetDoneFunc(func(buttonIndex int, buttonLabel string) {
switch buttonLabel {
case "Обновить":
addLog("[cyan]Логи обновлены[-]")
case "Очистить":
logView.SetText("")
addLog("[yellow]Логи очищены[-]")
}
app.SetRoot(createUI(), true)
})

modal.SetBorder(true).
SetTitle(" [yellow]📝 ПРОСМОТР ЛОГОВ[-] ")

app.SetRoot(modal, true)
}

func showSettings() {
modal := tview.NewModal().
SetText("Настройки программы\n\n• Управление расширениями\n• Настройка интервалов\n• Цветовая схема\n• Автозапуск").
AddButtons([]string{"Расширения", "Сохранить", "Отмена"}).
SetDoneFunc(func(buttonIndex int, buttonLabel string) {
if buttonLabel == "Расширения" {
showExtensionManager()
} else if buttonLabel == "Сохранить" {
addLog("[green]Настройки сохранены[-]")
app.SetRoot(createUI(), true)
} else {
app.SetRoot(createUI(), true)
}
})

modal.SetBorder(true).
SetTitle(" [yellow]⚙ НАСТРОЙКИ[-] ")

app.SetRoot(modal, true)
}

func showExtensionManager() {
modal := tview.NewModal().
SetText(fmt.Sprintf("Управление расширениями\n\nТекущие: %v\n\nДобавить новое расширение (.txt):", cfg.Extensions)).
AddButtons([]string{".php", ".html", ".js", ".css", ".txt", ".json", "Отмена"}).
SetDoneFunc(func(buttonIndex int, buttonLabel string) {
if buttonLabel != "Отмена" {
// Добавляем расширение
cfg.Extensions = append(cfg.Extensions, buttonLabel)
addLog(fmt.Sprintf("[green]Добавлено расширение: %s[-]", buttonLabel))
}
showSettings() // Возвращаемся к настройкам
})

modal.SetBorder(true).
SetTitle(" [yellow]📄 УПРАВЛЕНИЕ РАСШИРЕНИЯМИ[-] ")

app.SetRoot(modal, true)
}

func showHelp() {
modal := tview.NewModal().
SetText("[yellow]🚀 AILAN ARCHIVIST - СПРАВКА[-]\n\n" +
"[cyan]ГОРЯЧИЕ КЛАВИШИ:[-]\n" +
"[white]F1[-] - Эта справка\n" +
"[white]F4[-] - Добавить папку\n" +
"[white]F5[-] - Вкл/Выкл мониторинг\n" +
"[white]F7[-] - Быстрое сканирование\n" +
"[white]F8[-] - Удалить папку\n" +
"[white]F9[-] - Настройки\n" +
"[white]F10[-] - Выход\n\n" +
"[cyan]УПРАВЛЕНИЕ:[-]\n" +
"[white]Tab[-] - Переключение панелей\n" +
"[white]Enter[-] - Выполнить команду\n" +
"[white]Мышь[-] - Полная поддержка").
AddButtons([]string{"Закрыть"}).
SetDoneFunc(func(buttonIndex int, buttonLabel string) {
app.SetRoot(createUI(), true)
})

modal.SetBorder(true).
SetTitle(" [yellow]❓ СПРАВКА[-] ")

app.SetRoot(modal, true)
}

// === ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ ===

func addLog(message string) {
currentText := logView.GetText(true)
timestamp := time.Now().Format("15:04:05")
logView.SetText(currentText + fmt.Sprintf("[gray]%s[-] %s\n", timestamp, message))
logView.ScrollToEnd()
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
