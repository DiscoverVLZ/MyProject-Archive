package main

import (
"fmt"
"os"
"path/filepath"
"strings"
"time"

"github.com/gdamore/tcell/v2"
"github.com/rivo/tview"
)

var (
app       *tview.Application
pages     *tview.Pages
menuBar   *tview.TextView
mainArea  *tview.Flex
statusBar *tview.TextView
)

func main() {
app = tview.NewApplication()
pages = tview.NewPages()

// Создаем главный интерфейс
createMainUI()

// Добавляем горячие клавиши
setupHotkeys()

if err := app.SetRoot(pages, true).EnableMouse(true).Run(); err != nil {
panic(err)
}
}

func createMainUI() {
// === МЕНЮ БАР (как в Windows) ===
menuBar = tview.NewTextView().
SetDynamicColors(true).
SetRegions(true)

updateMenuBar()

// === ОСНОВНАЯ ОБЛАСТЬ ===
mainArea = tview.NewFlex().SetDirection(tview.FlexRow)

// Создаем начальный экран
createWelcomeScreen()

// === СТАТУС БАР ===
statusBar = tview.NewTextView().
SetDynamicColors(true).
SetRegions(true)

updateStatusBar("Готов к работе")

// === ГЛАВНЫЙ КОНТЕЙНЕР ===
mainContainer := tview.NewFlex().SetDirection(tview.FlexRow)
mainContainer.AddItem(menuBar, 1, 1, false)
mainContainer.AddItem(mainArea, 0, 4, true)
mainContainer.AddItem(statusBar, 1, 1, false)

pages.AddPage("main", mainContainer, true, true)
}

func updateMenuBar() {
menuText := `[white][#0080FF]Файл [#0000AA]|[-][#0080FF] Правка [#0000AA]|[-][#0080FF] Вид [#0000AA]|[-][#0080FF] Мониторинг [#0000AA]|[-][#0080FF] Сервис [#0000AA]|[-][#0080FF] Настройки [#0000AA]|[-][#0080FF] Справка[-]`
menuBar.SetText(menuText)
menuBar.SetBackgroundColor(tcell.ColorBlack)
}

func createWelcomeScreen() {
mainArea.Clear()

// Создаем красивый приветственный экран
welcome := tview.NewTextView().
SetDynamicColors(true).
SetTextAlign(tview.AlignCenter)

welcomeText := `
[#0080FF]╔══════════════════════════════════════════════════════════════╗[-]
[#0080FF]║[-]          [#FFFF00]🚀 AILAN ARCHIVIST - WINDOWS STYLE[-]          [#0080FF]║[-]
[#0080FF]║[-]     Автономный файловый монитор с графическим интерфейсом [#0080FF]║[-]
[#0080FF]╚══════════════════════════════════════════════════════════════╝[-]

[#00FF00]════════════════════ ОСНОВНЫЕ ВОЗМОЖНОСТИ ════════════════════[-]

[white]• 📁 Управление мышью и клавиатурой как в Windows[-]
[white]• 🖱  Полная поддержка мыши (клики, выделение)[-]
[white]• 📊 Две панели в стиле Total Commander[-]
[white]• ⚙  Графические диалоговые окна[-]
[white]• 📈 Реальная статистика в реальном времени[-]
[white]• 🚀 Автономная работа с флешки[-]

[#00FF00]══════════════════════ БЫСТРЫЙ СТАРТ ══════════════════════[-]

[yellow]F2[-] - Добавить папку     [yellow]F3[-] - Удалить папку
[yellow]F5[-] - Запустить мониторинг   [yellow]F6[-] - Остановить мониторинг
[yellow]F7[-] - Просмотреть логи   [yellow]F8[-] - Статистика
[yellow]F9[-] - Настройки          [yellow]F10[-] - Выход

[#0080FF]══════════════════════════════════════════════════════════════[-]
`
welcome.SetText(welcomeText)

// Добавляем кнопки быстрого запуска
buttons := tview.NewFlex().SetDirection(tview.FlexColumn)

buttonStyle := tcell.StyleDefault.
Background(tcell.ColorDarkBlue).
Foreground(tcell.ColorWhite)

addButton := func(label string, action func()) *tview.Button {
btn := tview.NewButton(label)
btn.SetStyle(buttonStyle)
btn.SetSelectedFunc(action)
return btn
}

buttons.AddItem(addButton(" 📁 УПРАВЛЕНИЕ ПАПКАМИ ", showFolderManager), 0, 1, false)
buttons.AddItem(addButton(" ▶ ЗАПУСТИТЬ МОНИТОРИНГ ", startMonitoring), 0, 1, false)
buttons.AddItem(addButton(" 📊 ПРОСМОТР ЛОГОВ ", showLogs), 0, 1, false)
buttons.AddItem(addButton(" ⚙ НАСТРОЙКИ ", showSettings), 0, 1, false)

mainArea.AddItem(welcome, 0, 3, false)
mainArea.AddItem(buttons, 3, 1, true)
}

func showFolderManager() {
// Создаем диалог в стиле Windows
dialog := tview.NewFlex().SetDirection(tview.FlexRow)
dialog.SetBorder(true).SetTitle(" 📁 УПРАВЛЕНИЕ ПАПКАМИ ")

// Список папок
list := tview.NewList().
ShowSecondaryText(false).
SetHighlightFullLine(true)

list.AddItem("C:\\Projects", "Основная папка проекта", 'C', nil)
list.AddItem("D:\\Web", "Веб-разработка", 'W', nil)
list.AddItem("E:\\Backup", "Резервные копии", 'B', nil)

// Кнопки
buttonRow := tview.NewFlex().SetDirection(tview.FlexColumn)

buttons := []struct {
label  string
action func()
}{
{"Добавить", addFolder},
{"Удалить", deleteFolder},
{"Изменить", editFolder},
{"Закрыть", func() { pages.HidePage("folderManager") }},
}

for _, btn := range buttons {
button := tview.NewButton(btn.label)
button.SetSelectedFunc(btn.action)
buttonRow.AddItem(button, 0, 1, false)
}

dialog.AddItem(list, 0, 1, true)
dialog.AddItem(buttonRow, 1, 1, false)

center := tview.NewCenter(dialog, 60, 20)
pages.AddPage("folderManager", center, true, true)
}

func startMonitoring() {
updateStatusBar("Мониторинг запущен...")

// Показываем индикатор прогресса
showProgressDialog("Запуск мониторинга", "Сканирование файловой системы...", 100)
}

func showLogs() {
// Создаем окно просмотра логов
textView := tview.NewTextView().
SetDynamicColors(true).
SetScrollable(true)

textView.SetBorder(true).SetTitle(" 📝 ЖУРНАЛ СОБЫТИЙ ")

// Заполняем тестовыми данными
logContent := `[gray]10:15:23[-] [green]Мониторинг запущен[-]
[gray]10:15:45[-] [yellow]Добавлена папка: C:\Projects[-]
[gray]10:16:10[-] [yellow]Найдено 156 файлов для отслеживания[-]
[gray]10:17:30[-] [cyan]Обнаружено изменение: index.html[-]
[gray]10:18:15[-] [cyan]Обнаружено изменение: styles.css[-]
[gray]10:19:00[-] [red]Удален файл: old_script.js[-]
[gray]10:20:45[-] [green]Создан новый файл: app.js[-]
[gray]10:21:30[-] [yellow]Обновлен сводный файл проекта[-]
[gray]10:22:00[-] [cyan]Обнаружено изменение: config.json[-]`

textView.SetText(logContent)

// Кнопки управления
buttonRow := tview.NewFlex().SetDirection(tview.FlexColumn)
buttonRow.AddItem(tview.NewButton("Обновить").SetSelectedFunc(func() {
// Обновление логов
}), 0, 1, false)
buttonRow.AddItem(tview.NewButton("Очистить").SetSelectedFunc(func() {
textView.SetText("")
}), 0, 1, false)
buttonRow.AddItem(tview.NewButton("Экспорт").SetSelectedFunc(func() {
// Экспорт логов
}), 0, 1, false)
buttonRow.AddItem(tview.NewButton("Закрыть").SetSelectedFunc(func() {
pages.HidePage("logs")
}), 0, 1, false)

dialog := tview.NewFlex().SetDirection(tview.FlexRow)
dialog.AddItem(textView, 0, 1, true)
dialog.AddItem(buttonRow, 1, 1, false)
dialog.SetBorder(true)

center := tview.NewCenter(dialog, 80, 25)
pages.AddPage("logs", center, true, true)
}

func showSettings() {
// Создаем форму настроек в стиле Windows
form := tview.NewForm()

form.AddDropDown("Тема интерфейса:", []string{"Синяя", "Зеленая", "Темная", "Классическая"}, 0, nil)
form.AddInputField("Папка для логов:", "docs\\changelog", 30, nil, nil)
form.AddInputField("Сводный файл:", "docs\\project_state.md", 30, nil, nil)
form.AddCheckbox("Автозапуск мониторинга", false, nil)
form.AddCheckbox("Звуковые уведомления", true, nil)
form.AddCheckbox("Показывать скрытые файлы", false, nil)
form.AddInputField("Интервал проверки (сек):", "30", 10, nil, nil)

form.AddButton("Сохранить", func() {
updateStatusBar("Настройки сохранены")
pages.HidePage("settings")
})
form.AddButton("Отмена", func() {
pages.HidePage("settings")
})

form.SetBorder(true).SetTitle(" ⚙ НАСТРОЙКИ ")
center := tview.NewCenter(form, 60, 20)
pages.AddPage("settings", center, true, true)
}

func showProgressDialog(title, message string, total int) {
modal := tview.NewModal().
SetText(fmt.Sprintf("%s\n\n%s", title, message)).
AddButtons([]string{"Отмена"})

progressBar := tview.NewTextView()

// Имитация прогресса
go func() {
for i := 0; i <= total; i++ {
time.Sleep(50 * time.Millisecond)
percent := (i * 100) / total
progress := strings.Repeat("█", percent/2) + strings.Repeat("░", 50-percent/2)
progressBar.SetText(fmt.Sprintf("[%s] %d%%", progress, percent))
app.Draw()

if i == total {
time.Sleep(500 * time.Millisecond)
pages.HidePage("progress")
updateStatusBar("Мониторинг успешно запущен")
break
}
}
}()

flex := tview.NewFlex().SetDirection(tview.FlexRow)
flex.AddItem(modal, 10, 1, true)
flex.AddItem(progressBar, 1, 1, false)

center := tview.NewCenter(flex, 60, 15)
pages.AddPage("progress", center, true, true)
}

func addFolder() {
// Диалог добавления папки
inputField := tview.NewInputField().
SetLabel("Путь к папке: ").
SetFieldWidth(40)

form := tview.NewForm().
AddFormItem(inputField).
AddButton("Выбрать", func() {
// Здесь была бы логика выбора папки
pages.HidePage("addFolder")
}).
AddButton("Отмена", func() {
pages.HidePage("addFolder")
})

form.SetBorder(true).SetTitle(" 📁 ДОБАВИТЬ ПАПКУ ")
center := tview.NewCenter(form, 60, 10)
pages.AddPage("addFolder", center, true, true)
}

func deleteFolder() {
modal := tview.NewModal().
SetText("Вы уверены, что хотите удалить выбранную папку?\n\nЭто действие нельзя отменить.").
AddButtons([]string{"Да, удалить", "Нет, отменить"}).
SetDoneFunc(func(buttonIndex int, buttonLabel string) {
if buttonLabel == "Да, удалить" {
updateStatusBar("Папка удалена")
}
pages.HidePage("deleteConfirm")
})

modal.SetTitle(" 🗑 ПОДТВЕРЖДЕНИЕ УДАЛЕНИЯ ")
pages.AddPage("deleteConfirm", modal, true, true)
}

func editFolder() {
updateStatusBar("Редактирование папки...")
// Реализация редактирования
}

func setupHotkeys() {
app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
switch event.Key() {
case tcell.KeyF1:
showHelp()
return nil
case tcell.KeyF2:
addFolder()
return nil
case tcell.KeyF3:
deleteFolder()
return nil
case tcell.KeyF5:
startMonitoring()
return nil
case tcell.KeyF6:
updateStatusBar("Мониторинг остановлен")
return nil
case tcell.KeyF7:
showLogs()
return nil
case tcell.KeyF8:
showStatistics()
return nil
case tcell.KeyF9:
showSettings()
return nil
case tcell.KeyF10:
app.Stop()
return nil
case tcell.KeyCtrlQ:
app.Stop()
return nil
}

// Обработка Alt+буква для меню
if event.Modifiers() == tcell.ModAlt {
switch event.Rune() {
case 'ф', 'Ф', 'a', 'A': // Файл
showFileMenu()
return nil
case 'п', 'П', 'e', 'E': // Правка
showEditMenu()
return nil
case 'в', 'В', 'v', 'V': // Вид
showViewMenu()
return nil
case 'с', 'С', 'h', 'H': // Справка
showHelp()
return nil
}
}

return event
})
}

func showFileMenu() {
// Выпадающее меню Файл
showDropdownMenu("Файл", []string{
"Новый проект",
"Открыть проект",
"Сохранить настройки",
"Экспорт логов",
"Выход",
})
}

func showEditMenu() {
showDropdownMenu("Правка", []string{
"Добавить папку",
"Удалить папку",
"Копировать",
"Вставить",
"Найти",
})
}

func showViewMenu() {
showDropdownMenu("Вид", []string{
"Две панели",
"Полный экран",
"Только логи",
"Статистика",
"Обновить",
})
}

func showDropdownMenu(title string, items []string) {
list := tview.NewList()

for _, item := range items {
list.AddItem(item, "", 0, nil)
}

list.SetBorder(true).SetTitle(" " + title + " ")

// Позиционируем меню под соответствующей кнопкой в меню баре
center := tview.NewCenter(list, 20, len(items)+2)
pages.AddPage("dropdownMenu", center, true, true)

// Автоматически закрываем меню через 5 секунд или при выборе
go func() {
time.Sleep(5 * time.Second)
app.QueueUpdateDraw(func() {
if pages.HasPage("dropdownMenu") {
pages.HidePage("dropdownMenu")
}
})
}()
}

func showStatistics() {
modal := tview.NewModal().
SetText("[::b]📊 СТАТИСТИКА ПРОЕКТА[::-]\n\n" +
"[yellow]Отслеживаемых файлов:[-] 1,248\n" +
"[yellow]Папок для мониторинга:[-] 5\n" +
"[yellow]Изменений за сегодня:[-] 42\n" +
"[yellow]Всего лог-файлов:[-] 15\n" +
"[yellow]Размер логов:[-] 2.4 МБ\n\n" +
"[green]🟢 МОНИТОРИНГ АКТИВЕН[-]").
AddButtons([]string{"Обновить", "Экспорт", "Закрыть"}).
SetDoneFunc(func(buttonIndex int, buttonLabel string) {
if buttonLabel == "Обновить" {
showStatistics() // Обновляем
} else if buttonLabel == "Экспорт" {
updateStatusBar("Статистика экспортирована")
}
pages.HidePage("statistics")
})

modal.SetTitle(" 📈 СТАТИСТИКА ")
pages.AddPage("statistics", modal, true, true)
}

func showHelp() {
textView := tview.NewTextView().
SetDynamicColors(true).
SetScrollable(true)

helpText := `[::b]🚀 AILAN ARCHIVIST - СПРАВКА[::-]

[yellow]НАЗНАЧЕНИЕ:[-]
Автономный файловый монитор для локальной разработки.
Отслеживает изменения файлов и автоматически создает лог.

[yellow]ГОРЯЧИЕ КЛАВИШИ:[-]
• F2 - Добавить папку для мониторинга
• F3 - Удалить выбранную папку
• F5 - Запустить/остановить мониторинг
• F7 - Просмотр логов изменений
• F8 - Статистика проекта
• F9 - Настройки программы
• F10 - Выход из программы
• Alt+Ф - Меню "Файл"
• Alt+П - Меню "Правка"
• Alt+В - Меню "Вид"

[yellow]УПРАВЛЕНИЕ МЫШЬЮ:[-]
• Левый клик - выбор элемента
• Двойной клик - выполнение действия
• Правый клик - контекстное меню

[yellow]ОСНОВНЫЕ ФУНКЦИИ:[-]
1. Отслеживание изменений в реальном времени
2. Автоматическое логирование в формате Markdown
3. Управление несколькими папками одновременно
4. Статистика и отчеты по проекту
5. Работа без интернета и внешних зависимостей

[green]Версия: 5.0 (Windows Style GUI)[-]`

textView.SetText(helpText)
textView.SetBorder(true).SetTitle(" ❓ СПРАВКА ")

button := tview.NewButton("Закрыть")
button.SetSelectedFunc(func() {
pages.HidePage("help")
})

flex := tview.NewFlex().SetDirection(tview.FlexRow)
flex.AddItem(textView, 0, 1, true)
flex.AddItem(button, 1, 1, false)

center := tview.NewCenter(flex, 70, 25)
pages.AddPage("help", center, true, true)
}

func updateStatusBar(message string) {
timeStr := time.Now().Format("15:04:05")
statusText := fmt.Sprintf("[white]%s | %s | Готов[-]", timeStr, message)
statusBar.SetText(statusText)
}
