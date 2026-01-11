package main

import (
	"fmt"
	"image/color"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

// Конфигурация
type Config struct {
	WatchDir     string
	Extensions   []string
	LogDir       string
	SummaryFile  string
}

// Главное окно
type MainWindow struct {
	*walk.MainWindow
	logText  *walk.TextEdit
	status   *walk.StatusBarItem
	startBtn *walk.PushButton
	stopBtn  *walk.PushButton
	watcher  *fsnotify.Watcher
	fileCountLabel *walk.Label
	lastEventLabel *walk.Label
}

var (
	config  Config
	mainWin *MainWindow
	icon    *walk.Icon
)

func main() {
	fmt.Println("Запуск AILAN Archivist GUI...")
	
	// Инициализация конфигурации
	initConfig()
	
	// Запуск GUI
	runGUI()
}

func initConfig() {
	watchDir, _ := os.Getwd()
	
	config = Config{
		WatchDir:    watchDir,
		Extensions:  []string{".php", ".html", ".js", ".css", ".txt", ".json"},
		LogDir:      "docs/changelog",
		SummaryFile: "docs/project_state.md",
	}
	
	// Создаем папки если их нет
	os.MkdirAll(config.LogDir, 0755)
	os.MkdirAll(filepath.Dir(config.SummaryFile), 0755)
}

func runGUI() {
	// Пытаемся создать иконку из ресурсов
	createIconFromResource()
	
	// Создаем главное окно
	mw := &MainWindow{}
	
	// Размеры окна
	windowWidth := 1000
	windowHeight := 700
	
	err := MainWindow{
		AssignTo: &mw.MainWindow,
		Title:    "📁 AILAN Archivist v2.0 - Автономный файловый монитор",
		MinSize:  Size{Width: 900, Height: 600},
		Size:     Size{Width: windowWidth, Height: windowHeight},
		Icon:     icon,
		Layout:   VBox{MarginsZero: true, SpacingZero: true},
		
		Children: []Widget{
			// Верхняя панель с информацией
			Composite{
				Layout: HBox{Margins: Margins{Left: 10, Right: 10, Top: 5, Bottom: 5}},
				Children: []Widget{
					// Логотип и название
					Label{
						Text:  "🛡️ AILAN ARCHIVIST",
						Font:  Font{Bold: true, PointSize: 16},
						TextColor: color.RGBA{0, 100, 200, 255},
					},
					HSpace{},
					// Статистика
					Composite{
						Layout: HBox{},
						Children: []Widget{
							Label{
								AssignTo: &mw.fileCountLabel,
								Text:     "📊 Файлов: 0",
								Font:     Font{PointSize: 10},
							},
							Label{
								Text: " | ",
							},
							Label{
								AssignTo: &mw.lastEventLabel,
								Text:     "⏰ Последнее: --:--:--",
								Font:     Font{PointSize: 10},
							},
						},
					},
				},
			},
			
			// Панель управления
			Composite{
				Layout:  HBox{Margins: Margins{Left: 10, Right: 10, Top: 0, Bottom: 10}},
				MaxSize: Size{Height: 50},
				Children: []Widget{
					// Кнопка запуска
					PushButton{
						AssignTo:  &mw.startBtn,
						Text:      "▶ ЗАПУСТИТЬ МОНИТОРИНГ",
						Font:      Font{Bold: true, PointSize: 11},
						MinSize:   Size{Width: 180, Height: 35},
						OnClicked: mw.startMonitoring,
					},
					
					// Кнопка остановки
					PushButton{
						AssignTo:  &mw.stopBtn,
						Text:      "⏹ ОСТАНОВИТЬ",
						Font:      Font{Bold: true, PointSize: 11},
						MinSize:   Size{Width: 120, Height: 35},
						Enabled:   false,
						OnClicked: mw.stopMonitoring,
					},
					
					VSeparator{MinSize: Size{Width: 20}},
					
					// Кнопка выбора папки
					PushButton{
						Text:    "📂 ВЫБРАТЬ ПАПКУ",
						Font:    Font{PointSize: 10},
						MinSize: Size{Width: 130, Height: 35},
						OnClicked: func() {
							mw.selectFolder()
						},
					},
					
					// Кнопка открытия логов
					PushButton{
						Text:    "📊 ОТКРЫТЬ ЛОГИ",
						Font:    Font{PointSize: 10},
						MinSize: Size{Width: 120, Height: 35},
						OnClicked: func() {
							mw.openLogs()
						},
					},
					
					// Кнопка настроек
					PushButton{
						Text:    "⚙ НАСТРОЙКИ",
						Font:    Font{PointSize: 10},
						MinSize: Size{Width: 110, Height: 35},
						OnClicked: func() {
							mw.showSettings()
						},
					},
					
					HSpace{},
					
					// Индикатор активности
					ProgressBar{
						MarqueeMode: true,
						Visible:     false,
						MinSize:     Size{Width: 100},
					},
				},
			},
			
			// Основная область с логами
			Composite{
				Layout: VBox{Margins: Margins{Left: 10, Right: 10, Top: 0, Bottom: 10}},
				Children: []Widget{
					Label{
						Text: "📝 ЖУРНАЛ СОБЫТИЙ В РЕАЛЬНОМ ВРЕМЕНИ:",
						Font: Font{Bold: true, PointSize: 11},
					},
					TextEdit{
						AssignTo:      &mw.logText,
						ReadOnly:      true,
						VScroll:       true,
						HScroll:       true,
						Font:          Font{Family: "Consolas", PointSize: 10},
						MinSize:       Size{Height: 400},
						Text:          getWelcomeMessage(),
					},
				},
			},
			
			// Нижняя панель с путем
			Composite{
				Layout: HBox{Margins: Margins{Left: 10, Right: 10, Top: 0, Bottom: 5}},
				Children: []Widget{
					Label{
						Text: "📁 Текущая папка: " + config.WatchDir,
						Font: Font{PointSize: 9},
					},
					HSpace{},
					Label{
						Text: "🔄 Автономный режим | PowerShell не требуется",
						Font: Font{PointSize: 9, Bold: true},
						TextColor: color.RGBA{0, 150, 0, 255},
					},
				},
			},
		},
		
		StatusBarItems: []StatusBarItem{
			{
				AssignTo: &mw.status,
				Text:     "✅ ГОТОВ К РАБОТЕ",
				Width:    200,
			},
			StatusBarItem{
				Text:     fmt.Sprintf("🖥 %s", getWindowsVersion()),
				Width:    120,
			},
			StatusBarItem{
				Text:     time.Now().Format("🕐 15:04:05"),
				Width:    100,
			},
		},
	}.Create()
	
	if err != nil {
		log.Fatal("Ошибка создания окна:", err)
	}
	
	mainWin = mw
	
	// Обновляем статистику файлов
	mw.updateFileCount()
	
	// Запускаем обновление времени в статусбаре
	go mw.updateTime()
	
	// Показываем окно
	mw.SetX((walk.ScreenWidth() - windowWidth) / 2)
	mw.SetY((walk.ScreenHeight() - windowHeight) / 2)
	mw.Run()
}

func getWelcomeMessage() string {
	return `================================================================================
                  ДОБРО ПОЖАЛОВАТЬ В AILAN ARCHIVIST v2.0
================================================================================

📌 НАЗНАЧЕНИЕ:
   Автономный файловый архивариус для локальной разработки.
   Работает без интернета, не требует установки PowerShell, Python или других сред.

📌 ВОЗМОЖНОСТИ:
   • Отслеживание изменений файлов в реальном времени
   • Автоматическое логирование в формате Markdown
   • Обновление сводного файла состояния проекта
   • Работа с флешки на любом ПК с Windows 10/11

📌 ОТСЛЕЖИВАЕМЫЕ РАСШИРЕНИЯ:
   .php .html .js .css .txt .json

📌 ПУТИ СОХРАНЕНИЯ:
   • Логи изменений: docs/changelog/ГГГГ-ММ-ДД_changes.md
   • Сводный файл: docs/project_state.md

================================================================================
   Нажмите "ЗАПУСТИТЬ МОНИТОРИНГ" для начала работы...
================================================================================
`
}

func (mw *MainWindow) updateFileCount() {
	count := 0
	filepath.WalkDir(config.WatchDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		for _, trackedExt := range config.Extensions {
			if ext == trackedExt {
				count++
				break
			}
		}
		return nil
	})
	
	mw.fileCountLabel.SetText(fmt.Sprintf("📊 Файлов: %d", count))
}

func (mw *MainWindow) startMonitoring() {
	mw.startBtn.SetEnabled(false)
	mw.stopBtn.SetEnabled(true)
	mw.status.SetText("🔴 МОНИТОРИНГ АКТИВЕН")
	
	mw.addLog("\n" + strings.Repeat("=", 80))
	mw.addLog("🚀 ЗАПУСК МОНИТОРИНГА ФАЙЛОВОЙ СИСТЕМЫ")
	mw.addLog(fmt.Sprintf("📁 Папка: %s", config.WatchDir))
	mw.addLog(fmt.Sprintf("🕐 Время запуска: %s", time.Now().Format("2006-01-02 15:04:05")))
	mw.addLog(strings.Repeat("=", 80) + "\n")
	
	// Инициализируем наблюдатель
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		mw.addLog("❌ ОШИБКА: Не удалось создать наблюдатель файловой системы")
		return
	}
	mw.watcher = watcher
	
	// Добавляем папки для наблюдения
	mw.addWatchRecursive(config.WatchDir)
	
	// Обновляем сводный файл
	mw.updateProjectSummary()
	
	// Запускаем обработчик событий
	go mw.handleWatcherEvents()
}

func (mw *MainWindow) stopMonitoring() {
	if mw.watcher != nil {
		mw.watcher.Close()
		mw.watcher = nil
	}
	
	mw.startBtn.SetEnabled(true)
	mw.stopBtn.SetEnabled(false)
	mw.status.SetText("✅ МОНИТОРИНГ ОСТАНОВЛЕН")
	mw.addLog("\n" + strings.Repeat("=", 80))
	mw.addLog("⏹ МОНИТОРИНГ ОСТАНОВЛЕН ПОЛЬЗОВАТЕЛЕМ")
	mw.addLog(fmt.Sprintf("🕐 Время остановки: %s", time.Now().Format("2006-01-02 15:04:05")))
	mw.addLog(strings.Repeat("=", 80) + "\n")
}

func (mw *MainWindow) addWatchRecursive(dir string) {
	filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		
		if d.IsDir() {
			// Пропускаем служебные папки
			name := d.Name()
			if name == ".git" || name == "node_modules" || name == ".vs" || name == ".idea" {
				return filepath.SkipDir
			}
			
			if err := mw.watcher.Add(path); err == nil {
				mw.addLog(fmt.Sprintf("👁 Добавлено наблюдение: %s", path))
			}
		}
		return nil
	})
}

func (mw *MainWindow) handleWatcherEvents() {
	for {
		select {
		case event, ok := <-mw.watcher.Events:
			if !ok {
				return
			}
			
			if mw.shouldTrackFile(event.Name) {
				mw.processFileEvent(event)
			}
			
			// Если создана новая директория
			if event.Op.Has(fsnotify.Create) {
				if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
					mw.watcher.Add(event.Name)
					mw.addWatchRecursive(event.Name)
				}
			}
			
		case err, ok := <-mw.watcher.Errors:
			if !ok {
				return
			}
			mw.addLog(fmt.Sprintf("⚠ ОШИБКА НАБЛЮДАТЕЛЯ: %v", err))
		}
	}
}

func (mw *MainWindow) shouldTrackFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	for _, trackedExt := range config.Extensions {
		if ext == trackedExt {
			return true
		}
	}
	return false
}

func (mw *MainWindow) processFileEvent(event fsnotify.Event) {
	now := time.Now()
	timeStr := now.Format("15:04:05")
	
	var action, emoji string
	switch {
	case event.Op.Has(fsnotify.Create):
		action, emoji = "СОЗДАН", "🆕"
	case event.Op.Has(fsnotify.Write):
		action, emoji = "ИЗМЕНЕН", "📝"
	case event.Op.Has(fsnotify.Remove):
		action, emoji = "УДАЛЕН", "🗑️"
	case event.Op.Has(fsnotify.Rename):
		action, emoji = "ПЕРЕИМЕНОВАН", "🏷️"
	default:
		action, emoji = "НЕИЗВЕСТНО", "❓"
	}
	
	relPath, _ := filepath.Rel(config.WatchDir, event.Name)
	message := fmt.Sprintf("[%s] %s %s: %s", timeStr, emoji, action, relPath)
	
	// Обновляем UI
	walk.MainWindowSynchronized(func() {
		mw.addLog(message)
		mw.lastEventLabel.SetText(fmt.Sprintf("⏰ Последнее: %s", timeStr))
	})
	
	// Записываем в лог файл
	mw.writeToChangeLog(now, action, emoji, relPath, event.Name)
	
	// Обновляем сводный файл
	mw.updateProjectSummary()
	
	// Обновляем счетчик файлов
	mw.updateFileCount()
}

func (mw *MainWindow) addLog(message string) {
	currentText := mw.logText.Text()
	if len(currentText) > 50000 {
		lines := strings.Split(currentText, "\n")
		if len(lines) > 500 {
			currentText = strings.Join(lines[len(lines)-500:], "\n")
		}
	}
	mw.logText.SetText(currentText + "\n" + message)
	mw.logText.SetCaretPos(len(mw.logText.Text()))
}

func (mw *MainWindow) writeToChangeLog(t time.Time, action, emoji, relPath, fullPath string) {
	dateStr := t.Format("2006-01-02")
	logFile := filepath.Join(config.LogDir, dateStr+"_changes.md")
	
	entry := fmt.Sprintf("### %s %s\n", emoji, t.Format("15:04:05"))
	entry += fmt.Sprintf("- **Файл:** `%s`\n", relPath)
	entry += fmt.Sprintf("- **Действие:** %s\n", action)
	entry += fmt.Sprintf("- **Время:** %s\n\n", t.Format("2006-01-02 15:04:05"))
	
	content, err := os.ReadFile(logFile)
	if err != nil {
		header := fmt.Sprintf("# Изменения за %s\n\n", dateStr)
		entry = header + entry
	} else {
		entry = string(content) + "\n" + entry
	}
	
	os.WriteFile(logFile, []byte(entry), 0644)
}

func (mw *MainWindow) updateProjectSummary() {
	summary := "# Состояние проекта\n\n"
	summary += fmt.Sprintf("**Последнее обновление:** %s\n\n", time.Now().Format("2006-01-02 15:04:05"))
	
	fileCounts := make(map[string]int)
	total := 0
	
	filepath.WalkDir(config.WatchDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		
		ext := filepath.Ext(path)
		if mw.shouldTrackFile(path) {
			fileCounts[ext]++
			total++
		}
		return nil
	})
	
	summary += fmt.Sprintf("**Всего отслеживаемых файлов:** %d\n\n", total)
	for ext, count := range fileCounts {
		if ext != "" {
			summary += fmt.Sprintf("- %s: %d файлов\n", ext, count)
		}
	}
	
	os.WriteFile(config.SummaryFile, []byte(summary), 0644)
}

func (mw *MainWindow) selectFolder() {
	dlg := new(walk.FileDialog)
	dlg.Title = "Выберите папку для мониторинга"
	dlg.Filter = "Папки|*.||"
	
	if ok, _ := dlg.ShowBrowseFolder(mw); ok {
		config.WatchDir = dlg.FilePath
		mw.addLog(fmt.Sprintf("📂 Изменена папка мониторинга на: %s", config.WatchDir))
		mw.updateFileCount()
	}
}

func (mw *MainWindow) openLogs() {
	logPath, _ := filepath.Abs(config.LogDir)
	os.StartProcess("explorer.exe", []string{logPath}, &os.ProcAttr{})
}

func (mw *MainWindow) showSettings() {
	walk.MsgBox(mw, "Настройки", 
		"Отслеживаемые расширения:\n" + strings.Join(config.Extensions, ", ") + 
		"\n\nПуть к логам: " + config.LogDir + 
		"\nСводный файл: " + config.SummaryFile, 
		walk.MsgBoxIconInformation)
}

func (mw *MainWindow) updateTime() {
	for {
		time.Sleep(1 * time.Second)
		walk.MainWindowSynchronized(func() {
			if mw.StatusBar().Items().At(2) != nil {
				mw.StatusBar().Items().At(2).SetText(time.Now().Format("🕐 15:04:05"))
			}
		})
	}
}

func createIconFromResource() {
	// Создаем простую иконку программно
	img := walk.NewBitmapWithTransparentPixels(walk.Size{Width: 32, Height: 32})
	
	// Рисуем синий круг
	canvas, _ := img.NewCanvas()
	canvas.FillEllipse(walk.NewSolidColorBrush(walk.RGB(0, 100, 200)), 
		walk.Rectangle{X: 0, Y: 0, Width: 32, Height: 32})
	
	// Рисуем белую папку
	canvas.FillRectangle(walk.NewSolidColorBrush(walk.RGB(255, 255, 255)), 
		walk.Rectangle{X: 8, Y: 10, Width: 16, Height: 12})
	
	// Сохраняем как временный файл и загружаем как иконку
	tmpFile := "temp_icon.png"
	img.SaveToFile(tmpFile)
	
	var err error
	icon, err = walk.NewIconFromFile(tmpFile)
	if err != nil {
		fmt.Println("Не удалось создать иконку:", err)
	}
	
	// Удаляем временный файл
	os.Remove(tmpFile)
}

func getWindowsVersion() string {
	return "Windows 10/11"
}