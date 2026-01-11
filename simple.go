package main

import (
"fmt"
"os"
"path/filepath"
"time"
)

func main() {
// Создаем окно консоли
fmt.Println("╔══════════════════════════════════════╗")
fmt.Println("║      AILAN ARCHIVIST v1.0           ║")
fmt.Println("║  Автономный монитор файлов          ║")
fmt.Println("╚══════════════════════════════════════╝")
fmt.Println()
fmt.Println("Функции:")
fmt.Println("• Отслеживание .php .html .js .css .txt .json")
fmt.Println("• Автономная работа с флешки")
fmt.Println("• Не требует установки PowerShell")
fmt.Println()

// Создаем структуру папок
createFolders()

// Показываем текущую папку
dir, _ := os.Getwd()
fmt.Printf("Текущая папка: %s\n", dir)
fmt.Printf("Время запуска: %s\n", time.Now().Format("2006-01-02 15:04:05"))
fmt.Println()

// Проверяем файлы
count := countFiles()
fmt.Printf("Найдено отслеживаемых файлов: %d\n", count)

// Создаем лог-файл
createLogFile()

fmt.Println()
fmt.Println("Для выхода нажмите Enter...")
fmt.Scanln()
}

func createFolders() {
os.MkdirAll("docs/changelog", 0755)
fmt.Println("✅ Создана структура папок")
}

func countFiles() int {
count := 0
exts := []string{".php", ".html", ".js", ".css", ".txt", ".json"}

filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
if err != nil {
return nil
}

if !info.IsDir() {
ext := filepath.Ext(path)
for _, e := range exts {
if ext == e {
count++
break
}
}
}
return nil
})

return count
}

func createLogFile() {
logFile := filepath.Join("docs", "changelog", time.Now().Format("2006-01-02")+"_start.md")
content := fmt.Sprintf("# Запуск AILAN Archivist\n\nВремя: %s\n", time.Now().Format("2006-01-02 15:04:05"))
os.WriteFile(logFile, []byte(content), 0644)
fmt.Println("✅ Создан лог-файл запуска")
}
