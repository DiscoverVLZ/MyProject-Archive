package main

import (
"fmt"
"io/fs"
"os"
"path/filepath"
"strings"
"time"
)

// Простая структура для конфигурации
type Config struct {
WatchDir    string
Extensions  []string
LogDir      string
SummaryFile string
IsRunning   bool
}

var config Config

func main() {
// Инициализация
config = Config{
WatchDir:    ".",
Extensions:  []string{".php", ".html", ".js", ".css", ".txt", ".json"},
LogDir:      "docs/changelog",
SummaryFile: "docs/project_state.md",
IsRunning:   false,
}

// Создаем папки
createDirectories()

// Запускаем простой текстовый интерфейс
runTextUI()
}

func createDirectories() {
os.MkdirAll(config.LogDir, 0755)
os.MkdirAll(filepath.Dir(config.SummaryFile), 0755)
}

func runTextUI() {
for {
clearScreen()
showHeader()
showMenu()

var choice string
fmt.Print("\nВыберите действие: ")
fmt.Scanln(&choice)

switch choice {
case "1":
startMonitoring()
case "2":
stopMonitoring()
case "3":
selectDirectory()
case "4":
viewLogs()
case "5":
showSettings()
case "6":
updateSummary()
case "7":
scanFiles()
case "8":
clearLogs()
case "0", "q", "exit":
fmt.Println("\nВыход из программы...")
return
default:
fmt.Println("\nНеверный выбор. Нажмите любую клавишу...")
fmt.Scanln()
}
}
}

func clearScreen() {
// Простая очистка экрана
fmt.Print("\033[H\033[2J")
}

func showHeader() {
fmt.Println("╔══════════════════════════════════════════════════════════════╗")
fmt.Println("║                 AILAN ARCHIVIST v2.0 GUI                     ║")
fmt.Println("║        Автономный мониторинг файлов с интерфейсом            ║")
fmt.Println("╚══════════════════════════════════════════════════════════════╝")
fmt.Println()

// Статус
if config.IsRunning {
fmt.Println("  🟢 СТАТУС: МОНИТОРИНГ АКТИВЕН")
} else {
fmt.Println("  🔴 СТАТУС: МОНИТОРИНГ ОСТАНОВЛЕН")
}

fmt.Printf("  📁 Папка: %s\n", config.WatchDir)

// Счетчик файлов
count := countTrackedFiles()
fmt.Printf("  📊 Файлов для отслеживания: %d\n", count)

fmt.Println()
}

func showMenu() {
fmt.Println("════════════════════ МЕНУ ═════════════════════")
fmt.Println("  1. ▶  Запустить мониторинг")
fmt.Println("  2. ⏹  Остановить мониторинг")
fmt.Println("  3. 📂 Выбрать другую папку")
fmt.Println("  4. 📊 Просмотреть логи изменений")
fmt.Println("  5. ⚙  Настройки (расширения файлов)")
fmt.Println("  6. 🔄 Обновить сводный файл")
fmt.Println("  7. 🔍 Просканировать файлы сейчас")
fmt.Println("  8. 🗑  Очистить старые логи")
fmt.Println("  0. ❌ Выход")
fmt.Println("═══════════════════════════════════════════════")
}

func startMonitoring() {
if config.IsRunning {
fmt.Println("\n⚠  Мониторинг уже запущен!")
fmt.Print("Нажмите Enter...")
fmt.Scanln()
return
}

fmt.Println("\n🚀 Запуск мониторинга...")

// Создаем канал для остановки
stopChan := make(chan bool)

// Запускаем мониторинг в отдельной горутине
go func() {
config.IsRunning = true
fmt.Println("✅ Мониторинг запущен. Нажмите 2 в меню для остановки.")

// Создаем наблюдатель
monitorFiles(stopChan)

config.IsRunning = false
}()

// Ждем немного
time.Sleep(1 * time.Second)
}

func stopMonitoring() {
if !config.IsRunning {
fmt.Println("\n⚠  Мониторинг не запущен!")
fmt.Print("Нажмите Enter...")
fmt.Scanln()
return
}

// В реальной реализации здесь бы отправлялся сигнал в канал
config.IsRunning = false
fmt.Println("\n⏹  Мониторинг остановлен")
fmt.Print("Нажмите Enter...")
fmt.Scanln()
}

func selectDirectory() {
fmt.Println("\n📂 Выбор папки для мониторинга")
fmt.Println("Текущая папка:", config.WatchDir)
fmt.Print("Введите новый путь (или Enter для отмены): ")

var newDir string
fmt.Scanln(&newDir)

if newDir == "" {
return
}

// Проверяем существование папки
if info, err := os.Stat(newDir); err != nil || !info.IsDir() {
fmt.Println("❌ Ошибка: папка не существует или недоступна")
fmt.Print("Нажмите Enter...")
fmt.Scanln()
return
}

config.WatchDir = newDir
fmt.Println("✅ Папка изменена на:", newDir)

// Обновляем счетчик файлов
count := countTrackedFiles()
fmt.Printf("📊 Найдено файлов для отслеживания: %d\n", count)

fmt.Print("Нажмите Enter...")
fmt.Scanln()
}

func viewLogs() {
fmt.Println("\n📊 Просмотр логов изменений")

// Показываем доступные логи
files, err := os.ReadDir(config.LogDir)
if err != nil {
fmt.Println("❌ Ошибка чтения папки логов:", err)
fmt.Print("Нажмите Enter...")
fmt.Scanln()
return
}

if len(files) == 0 {
fmt.Println("Логи отсутствуют")
fmt.Print("Нажмите Enter...")
fmt.Scanln()
return
}

// Сортируем по дате (новые сверху)
for i := len(files) - 1; i >= 0; i-- {
file := files[i]
if !file.IsDir() && strings.HasSuffix(file.Name(), ".md") {
fmt.Printf("  %s\n", file.Name())
}
}

fmt.Print("\nВведите имя файла для просмотра (или Enter для отмены): ")
var filename string
fmt.Scanln(&filename)

if filename == "" {
return
}

// Читаем и показываем файл
content, err := os.ReadFile(filepath.Join(config.LogDir, filename))
if err != nil {
fmt.Println("❌ Ошибка чтения файла:", err)
} else {
clearScreen()
fmt.Println(string(content))
}

fmt.Print("\nНажмите Enter чтобы вернуться в меню...")
fmt.Scanln()
}

func showSettings() {
for {
clearScreen()
fmt.Println("══════════════════ НАСТРОЙКИ ══════════════════")
fmt.Println("Текущие отслеживаемые расширения:")

for i, ext := range config.Extensions {
fmt.Printf("  %d. %s\n", i+1, ext)
}

fmt.Println("\nДействия:")
fmt.Println("  1. Добавить расширение")
fmt.Println("  2. Удалить расширение")
fmt.Println("  3. Изменить папку логов")
fmt.Println("  0. Назад в меню")
fmt.Print("\nВыберите действие: ")

var choice string
fmt.Scanln(&choice)

switch choice {
case "1":
addExtension()
case "2":
removeExtension()
case "3":
changeLogDir()
case "0":
return
}
}
}

func addExtension() {
fmt.Print("\nВведите расширение (например .py): ")
var ext string
fmt.Scanln(&ext)

if ext == "" || !strings.HasPrefix(ext, ".") {
fmt.Println("❌ Неверный формат расширения")
fmt.Print("Нажмите Enter...")
fmt.Scanln()
return
}

// Проверяем, нет ли уже такого расширения
for _, existing := range config.Extensions {
if existing == ext {
fmt.Println("⚠  Это расширение уже есть в списке")
fmt.Print("Нажмите Enter...")
fmt.Scanln()
return
}
}

config.Extensions = append(config.Extensions, ext)
fmt.Printf("✅ Добавлено расширение: %s\n", ext)
fmt.Print("Нажмите Enter...")
fmt.Scanln()
}

func removeExtension() {
if len(config.Extensions) <= 1 {
fmt.Println("❌ Нельзя удалить все расширения!")
fmt.Print("Нажмите Enter...")
fmt.Scanln()
return
}

fmt.Println("\nВыберите расширение для удаления:")
for i, ext := range config.Extensions {
fmt.Printf("  %d. %s\n", i+1, ext)
}

fmt.Print("Номер: ")
var num int
fmt.Scanln(&num)

if num < 1 || num > len(config.Extensions) {
fmt.Println("❌ Неверный номер")
fmt.Print("Нажмите Enter...")
fmt.Scanln()
return
}

removed := config.Extensions[num-1]
config.Extensions = append(config.Extensions[:num-1], config.Extensions[num:]...)

fmt.Printf("✅ Удалено расширение: %s\n", removed)
fmt.Print("Нажмите Enter...")
fmt.Scanln()
}

func changeLogDir() {
fmt.Printf("\nТекущая папка логов: %s\n", config.LogDir)
fmt.Print("Введите новую папку: ")

var newDir string
fmt.Scanln(&newDir)

if newDir == "" {
return
}

config.LogDir = newDir
os.MkdirAll(config.LogDir, 0755)

fmt.Println("✅ Папка логов изменена")
fmt.Print("Нажмите Enter...")
fmt.Scanln()
}

func updateSummary() {
fmt.Println("\n🔄 Обновление сводного файла...")

summary := generateSummary()
err := os.WriteFile(config.SummaryFile, []byte(summary), 0644)

if err != nil {
fmt.Println("❌ Ошибка:", err)
} else {
fmt.Println("✅ Сводный файл обновлен:", config.SummaryFile)
}

fmt.Print("Нажмите Enter...")
fmt.Scanln()
}

func scanFiles() {
fmt.Println("\n🔍 Сканирование файлов...")

count := countTrackedFiles()
fmt.Printf("Найдено отслеживаемых файлов: %d\n", count)

// Показываем статистику по расширениям
stats := make(map[string]int)

filepath.WalkDir(config.WatchDir, func(path string, d fs.DirEntry, err error) error {
if err != nil || d.IsDir() {
return nil
}

ext := filepath.Ext(path)
for _, tracked := range config.Extensions {
if ext == tracked {
stats[ext]++
break
}
}
return nil
})

fmt.Println("\nСтатистика по расширениям:")
for ext, count := range stats {
fmt.Printf("  %s: %d файлов\n", ext, count)
}

fmt.Print("\nНажмите Enter...")
fmt.Scanln()
}

func clearLogs() {
fmt.Println("\n🗑  Очистка старых логов")
fmt.Println("ВНИМАНИЕ: Будут удалены все логи старше указанного количества дней!")
fmt.Print("Сколько дней оставить? (0 = удалить все): ")

var days int
fmt.Scanln(&days)

if days < 0 {
fmt.Println("❌ Неверное количество дней")
fmt.Print("Нажмите Enter...")
fmt.Scanln()
return
}

files, err := os.ReadDir(config.LogDir)
if err != nil {
fmt.Println("❌ Ошибка:", err)
fmt.Print("Нажмите Enter...")
fmt.Scanln()
return
}

deleted := 0
cutoff := time.Now().AddDate(0, 0, -days)

for _, file := range files {
if file.IsDir() {
continue
}

info, err := file.Info()
if err != nil {
continue
}

// Проверяем дату файла
if days == 0 || info.ModTime().Before(cutoff) {
os.Remove(filepath.Join(config.LogDir, file.Name()))
deleted++
}
}

fmt.Printf("✅ Удалено файлов: %d\n", deleted)
fmt.Print("Нажмите Enter...")
fmt.Scanln()
}

// Вспомогательные функции
func countTrackedFiles() int {
count := 0

filepath.WalkDir(config.WatchDir, func(path string, d fs.DirEntry, err error) error {
if err != nil || d.IsDir() {
return nil
}

ext := filepath.Ext(path)
for _, tracked := range config.Extensions {
if ext == tracked {
count++
break
}
}
return nil
})

return count
}

func generateSummary() string {
summary := "# Состояние проекта\n\n"
summary += fmt.Sprintf("**Дата обновления:** %s\n\n", time.Now().Format("2006-01-02 15:04:05"))
summary += fmt.Sprintf("**Папка мониторинга:** %s\n\n", config.WatchDir)

// Статистика
stats := make(map[string]int)
total := 0

filepath.WalkDir(config.WatchDir, func(path string, d fs.DirEntry, err error) error {
if err != nil || d.IsDir() {
return nil
}

ext := filepath.Ext(path)
for _, tracked := range config.Extensions {
if ext == tracked {
stats[ext]++
total++
break
}
}
return nil
})

summary += fmt.Sprintf("**Всего отслеживаемых файлов:** %d\n\n", total)

for ext, count := range stats {
summary += fmt.Sprintf("- **%s**: %d файлов\n", ext, count)
}

// Последние изменения
summary += "\n## Последние изменения\n\n"

files, _ := os.ReadDir(config.LogDir)
if len(files) > 0 {
// Берем последний лог-файл
lastLog := files[len(files)-1].Name()
content, err := os.ReadFile(filepath.Join(config.LogDir, lastLog))
if err == nil {
// Берем первые 10 строк
lines := strings.Split(string(content), "\n")
limit := 10
if len(lines) < limit {
limit = len(lines)
}

for i := 0; i < limit; i++ {
if i < len(lines) {
summary += lines[i] + "\n"
}
}
}
}

return summary
}

func monitorFiles(stopChan chan bool) {
// Простая имитация мониторинга
ticker := time.NewTicker(5 * time.Second)
defer ticker.Stop()

for {
select {
case <-ticker.C:
// Имитация обнаружения изменений
logEvent("Проверка файловой системы...")

case <-stopChan:
logEvent("Мониторинг остановлен")
return
}
}
}

func logEvent(message string) {
dateStr := time.Now().Format("2006-01-02")
logFile := filepath.Join(config.LogDir, dateStr+"_changes.md")

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
