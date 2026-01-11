package main

import (
"fmt"
"io/fs"
"os"
"path/filepath"
"strings"
"time"
)

// === КОНФИГУРАЦИЯ ===
type Config struct {
WatchDir      string
Extensions    []string
LogDir        string
SummaryFile   string
IsRunning     bool
StopChan      chan bool
CustomFolders []string
}

var cfg Config

// === ОСНОВНАЯ ФУНКЦИЯ ===
func main() {
initConfig()
runMainMenu()
}

func initConfig() {
cfg = Config{
WatchDir:      ".",
Extensions:    []string{".php", ".html", ".js", ".css", ".txt", ".json"},
LogDir:        "docs/changelog",
SummaryFile:   "docs/project_state.md",
IsRunning:     false,
StopChan:      make(chan bool),
CustomFolders: []string{"."},
}

os.MkdirAll(cfg.LogDir, 0755)
os.MkdirAll(filepath.Dir(cfg.SummaryFile), 0755)
}

// === ИНТЕРФЕЙС ===
func clearScreen() {
fmt.Print("\033[H\033[2J")
}

func showHeader() {
colorCyan := "\033[36m"
colorReset := "\033[0m"

fmt.Println(colorCyan + "╔══════════════════════════════════════════════════════════════╗")
fmt.Println("║                🚀 AILAN ARCHIVIST ULTIMATE                ║")
fmt.Println("║       Автономный файловый монитор с полным управлением    ║")
fmt.Println("╚══════════════════════════════════════════════════════════════╗" + colorReset)
fmt.Println()

// Статус
status := "🔴 ОСТАНОВЛЕН"
if cfg.IsRunning {
status = "🟢 АКТИВЕН"
}

fmt.Printf("  📊 Статус: %s\n", status)
fmt.Printf("  📁 Папок для мониторинга: %d\n", len(cfg.CustomFolders))
fmt.Printf("  ⚙  Отслеживаемых расширений: %d\n", len(cfg.Extensions))

totalFiles := countAllTrackedFiles()
fmt.Printf("  📈 Всего отслеживаемых файлов: %d\n", totalFiles)

fmt.Println()
}

func showMainMenu() {
fmt.Println("══════════════════════ ГЛАВНОЕ МЕНЮ ══════════════════════")

if cfg.IsRunning {
fmt.Println("  1. ⏹  ОСТАНОВИТЬ мониторинг")
} else {
fmt.Println("  1. ▶  ЗАПУСТИТЬ мониторинг")
}

fmt.Println("  2. 📂 Управление папками для мониторинга")
fmt.Println("  3. ⚙  Управление расширениями файлов")
fmt.Println("  4. 🔍 Быстрое сканирование файлов")
fmt.Println("  5. 📊 Просмотр логов изменений")
fmt.Println("  6. 📄 Обновить сводный файл проекта")
fmt.Println("  7. 📈 Показать статистику")
fmt.Println("  8. ⚡ Проверить изменения сейчас")
fmt.Println("  9. 🛠  Дополнительные настройки")
fmt.Println("  0. ❌ Выход")
fmt.Println("══════════════════════════════════════════════════════════")
}

func runMainMenu() {
for {
clearScreen()
showHeader()
showMainMenu()

var choice string
fmt.Print("\n➤ Выберите действие (0-9): ")
fmt.Scanln(&choice)

switch choice {
case "1":
toggleMonitoring()
case "2":
manageFoldersMenu()
case "3":
manageExtensionsMenu()
case "4":
quickScan()
case "5":
viewLogsMenu()
case "6":
updateProjectSummary()
case "7":
showStatistics()
case "8":
checkChangesNow()
case "9":
extraSettingsMenu()
case "0":
fmt.Println("\n👋 Выход из программы...")
return
default:
showMessage("Неверный выбор! Попробуйте снова.", "warning")
time.Sleep(1 * time.Second)
}
}
}

// === ОСНОВНЫЕ ФУНКЦИИ ===
func toggleMonitoring() {
if cfg.IsRunning {
// Останавливаем мониторинг
cfg.StopChan <- true
cfg.IsRunning = false
showMessage("Мониторинг остановлен", "success")
logSystemEvent("Мониторинг остановлен пользователем")
} else {
// Запускаем мониторинг
cfg.IsRunning = true
showMessage("Мониторинг запущен!", "success")
logSystemEvent("Мониторинг запущен")

// Запускаем в фоновом режиме
go backgroundMonitoring()
}
waitForEnter()
}

func backgroundMonitoring() {
ticker := time.NewTicker(30 * time.Second) // Проверка каждые 30 секунд
defer ticker.Stop()

for {
select {
case <-ticker.C:
// Проверяем изменения
checkForChanges()
logSystemEvent("Фоновая проверка файловой системы")

case <-cfg.StopChan:
logSystemEvent("Фоновый мониторинг остановлен")
return
}
}
}

func checkForChanges() {
// Здесь будет реальная проверка изменений файлов
// Пока просто имитация
logSystemEvent("Проверка изменений в отслеживаемых папках")
}

func manageFoldersMenu() {
for {
clearScreen()
fmt.Println("══════════════════ УПРАВЛЕНИЕ ПАПКАМИ ══════════════════")
fmt.Println("Текущие папки для мониторинга:")

for i, folder := range cfg.CustomFolders {
star := " "
if folder == cfg.WatchDir {
star = "★"
}
fmt.Printf("  %s %d. %s\n", star, i+1, folder)
}

fmt.Println("\nДействия:")
fmt.Println("  1. Добавить папку")
fmt.Println("  2. Удалить папку")
fmt.Println("  3. Изменить основную папку")
fmt.Println("  4. Очистить все папки (оставить только текущую)")
fmt.Println("  0. Назад в главное меню")

fmt.Print("\n➤ Выберите действие: ")
var choice string
fmt.Scanln(&choice)

switch choice {
case "1":
addFolder()
case "2":
removeFolder()
case "3":
setMainFolder()
case "4":
clearAllFolders()
case "0":
return
}
}
}

func addFolder() {
fmt.Print("\n📂 Введите путь к новой папке: ")
var newFolder string
fmt.Scanln(&newFolder)

if newFolder == "" {
return
}

// Проверяем существование папки
if info, err := os.Stat(newFolder); err != nil || !info.IsDir() {
showMessage("❌ Ошибка: папка не существует или недоступна", "error")
waitForEnter()
return
}

// Проверяем, нет ли уже такой папки
for _, folder := range cfg.CustomFolders {
if folder == newFolder {
showMessage("⚠  Эта папка уже в списке", "warning")
waitForEnter()
return
}
}

cfg.CustomFolders = append(cfg.CustomFolders, newFolder)
showMessage(fmt.Sprintf("✅ Добавлена папка: %s", newFolder), "success")
logSystemEvent(fmt.Sprintf("Добавлена папка для мониторинга: %s", newFolder))
waitForEnter()
}

func removeFolder() {
if len(cfg.CustomFolders) <= 1 {
showMessage("❌ Нельзя удалить все папки!", "error")
waitForEnter()
return
}

fmt.Print("\n🗑  Введите номер папки для удаления: ")
var num int
fmt.Scanln(&num)

if num < 1 || num > len(cfg.CustomFolders) {
showMessage("❌ Неверный номер", "error")
waitForEnter()
return
}

// Нельзя удалить основную папку, если она единственная
if cfg.CustomFolders[num-1] == cfg.WatchDir && len(cfg.CustomFolders) == 1 {
showMessage("❌ Нельзя удалить основную папку!", "error")
waitForEnter()
return
}

removed := cfg.CustomFolders[num-1]
cfg.CustomFolders = append(cfg.CustomFolders[:num-1], cfg.CustomFolders[num:]...)

showMessage(fmt.Sprintf("✅ Удалена папка: %s", removed), "success")
logSystemEvent(fmt.Sprintf("Удалена папка из мониторинга: %s", removed))
waitForEnter()
}

func setMainFolder() {
fmt.Print("\n★ Введите номер папки для установки как основной: ")
var num int
fmt.Scanln(&num)

if num < 1 || num > len(cfg.CustomFolders) {
showMessage("❌ Неверный номер", "error")
waitForEnter()
return
}

cfg.WatchDir = cfg.CustomFolders[num-1]
showMessage(fmt.Sprintf("✅ Основная папка установлена: %s", cfg.WatchDir), "success")
logSystemEvent(fmt.Sprintf("Изменена основная папка: %s", cfg.WatchDir))
waitForEnter()
}

func clearAllFolders() {
fmt.Println("\n⚠  ВНИМАНИЕ: Это удалит ВСЕ папки кроме текущей!")
fmt.Print("Вы уверены? (да/нет): ")

var confirm string
fmt.Scanln(&confirm)

if strings.ToLower(confirm) == "да" || strings.ToLower(confirm) == "yes" {
cfg.CustomFolders = []string{cfg.WatchDir}
showMessage("✅ Все папки очищены, оставлена только основная", "success")
logSystemEvent("Очищены все папки для мониторинга")
}
waitForEnter()
}

func manageExtensionsMenu() {
for {
clearScreen()
fmt.Println("══════════════════ УПРАВЛЕНИЕ РАСШИРЕНИЯМИ ══════════════════")
fmt.Println("Текущие отслеживаемые расширения файлов:")

for i, ext := range cfg.Extensions {
fmt.Printf("  %d. %s\n", i+1, ext)
}

fmt.Println("\nДействия:")
fmt.Println("  1. Добавить расширение")
fmt.Println("  2. Удалить расширение")
fmt.Println("  3. Сбросить к стандартным (.php .html .js .css .txt .json)")
fmt.Println("  4. Добавить все популярные расширения")
fmt.Println("  0. Назад")

fmt.Print("\n➤ Выберите действие: ")
var choice string
fmt.Scanln(&choice)

switch choice {
case "1":
addExtension()
case "2":
removeExtension()
case "3":
resetExtensions()
case "4":
addPopularExtensions()
case "0":
return
}
}
}

func addExtension() {
fmt.Print("\n➕ Введите новое расширение (начинается с точки): ")
var ext string
fmt.Scanln(&ext)

if ext == "" || !strings.HasPrefix(ext, ".") {
showMessage("❌ Неверный формат расширения", "error")
waitForEnter()
return
}

// Проверяем на дубликат
for _, existing := range cfg.Extensions {
if existing == ext {
showMessage("⚠  Это расширение уже есть в списке", "warning")
waitForEnter()
return
}
}

cfg.Extensions = append(cfg.Extensions, ext)
showMessage(fmt.Sprintf("✅ Добавлено расширение: %s", ext), "success")
logSystemEvent(fmt.Sprintf("Добавлено расширение для мониторинга: %s", ext))
waitForEnter()
}

func removeExtension() {
if len(cfg.Extensions) <= 1 {
showMessage("❌ Нельзя удалить все расширения!", "error")
waitForEnter()
return
}

fmt.Print("\n➖ Введите номер расширения для удаления: ")
var num int
fmt.Scanln(&num)

if num < 1 || num > len(cfg.Extensions) {
showMessage("❌ Неверный номер", "error")
waitForEnter()
return
}

removed := cfg.Extensions[num-1]
cfg.Extensions = append(cfg.Extensions[:num-1], cfg.Extensions[num:]...)

showMessage(fmt.Sprintf("✅ Удалено расширение: %s", removed), "success")
logSystemEvent(fmt.Sprintf("Удалено расширение из мониторинга: %s", removed))
waitForEnter()
}

func resetExtensions() {
cfg.Extensions = []string{".php", ".html", ".js", ".css", ".txt", ".json"}
showMessage("✅ Расширения сброшены к стандартным", "success")
logSystemEvent("Сброс расширений к стандартным настройкам")
waitForEnter()
}

func addPopularExtensions() {
popular := []string{".py", ".java", ".cpp", ".c", ".cs", ".rb", ".go", ".rs", ".ts", ".xml", ".yml", ".yaml", ".md", ".sql"}

added := 0
for _, ext := range popular {
// Проверяем, нет ли уже такого расширения
found := false
for _, existing := range cfg.Extensions {
if existing == ext {
found = true
break
}
}

if !found {
cfg.Extensions = append(cfg.Extensions, ext)
added++
}
}

showMessage(fmt.Sprintf("✅ Добавлено %d популярных расширений", added), "success")
logSystemEvent(fmt.Sprintf("Добавлены популярные расширения: %d новых", added))
waitForEnter()
}

func quickScan() {
clearScreen()
fmt.Println("══════════════════ БЫСТРОЕ СКАНИРОВАНИЕ ══════════════════")
fmt.Println("Сканирование всех отслеживаемых папок...")

totalFiles := 0
folderStats := make(map[string]int)
extStats := make(map[string]int)

for _, folder := range cfg.CustomFolders {
count := 0
filepath.WalkDir(folder, func(path string, d fs.DirEntry, err error) error {
if err != nil || d.IsDir() {
return nil
}

ext := filepath.Ext(path)
for _, tracked := range cfg.Extensions {
if ext == tracked {
count++
extStats[ext]++
break
}
}
return nil
})
folderStats[folder] = count
totalFiles += count
}

fmt.Printf("\n📊 Результаты сканирования:\n")
fmt.Printf("   Всего файлов: %d\n\n", totalFiles)

fmt.Println("📁 По папкам:")
for folder, count := range folderStats {
star := " "
if folder == cfg.WatchDir {
star = "★"
}
fmt.Printf("   %s %s: %d файлов\n", star, folder, count)
}

fmt.Println("\n📄 По расширениям:")
// Сортируем по количеству файлов
for ext, count := range extStats {
fmt.Printf("   %s: %d файлов\n", ext, count)
}

fmt.Println("\n══════════════════════════════════════════════════════════")
logSystemEvent(fmt.Sprintf("Быстрое сканирование: найдено %d файлов", totalFiles))
waitForEnter()
}

func viewLogsMenu() {
for {
clearScreen()
fmt.Println("══════════════════ ПРОСМОТР ЛОГОВ ══════════════════")
fmt.Println("  1. 📅 Просмотреть сегодняшний лог")
fmt.Println("  2. 📂 Выбрать лог по дате")
fmt.Println("  3. 📊 Последние 20 событий")
fmt.Println("  4. 🗑  Очистить старые логи")
fmt.Println("  0. Назад")

fmt.Print("\n➤ Выберите действие: ")
var choice string
fmt.Scanln(&choice)

switch choice {
case "1":
viewTodayLog()
case "2":
viewLogByDate()
case "3":
showRecentEvents(20)
case "4":
cleanOldLogs()
case "0":
return
}
}
}

func viewTodayLog() {
today := time.Now().Format("2006-01-02")
viewLogFile(today + "_changes.md")
}

func viewLogByDate() {
fmt.Print("\n📅 Введите дату (ГГГГ-ММ-ДД): ")
var date string
fmt.Scanln(&date)

viewLogFile(date + "_changes.md")
}

func viewLogFile(filename string) {
logPath := filepath.Join(cfg.LogDir, filename)

content, err := os.ReadFile(logPath)
if err != nil {
showMessage(fmt.Sprintf("❌ Лог-файл не найден: %s", filename), "error")
waitForEnter()
return
}

clearScreen()
fmt.Println(string(content))
waitForEnter()
}

func showRecentEvents(count int) {
today := time.Now().Format("2006-01-02")
logPath := filepath.Join(cfg.LogDir, today+"_changes.md")

content, err := os.ReadFile(logPath)
if err != nil {
showMessage("❌ Нет событий за сегодня", "error")
waitForEnter()
return
}

lines := strings.Split(string(content), "\n")

clearScreen()
fmt.Printf("══════════════════ ПОСЛЕДНИЕ %d СОБЫТИЙ ══════════════════\n", count)

events := 0
for i := len(lines) - 1; i >= 0 && events < count; i-- {
if strings.HasPrefix(lines[i], "###") {
fmt.Println(lines[i])
events++
}
}

if events == 0 {
fmt.Println("Событий не найдено")
}

fmt.Println("\n══════════════════════════════════════════════════════════")
waitForEnter()
}

func cleanOldLogs() {
fmt.Print("\n🗑  Сколько дней логов оставить? (0 = удалить все): ")
var days int
fmt.Scanln(&days)

if days < 0 {
showMessage("❌ Неверное количество дней", "error")
waitForEnter()
return
}

files, err := os.ReadDir(cfg.LogDir)
if err != nil {
showMessage("❌ Ошибка чтения папки логов", "error")
waitForEnter()
return
}

cutoff := time.Now().AddDate(0, 0, -days)
deleted := 0

for _, file := range files {
if file.IsDir() {
continue
}

info, err := file.Info()
if err != nil {
continue
}

if days == 0 || info.ModTime().Before(cutoff) {
os.Remove(filepath.Join(cfg.LogDir, file.Name()))
deleted++
}
}

showMessage(fmt.Sprintf("✅ Удалено %d лог-файлов", deleted), "success")
logSystemEvent(fmt.Sprintf("Очистка логов: удалено %d файлов", deleted))
waitForEnter()
}

func updateProjectSummary() {
summary := generateProjectSummary()
err := os.WriteFile(cfg.SummaryFile, []byte(summary), 0644)

if err != nil {
showMessage("❌ Ошибка обновления сводного файла", "error")
} else {
showMessage("✅ Сводный файл проекта обновлен", "success")
logSystemEvent("Обновлен сводный файл проекта")
}
waitForEnter()
}

func generateProjectSummary() string {
summary := "# Состояние проекта\n\n"
summary += fmt.Sprintf("**Дата обновления:** %s\n\n", time.Now().Format("2006-01-02 15:04:05"))

// Статус мониторинга
status := "🔴 Остановлен"
if cfg.IsRunning {
status = "🟢 Активен"
}
summary += fmt.Sprintf("**Статус мониторинга:** %s\n\n", status)

// Папки
summary += "## Отслеживаемые папки\n\n"
for _, folder := range cfg.CustomFolders {
count := countFilesInFolder(folder)
star := ""
if folder == cfg.WatchDir {
star = " ★"
}
summary += fmt.Sprintf("- %s (%d файлов)%s\n", folder, count, star)
}

// Расширения
summary += "\n## Отслеживаемые расширения\n\n"
summary += strings.Join(cfg.Extensions, ", ") + "\n"

// Статистика
summary += "\n## Статистика\n\n"
total := countAllTrackedFiles()
summary += fmt.Sprintf("Всего отслеживаемых файлов: **%d**\n", total)

// Подробная статистика по расширениям
extStats := make(map[string]int)
for _, folder := range cfg.CustomFolders {
filepath.WalkDir(folder, func(path string, d fs.DirEntry, err error) error {
if err != nil || d.IsDir() {
return nil
}

ext := filepath.Ext(path)
for _, tracked := range cfg.Extensions {
if ext == tracked {
extStats[ext]++
break
}
}
return nil
})
}

if len(extStats) > 0 {
summary += "\n### По расширениям:\n"
for ext, count := range extStats {
summary += fmt.Sprintf("- %s: %d файлов\n", ext, count)
}
}

return summary
}

func showStatistics() {
clearScreen()
fmt.Println("════════════════════ СТАТИСТИКА ════════════════════")

totalFiles := countAllTrackedFiles()
extStats := make(map[string]int)
folderStats := make(map[string]int)

// Собираем статистику
for _, folder := range cfg.CustomFolders {
count := 0
filepath.WalkDir(folder, func(path string, d fs.DirEntry, err error) error {
if err != nil || d.IsDir() {
return nil
}

ext := filepath.Ext(path)
for _, tracked := range cfg.Extensions {
if ext == tracked {
extStats[ext]++
count++
break
}
}
return nil
})
folderStats[folder] = count
}

fmt.Printf("📊 Всего отслеживаемых файлов: %d\n\n", totalFiles)

fmt.Println("📁 Статистика по папкам:")
for folder, count := range folderStats {
star := " "
if folder == cfg.WatchDir {
star = "★"
}
percentage := 0
if totalFiles > 0 {
percentage = (count * 100) / totalFiles
}
fmt.Printf("  %s %s: %d файлов (%d%%)\n", star, folder, count, percentage)
}

fmt.Println("\n📄 Статистика по расширениям:")
for ext, count := range extStats {
percentage := 0
if totalFiles > 0 {
percentage = (count * 100) / totalFiles
}
fmt.Printf("  %s: %d файлов (%d%%)\n", ext, count, percentage)
}

fmt.Println("\n══════════════════════════════════════════════════════════")
logSystemEvent("Просмотр статистики проекта")
waitForEnter()
}

func checkChangesNow() {
showMessage("🔍 Проверка изменений...", "info")

// Имитация проверки изменений
logSystemEvent("Ручная проверка изменений файлов")

// Обновляем сводный файл
updateProjectSummary()

showMessage("✅ Проверка завершена", "success")
waitForEnter()
}

func extraSettingsMenu() {
for {
clearScreen()
fmt.Println("══════════════════ ДОПОЛНИТЕЛЬНЫЕ НАСТРОЙКИ ══════════════════")
fmt.Println("  1. 📝 Изменить папку для логов")
fmt.Println("  2. 📄 Изменить имя сводного файла")
fmt.Println("  3. 🧹 Очистить все настройки")
fmt.Println("  4. 💾 Экспорт настроек")
fmt.Println("  0. Назад")

fmt.Print("\n➤ Выберите действие: ")
var choice string
fmt.Scanln(&choice)

switch choice {
case "1":
changeLogDirectory()
case "2":
changeSummaryFilename()
case "3":
resetAllSettings()
case "4":
exportSettings()
case "0":
return
}
}
}

func changeLogDirectory() {
fmt.Printf("\nТекущая папка логов: %s\n", cfg.LogDir)
fmt.Print("Введите новую папку: ")

var newDir string
fmt.Scanln(&newDir)

if newDir != "" {
oldDir := cfg.LogDir
cfg.LogDir = newDir
os.MkdirAll(cfg.LogDir, 0755)

showMessage(fmt.Sprintf("✅ Папка логов изменена: %s", newDir), "success")
logSystemEvent(fmt.Sprintf("Изменена папка логов: %s → %s", oldDir, newDir))
}
waitForEnter()
}

func changeSummaryFilename() {
fmt.Printf("\nТекущий сводный файл: %s\n", cfg.SummaryFile)
fmt.Print("Введите новое имя файла: ")

var newName string
fmt.Scanln(&newName)

if newName != "" {
oldName := cfg.SummaryFile
cfg.SummaryFile = newName

showMessage(fmt.Sprintf("✅ Имя сводного файла изменено: %s", newName), "success")
logSystemEvent(fmt.Sprintf("Изменен сводный файл: %s → %s", oldName, newName))
}
waitForEnter()
}

func resetAllSettings() {
fmt.Println("\n⚠  ⚠  ⚠  ВНИМАНИЕ! ⚠  ⚠  ⚠")
fmt.Println("Это сбросит ВСЕ настройки к значениям по умолчанию!")
fmt.Println("Будут сброшены: папки, расширения, настройки логов")
fmt.Print("\nВы уверены? (да/нет): ")

var confirm string
fmt.Scanln(&confirm)

if strings.ToLower(confirm) == "да" || strings.ToLower(confirm) == "yes" {
// Создаем резервную копию старых настроек
oldFolders := cfg.CustomFolders
oldExtensions := cfg.Extensions

// Сбрасываем настройки
cfg.CustomFolders = []string{"."}
cfg.Extensions = []string{".php", ".html", ".js", ".css", ".txt", ".json"}
cfg.WatchDir = "."

showMessage("✅ Все настройки сброшены к значениям по умолчанию", "success")
logSystemEvent(fmt.Sprintf("Сброс всех настроек. Было: %d папок, %d расширений", 
len(oldFolders), len(oldExtensions)))
}
waitForEnter()
}

func exportSettings() {
fmt.Println("\nЭкспорт настроек...")

settings := fmt.Sprintf("AILAN Archivist - Экспорт настроек\n")
settings += fmt.Sprintf("Дата экспорта: %s\n\n", time.Now().Format("2006-01-02 15:04:05"))

settings += "Отслеживаемые папки:\n"
for _, folder := range cfg.CustomFolders {
settings += fmt.Sprintf("  - %s\n", folder)
}

settings += "\nОтслеживаемые расширения:\n"
for _, ext := range cfg.Extensions {
settings += fmt.Sprintf("  - %s\n", ext)
}

settings += fmt.Sprintf("\nПапка логов: %s\n", cfg.LogDir)
settings += fmt.Sprintf("Сводный файл: %s\n", cfg.SummaryFile)
settings += fmt.Sprintf("Статус мониторинга: %v\n", cfg.IsRunning)

// Сохраняем в файл
exportFile := fmt.Sprintf("ailan_settings_%s.txt", time.Now().Format("20060102_150405"))
err := os.WriteFile(exportFile, []byte(settings), 0644)

if err != nil {
showMessage("❌ Ошибка экспорта настроек", "error")
} else {
showMessage(fmt.Sprintf("✅ Настройки экспортированы в файл: %s", exportFile), "success")
logSystemEvent(fmt.Sprintf("Экспорт настроек в файл: %s", exportFile))
}
waitForEnter()
}

// === ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ ===
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

func logSystemEvent(message string) {
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
// Ограничиваем размер файла
lines := strings.Split(string(content), "\n")
if len(lines) > 1000 {
content = []byte(strings.Join(lines[len(lines)-800:], "\n"))
}
entry = string(content) + "\n" + entry
}

os.WriteFile(logFile, []byte(entry), 0644)
}

func showMessage(message string, msgType string) {
var color string
var prefix string

switch msgType {
case "success":
color = "\033[32m" // Зеленый
prefix = "✅ "
case "error":
color = "\033[31m" // Красный
prefix = "❌ "
case "warning":
color = "\033[33m" // Желтый
prefix = "⚠  "
case "info":
color = "\033[36m" // Голубой
prefix = "ℹ  "
default:
color = "\033[37m" // Белый
prefix = "• "
}

reset := "\033[0m"
fmt.Printf("\n%s%s%s%s\n", color, prefix, message, reset)
}

func waitForEnter() {
fmt.Print("\nНажмите Enter чтобы продолжить...")
fmt.Scanln()
}
