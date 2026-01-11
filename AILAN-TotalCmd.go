package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
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
	LastScanTime  time.Time
	TotalFiles    int
}

var (
	cfg           Config
	app           *tview.Application
	pages         *tview.Pages
	leftPane      *tview.List
	rightPane     *tview.List
	statusBar     *tview.TextView
	logView       *tview.TextView
	mainFlex      *tview.Flex
	currentPanel  string
	monitoring    bool
)

// === ОСНОВНАЯ ФУНКЦИЯ ===
func main() {
	fmt.Print("\033]0;🚀 AILAN ARCHIVIST - Total Commander Style\007")
	
	initConfig()
	initUI()
	
	if err := app.SetRoot(pages, true).EnableMouse(true).Run(); err != nil {
		fmt.Printf("Ошибка запуска: %v\n", err)
		os.Exit(1)
	}
}

func initConfig() {
	exeDir, err := os.Getwd()
	if err != nil {
		exeDir = "."
	}
	
	cfg = Config{
		WatchDir:      exeDir,
		Extensions:    []string{".php", ".html", ".js", ".css", ".txt", ".json", ".md", ".go"},
		LogDir:        filepath.Join(exeDir, "docs", "changelog"),
		SummaryFile:   filepath.Join(exeDir, "docs", "project_state.md"),
		IsRunning:     false,
		CustomFolders: []string{exeDir},
		LastScanTime:  time.Now(),
	}
	
	os.MkdirAll(cfg.LogDir, 0755)
	os.MkdirAll(filepath.Dir(cfg.SummaryFile), 0755)
	
	currentPanel = "left"
	monitoring = false
}

func initUI() {
	app = tview.NewApplication()
	pages = tview.NewPages()
	
	tview.Styles.PrimitiveBackgroundColor = tcell.ColorBlack
	tview.Styles.ContrastBackgroundColor = tcell.ColorBlue
	tview.Styles.BorderColor = tcell.ColorWhite
	tview.Styles.TitleColor = tcell.ColorYellow
	tview.Styles.PrimaryTextColor = tcell.ColorWhite
	
	createMainUI()
	pages.AddPage("main", mainFlex, true, true)
	setupHotkeys()
	
	go updateStatusPeriodically()
}

func createMainUI() {
	topPanel := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)
	
	topPanel.SetBorder(false).
		SetBackgroundColor(tcell.ColorDarkBlue)
	
	updateTopPanel(topPanel)
	
	leftPane = tview.NewList().
		ShowSecondaryText(true).
		SetHighlightFullLine(true)
	
	leftPane.SetBorder(true).
		SetBorderColor(tcell.ColorWhite).
		SetTitle(" [yellow]🖥 КОМАНДЫ[-] ").
		SetTitleAlign(tview.AlignLeft).
		SetBackgroundColor(tcell.ColorBlack)
	
	updateLeftPane()
	
	rightPane = tview.NewList().
		ShowSecondaryText(true).
		SetHighlightFullLine(true)
	
	rightPane.SetBorder(true).
		SetBorderColor(tcell.ColorWhite).
		SetTitle(" [yellow]📁 МОНИТОРИНГ[-] ").
		SetTitleAlign(tview.AlignLeft).
		SetBackgroundColor(tcell.ColorBlack)
	
	updateRightPane()
	
	logView = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetChangedFunc(func() {
			app.Draw()
		})
	
	logView.SetBorder(true).
		SetBorderColor(tcell.ColorWhite).
		SetTitle(" [yellow]📝 ЖУРНАЛ[-] ").
		SetBackgroundColor(tcell.ColorBlack)
	
	statusBar = tview.NewTextView().
		SetDynamicColors(true)
	
	statusBar.SetBorder(false).
		SetBackgroundColor(tcell.ColorDarkBlue)
	
	updateStatusBar()
	
	mainFlex = tview.NewFlex().SetDirection(tview.FlexRow)
	mainFlex.AddItem(topPanel, 1, 0, false)
	
	contentFlex := tview.NewFlex().SetDirection(tview.FlexColumn)
	contentFlex.AddItem(leftPane, 0, 1, true)
	contentFlex.AddItem(rightPane, 0, 1, false)
	
	mainFlex.AddItem(contentFlex, 0, 3, true)
	mainFlex.AddItem(logView, 10, 1, false)
	mainFlex.AddItem(statusBar, 1, 0, false)
}

func updateTopPanel(panel *tview.TextView) {
	if panel == nil {
		return
	}
	
	status := "[red]🔴 ВЫКЛ"
	if monitoring {
		status = "[green]🟢 ВКЛ"
	}
	
	text := fmt.Sprintf(`[white]🚀 AILAN ARCHIVIST | Мониторинг: %s | Файлов: [yellow]%d[-] | Папок: [yellow]%d[-]`, 
		status, cfg.TotalFiles, len(cfg.CustomFolders))
	
	panel.SetText(text)
}

func updateLeftPane() {
	leftPane.Clear()
	
	leftPane.AddItem("[yellow]🚀 ОСНОВНЫЕ КОМАНДЫ[-]", "", 0, nil)
	
	if monitoring {
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
	
	leftPane.AddItem("[cyan]⚙ Управление расширениями[-]", "Нажмите Enter или F9", 'E', func() {
		showExtensionManager()
	})
	
	leftPane.AddItem("", "", 0, nil)
	leftPane.AddItem("[yellow]📊 ИНФОРМАЦИЯ[-]", "", 0, nil)
	
	leftPane.AddItem("[white]📈 Статистика[-]", "Нажмите Enter или F2", 'T', func() {
		showStatistics()
	})
	
	leftPane.AddItem("[white]📝 Просмотр логов[-]", "Нажмите Enter или F3", 'L', func() {
		showLogViewer()
	})
	
	leftPane.AddItem("[white]🛠 Настройки[-]", "Нажмите Enter или F6", 'N', func() {
		showSettings()
	})
	
	leftPane.AddItem("", "", 0, nil)
	leftPane.AddItem("[yellow]🚪 ВЫХОД[-]", "", 0, nil)
	
	leftPane.AddItem("[red]❌ Выход[-]", "Нажмите Enter или F10", 'X', func() {
		app.Stop()
	})
}

func updateRightPane() {
	rightPane.Clear()
	
	rightPane.AddItem("[yellow]📂 ОТСЛЕЖИВАЕМЫЕ ПАПКИ[-]", "", 0, nil)
	rightPane.AddItem("", "", 0, nil)
	
	if len(cfg.CustomFolders) == 0 {
		rightPane.AddItem("[gray]Нет добавленных папок[-]", "Нажмите F4 чтобы добавить", 0, func() {
			showAddFolderDialog()
		})
	} else {
		for i, folder := range cfg.CustomFolders {
			folderName := folder
			if len(folderName) > 35 {
				folderName = "..." + folderName[len(folderName)-32:]
			}
			
			count := countFilesInFolder(folder)
			
			text := fmt.Sprintf("📁 [yellow]%s[-]", folderName)
			secondary := fmt.Sprintf("[gray]Файлов: %d | Нажмите Enter для управления[-]", count)
			
			idx := i
			rightPane.AddItem(text, secondary, 0, func() {
				manageFolder(idx)
			})
		}
	}
	
	rightPane.AddItem("", "", 0, nil)
	rightPane.AddItem("[yellow]📄 ОТСЛЕЖИВАЕМЫЕ РАСШИРЕНИЯ[-]", "", 0, nil)
	rightPane.AddItem("", "", 0, nil)
	
	for i, ext := range cfg.Extensions {
		if i < 10 {
			rightPane.AddItem(fmt.Sprintf("  [cyan]%s[-]", ext), "", 0, nil)
		} else if i == 10 {
			rightPane.AddItem("  [gray]... и еще[-]", fmt.Sprintf("%d расширений", len(cfg.Extensions)-10), 0, nil)
		}
	}
}

func updateStatusBar() {
	timeStr := time.Now().Format("15:04:05")
	dateStr := time.Now().Format("02.01.2006")
	
	var helpText string
	if currentPanel == "left" {
		helpText = "[F1]Помощь [F4]Добавить [F5]Мониторинг [F7]Сканировать [F8]Удалить [F10]Выход"
	} else {
		helpText = "[Tab]Панели [Enter]Выбрать [F4]Добавить [F8]Удалить [F9]Расширения [Ctrl+Q]Выход"
	}
	
	statusText := fmt.Sprintf("[white]%s %s | %s", dateStr, timeStr, helpText)
	statusBar.SetText(statusText)
}

func updateStatusPeriodically() {
	for {
		time.Sleep(1 * time.Second)
		if app != nil {
			app.QueueUpdateDraw(func() {
				updateStatusBar()
			})
		}
	}
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
		case tcell.KeyF6:
			showSettings()
			return nil
		case tcell.KeyF7:
			quickScan()
			return nil
		case tcell.KeyF8:
			deleteSelectedFolder()
			return nil
		case tcell.KeyF9:
			showExtensionManager()
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
		case tcell.KeyCtrlQ:
			app.Stop()
			return nil
		case tcell.KeyCtrlS:
			quickScan()
			return nil
		case tcell.KeyCtrlM:
			toggleMonitoring()
			return nil
		case tcell.KeyInsert:
			showAddFolderDialog()
			return nil
		case tcell.KeyDelete:
			deleteSelectedFolder()
			return nil
		}
		
		return event
	})
}

func togglePanel() {
	if currentPanel == "left" {
		currentPanel = "right"
		app.SetFocus(rightPane)
		rightPane.SetTitle(" [yellow]📁 МОНИТОРИНГ[-] [green]◄ АКТИВНА[-] ")
		leftPane.SetTitle(" [yellow]🖥 КОМАНДЫ[-] ")
	} else {
		currentPanel = "left"
		app.SetFocus(leftPane)
		leftPane.SetTitle(" [yellow]🖥 КОМАНДЫ[-] [green]◄ АКТИВНА[-] ")
		rightPane.SetTitle(" [yellow]📁 МОНИТОРИНГ[-] ")
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
		SetText("[yellow]🚀 AILAN ARCHIVIST - ГОРЯЧИЕ КЛАВИШИ[-]\n\n" +
			"[cyan]ОСНОВНЫЕ КЛАВИШИ:[-]\n" +
			"[white]F1[-] - Эта справка\n" +
			"[white]F4/Ins[-] - Добавить папку\n" +
			"[white]F5[-] - Вкл/Выкл мониторинг\n" +
			"[white]F7[-] - Быстрое сканирование\n" +
			"[white]F8/Del[-] - Удалить папку\n" +
			"[white]F9[-] - Управление расширениями\n" +
			"[white]F10[-] - Выход\n\n" +
			"[cyan]УПРАВЛЕНИЕ:[-]\n" +
			"[white]Tab[-] - Переключение панелей\n" +
			"[white]Enter[-] - Выполнить команду\n" +
			"[white]↑↓[-] - Навигация").
		AddButtons([]string{"[white]Закрыть[-]"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			pages.HidePage("help")
		})
	
	modal.SetBorder(true).
		SetTitle(" [yellow]❓ СПРАВКА[-] ").
		SetBackgroundColor(tcell.ColorBlack)
	
	pages.AddPage("help", modal, true, true)
}

func showAddFolderDialog() {
// Создаем форму
form := tview.NewForm().
AddInputField("Путь к папке:", "", 50, nil, nil).
AddButton("[green]Добавить[-]", func() {
folderPath := form.GetFormItem(0).(*tview.InputField).GetText()

if folderPath != "" {
if info, err := os.Stat(folderPath); err == nil && info.IsDir() {
cfg.CustomFolders = append(cfg.CustomFolders, folderPath)
cfg.TotalFiles = countAllTrackedFiles()

addLogEntry(fmt.Sprintf("Добавлена папка: %s", folderPath))

updateRightPane()
updateTopPanel(nil)

pages.HidePage("addFolder")
} else {
showErrorDialog("Ошибка", "Папка не существует")
}
}
}).
AddButton("[red]Отмена[-]", func() {
pages.HidePage("addFolder")
})

form.SetBorder(true).
SetTitle(" [yellow]📁 ДОБАВЛЕНИЕ ПАПКИ[-] ").
SetBackgroundColor(tcell.ColorBlack)

flex := tview.NewFlex().
AddItem(nil, 0, 1, false).
AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
AddItem(nil, 0, 1, false).
AddItem(form, 7, 0, true).
AddItem(nil, 0, 1, false), 60, 0, true).
AddItem(nil, 0, 1, false)

pages.AddPage("addFolder", flex, true, true)
app.SetFocus(form)
	// Создаем форму внутри функции
	form := tview.NewForm().
		AddInputField("Путь к папке:", "", 50, nil, nil).
		AddButton("[green]Добавить[-]", func() {
			// Получаем значение из поля
			folderPath := form.GetFormItem(0).(*tview.InputField).GetText()
			
			if folderPath != "" {
				if info, err := os.Stat(folderPath); err == nil && info.IsDir() {
					cfg.CustomFolders = append(cfg.CustomFolders, folderPath)
					cfg.TotalFiles = countAllTrackedFiles()
					
					addLogEntry(fmt.Sprintf("Добавлена папка: %s", folderPath))
					
					updateRightPane()
					updateTopPanel(nil)
					
					pages.HidePage("addFolder")
				} else {
					showErrorDialog("Ошибка", "Папка не существует")
				}
			}
		}).
		AddButton("[red]Отмена[-]", func() {
			pages.HidePage("addFolder")
		})
	
	form.SetBorder(true).
		SetTitle(" [yellow]📁 ДОБАВЛЕНИЕ ПАПКИ[-] ").
		SetBackgroundColor(tcell.ColorBlack)
	
	flex := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(form, 7, 0, true).
			AddItem(nil, 0, 1, false), 60, 0, true).
		AddItem(nil, 0, 1, false)
	
	pages.AddPage("addFolder", flex, true, true)
	app.SetFocus(form)
}

func showExtensionManager() {
	list := tview.NewList().
		ShowSecondaryText(false).
		SetHighlightFullLine(true)
	
	list.SetBorder(true).
		SetTitle(" [yellow]⚙ УПРАВЛЕНИЕ РАСШИРЕНИЯМИ[-] ").
		SetBackgroundColor(tcell.ColorBlack)
	
	for _, ext := range cfg.Extensions {
		list.AddItem(fmt.Sprintf("[cyan]%s[-]", ext), "", 0, nil)
	}
	
	buttons := tview.NewFlex().SetDirection(tview.FlexColumn)
	
	buttons.AddItem(tview.NewButton("[green]Добавить[-]").SetSelectedFunc(func() {
		showAddExtensionDialog()
	}), 0, 1, false)
	
	buttons.AddItem(tview.NewButton("[red]Удалить[-]").SetSelectedFunc(func() {
		idx := list.GetCurrentItem()
		if idx >= 0 {
			ext := cfg.Extensions[idx]
			cfg.Extensions = append(cfg.Extensions[:idx], cfg.Extensions[idx+1:]...)
			addLogEntry(fmt.Sprintf("Удалено расширение: %s", ext))
			showExtensionManager()
		}
	}), 0, 1, false)
	
	buttons.AddItem(tview.NewButton("[white]Закрыть[-]").SetSelectedFunc(func() {
		pages.HidePage("extensionManager")
	}), 0, 1, false)
	
	flex := tview.NewFlex().SetDirection(tview.FlexRow)
	flex.AddItem(list, 0, 1, true)
	flex.AddItem(buttons, 1, 0, false)
	
	center := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(flex, 15, 0, true).
			AddItem(nil, 0, 1, false), 50, 0, true).
		AddItem(nil, 0, 1, false)
	
	pages.AddPage("extensionManager", center, true, true)
}

func showAddExtensionDialog() {
// Создаем форму
form := tview.NewForm().
AddInputField("Расширение (начинается с точки):", ".", 20, nil, nil).
AddButton("[green]Добавить[-]", func() {
ext := form.GetFormItem(0).(*tview.InputField).GetText()
ext = strings.TrimSpace(ext)

if ext != "" && strings.HasPrefix(ext, ".") {
found := false
for _, existing := range cfg.Extensions {
if existing == ext {
found = true
break
}
}

if !found {
cfg.Extensions = append(cfg.Extensions, ext)
addLogEntry(fmt.Sprintf("Добавлено расширение: %s", ext))

cfg.TotalFiles = countAllTrackedFiles()
updateTopPanel(nil)

pages.HidePage("addExtension")
showExtensionManager()
} else {
showErrorDialog("Ошибка", "Расширение уже существует")
}
} else {
showErrorDialog("Ошибка", "Расширение должно начинаться с точки (.txt)")
}
}).
AddButton("[red]Отмена[-]", func() {
pages.HidePage("addExtension")
})

form.SetBorder(true).
SetTitle(" [yellow]➕ ДОБАВИТЬ РАСШИРЕНИЕ[-] ").
SetBackgroundColor(tcell.ColorBlack)

center := tview.NewFlex().
AddItem(nil, 0, 1, false).
AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
AddItem(nil, 0, 1, false).
AddItem(form, 7, 0, true).
AddItem(nil, 0, 1, false), 50, 0, true).
AddItem(nil, 0, 1, false)

pages.AddPage("addExtension", center, true, true)
	// Создаем форму внутри функции
	form := tview.NewForm().
		AddInputField("Расширение (начинается с точки):", ".", 20, nil, nil).
		AddButton("[green]Добавить[-]", func() {
			ext := form.GetFormItem(0).(*tview.InputField).GetText()
			ext = strings.TrimSpace(ext)
			
			if ext != "" && strings.HasPrefix(ext, ".") {
				found := false
				for _, existing := range cfg.Extensions {
					if existing == ext {
						found = true
						break
					}
				}
				
				if !found {
					cfg.Extensions = append(cfg.Extensions, ext)
					addLogEntry(fmt.Sprintf("Добавлено расширение: %s", ext))
					
					cfg.TotalFiles = countAllTrackedFiles()
					updateTopPanel(nil)
					
					pages.HidePage("addExtension")
					showExtensionManager()
				} else {
					showErrorDialog("Ошибка", "Расширение уже существует")
				}
			} else {
				showErrorDialog("Ошибка", "Расширение должно начинаться с точки (.txt)")
			}
		}).
		AddButton("[red]Отмена[-]", func() {
			pages.HidePage("addExtension")
		})
	
	form.SetBorder(true).
		SetTitle(" [yellow]➕ ДОБАВИТЬ РАСШИРЕНИЕ[-] ").
		SetBackgroundColor(tcell.ColorBlack)
	
	center := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(form, 7, 0, true).
			AddItem(nil, 0, 1, false), 50, 0, true).
		AddItem(nil, 0, 1, false)
	
	pages.AddPage("addExtension", center, true, true)
}

func toggleMonitoring() {
	monitoring = !monitoring
	
	if monitoring {
		addLogEntry("Мониторинг запущен")
		go startMonitoringProcess()
	} else {
		addLogEntry("Мониторинг остановлен")
	}
	
	updateLeftPane()
	updateTopPanel(nil)
}

func startMonitoringProcess() {
	for monitoring {
		for _, folder := range cfg.CustomFolders {
			scanFolderForChanges(folder)
		}
		
		time.Sleep(5 * time.Second)
		
		app.QueueUpdateDraw(func() {
			cfg.TotalFiles = countAllTrackedFiles()
			updateTopPanel(nil)
		})
	}
}

func scanFolderForChanges(folder string) {
	now := time.Now()
	if now.Sub(cfg.LastScanTime) > 30*time.Second {
		cfg.LastScanTime = now
		count := countFilesInFolder(folder)
		addLogEntry(fmt.Sprintf("Сканирование %s: %d файлов", filepath.Base(folder), count))
	}
}

func quickScan() {
	go func() {
		total := 0
		for _, folder := range cfg.CustomFolders {
			count := countFilesInFolder(folder)
			total += count
			addLogEntry(fmt.Sprintf("Папка %s: %d файлов", filepath.Base(folder), count))
		}
		
		cfg.TotalFiles = total
		
		app.QueueUpdateDraw(func() {
			updateTopPanel(nil)
			updateRightPane()
		})
		
		addLogEntry(fmt.Sprintf("Быстрое сканирование: всего %d файлов", total))
	}()
}

func showLogViewer() {
	textView := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetChangedFunc(func() {
			app.Draw()
		})
	
	textView.SetBorder(true).
		SetTitle(" [yellow]📝 ПРОСМОТР ЛОГОВ[-] ").
		SetBackgroundColor(tcell.ColorBlack)
	
	textView.SetText(loadRecentLogs(100))
	
	buttons := tview.NewFlex().SetDirection(tview.FlexColumn)
	
	buttons.AddItem(tview.NewButton("[green]Обновить[-]").SetSelectedFunc(func() {
		textView.SetText(loadRecentLogs(100))
	}), 0, 1, false)
	
	buttons.AddItem(tview.NewButton("[yellow]Очистить[-]").SetSelectedFunc(func() {
		textView.SetText("")
	}), 0, 1, false)
	
	buttons.AddItem(tview.NewButton("[white]Закрыть[-]").SetSelectedFunc(func() {
		pages.HidePage("logViewer")
	}), 0, 1, false)
	
	flex := tview.NewFlex().SetDirection(tview.FlexRow)
	flex.AddItem(textView, 0, 1, true)
	flex.AddItem(buttons, 1, 0, false)
	
	center := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(flex, 20, 0, true).
			AddItem(nil, 0, 1, false), 80, 0, true).
		AddItem(nil, 0, 1, false)
	
	pages.AddPage("logViewer", center, true, true)
}

func showStatistics() {
	modal := tview.NewModal()
	
	stats := gatherStatistics()
	
	text := fmt.Sprintf("[yellow]📊 СТАТИСТИКА[-]\n\n"+
		"[white]Всего файлов:[-] [cyan]%d[-]\n"+
		"[white]Папок:[-] [cyan]%d[-]\n"+
		"[white]Расширений:[-] [cyan]%d[-]\n\n"+
		"[white]МОНИТОРИНГ:[-] ", 
		stats.TotalFiles, len(cfg.CustomFolders), len(cfg.Extensions))
	
	if monitoring {
		text += "[green]🟢 АКТИВЕН[-]"
	} else {
		text += "[red]🔴 ОСТАНОВЛЕН[-]"
	}
	
	text += "\n\n[white]РАСШИРЕНИЯ:[-]\n"
	
	for ext, count := range stats.ByExtension {
		percentage := 0
		if stats.TotalFiles > 0 {
			percentage = (count * 100) / stats.TotalFiles
		}
		bar := strings.Repeat("█", percentage/5) + strings.Repeat("░", 20-percentage/5)
		text += fmt.Sprintf("  [cyan]%s[-]: %d [gray]%s %d%%[-]\n", ext, count, bar, percentage)
	}
	
	modal.SetText(text).
		AddButtons([]string{"[white]Закрыть[-]", "[green]Экспорт[-]", "[yellow]Обновить[-]"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "[yellow]Обновить[-]" {
				showStatistics()
			} else if buttonLabel == "[green]Экспорт[-]" {
				exportStatistics()
			} else {
				pages.HidePage("statistics")
			}
		})
	
	modal.SetBorder(true).
		SetTitle(" [yellow]📈 СТАТИСТИКА[-] ").
		SetBackgroundColor(tcell.ColorBlack)
	
	pages.AddPage("statistics", modal, true, true)
}

func showSettings() {
// Создаем форму
form := tview.NewForm().
AddInputField("Папка для логов:", cfg.LogDir, 40, nil, nil).
AddInputField("Сводный файл:", cfg.SummaryFile, 40, nil, nil).
AddInputField("Интервал (сек):", "5", 10, nil, nil).
AddCheckbox("Автозапуск", false, nil).
AddCheckbox("Подсветка", true, nil).
AddButton("[green]Сохранить[-]", func() {
addLogEntry("Настройки сохранены")
pages.HidePage("settings")
}).
AddButton("[red]Отмена[-]", func() {
pages.HidePage("settings")
})

form.SetBorder(true).
SetTitle(" [yellow]🛠 НАСТРОЙКИ[-] ").
SetBackgroundColor(tcell.ColorBlack)

center := tview.NewFlex().
AddItem(nil, 0, 1, false).
AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
AddItem(nil, 0, 1, false).
AddItem(form, 12, 0, true).
AddItem(nil, 0, 1, false), 60, 0, true).
AddItem(nil, 0, 1, false)

pages.AddPage("settings", center, true, true)
	// Создаем форму внутри функции
	form := tview.NewForm().
		AddInputField("Папка для логов:", cfg.LogDir, 40, nil, nil).
		AddInputField("Сводный файл:", cfg.SummaryFile, 40, nil, nil).
		AddInputField("Интервал (сек):", "5", 10, nil, nil).
		AddCheckbox("Автозапуск", false, nil).
		AddCheckbox("Подсветка", true, nil).
		AddButton("[green]Сохранить[-]", func() {
			addLogEntry("Настройки сохранены")
			pages.HidePage("settings")
		}).
		AddButton("[red]Отмена[-]", func() {
			pages.HidePage("settings")
		})
	
	form.SetBorder(true).
		SetTitle(" [yellow]🛠 НАСТРОЙКИ[-] ").
		SetBackgroundColor(tcell.ColorBlack)
	
	center := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(form, 12, 0, true).
			AddItem(nil, 0, 1, false), 60, 0, true).
		AddItem(nil, 0, 1, false)
	
	pages.AddPage("settings", center, true, true)
}

func deleteSelectedFolder() {
	if len(cfg.CustomFolders) == 0 {
		showErrorDialog("Ошибка", "Нет папок для удаления")
		return
	}
	
	idx := rightPane.GetCurrentItem()
	if idx >= 2 && idx-2 < len(cfg.CustomFolders) {
		folderIdx := idx - 2
		folder := cfg.CustomFolders[folderIdx]
		
		modal := tview.NewModal().
			SetText(fmt.Sprintf("Удалить папку?\n\n[yellow]%s[-]", folder)).
			AddButtons([]string{"[green]Да[-]", "[red]Нет[-]"}).
			SetDoneFunc(func(buttonIndex int, buttonLabel string) {
				if buttonLabel == "[green]Да[-]" {
					removed := cfg.CustomFolders[folderIdx]
					cfg.CustomFolders = append(cfg.CustomFolders[:folderIdx], cfg.CustomFolders[folderIdx+1:]...)
					
					cfg.TotalFiles = countAllTrackedFiles()
					
					addLogEntry(fmt.Sprintf("Удалена папка: %s", removed))
					
					updateRightPane()
					updateTopPanel(nil)
				}
				pages.HidePage("confirmDelete")
			})
		
		modal.SetBorder(true).
			SetTitle(" [yellow]🗑 УДАЛЕНИЕ[-] ").
			SetBackgroundColor(tcell.ColorBlack)
		
		pages.AddPage("confirmDelete", modal, true, true)
	}
}

func manageFolder(index int) {
	if index < len(cfg.CustomFolders) {
		folder := cfg.CustomFolders[index]
		
		modal := tview.NewModal().
			SetText(fmt.Sprintf("Папка: [yellow]%s[-]\n\nДействие:", folder)).
			AddButtons([]string{"[yellow]Сканировать[-]", "[red]Удалить[-]", "[white]Отмена[-]"}).
			SetDoneFunc(func(buttonIndex int, buttonLabel string) {
				switch buttonLabel {
				case "[yellow]Сканировать[-]":
					count := countFilesInFolder(folder)
					addLogEntry(fmt.Sprintf("Сканирование %s: %d файлов", folder, count))
					cfg.TotalFiles = countAllTrackedFiles()
					updateTopPanel(nil)
				case "[red]Удалить[-]":
					deleteFolder(index)
				}
				pages.HidePage("manageFolder")
			})
		
		modal.SetBorder(true).
		SetTitle(" [yellow]📂 ПАПКА[-] ").
		SetBackgroundColor(tcell.ColorBlack)
		
		pages.AddPage("manageFolder", modal, true, true)
	}
}

func deleteFolder(index int) {
	folder := cfg.CustomFolders[index]
	
	modal := tview.NewModal().
		SetText(fmt.Sprintf("Удалить папку?\n\n[yellow]%s[-]", folder)).
		AddButtons([]string{"[green]Да[-]", "[red]Нет[-]"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "[green]Да[-]" {
				cfg.CustomFolders = append(cfg.CustomFolders[:index], cfg.CustomFolders[index+1:]...)
				addLogEntry(fmt.Sprintf("Удалена папка: %s", folder))
				
				cfg.TotalFiles = countAllTrackedFiles()
				
				updateRightPane()
				updateTopPanel(nil)
			}
			pages.HidePage("deleteFolder")
		})
	
	modal.SetBorder(true).
		SetTitle(" [yellow]🗑 УДАЛЕНИЕ[-] ").
		SetBackgroundColor(tcell.ColorBlack)
	
	pages.AddPage("deleteFolder", modal, true, true)
}

func showErrorDialog(title, message string) {
	modal := tview.NewModal().
		SetText(fmt.Sprintf("[red]%s[-]\n\n%s", title, message)).
		AddButtons([]string{"[white]OK[-]"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			pages.HidePage("error")
		})
	
	modal.SetBorder(true).
		SetTitle(" [red]⚠ ОШИБКА[-] ").
		SetBackgroundColor(tcell.ColorBlack)
	
	pages.AddPage("error", modal, true, true)
}

func addLogEntry(message string) {
	timestamp := time.Now().Format("15:04:05")
	logEntry := fmt.Sprintf("[gray]%s[-] %s\n", timestamp, message)
	
	currentText := logView.GetText(true)
	logView.SetText(currentText + logEntry)
	
	logView.ScrollToEnd()
	
	logToFile(message)
}

func logToFile(message string) {
	dateStr := time.Now().Format("2006-01-02")
	logFile := filepath.Join(cfg.LogDir, dateStr+"_changes.md")
	
	entry := fmt.Sprintf("### %s\n", time.Now().Format("15:04:05"))
	cleanMessage := strings.ReplaceAll(message, "[", "")
	cleanMessage = strings.ReplaceAll(cleanMessage, "]", "")
	entry += fmt.Sprintf("- **Событие:** %s\n", cleanMessage)
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
		return "[gray]Логи отсутствуют[/]"
	}
	
	lines := strings.Split(string(content), "\n")
	
	start := len(lines) - count
	if start < 0 {
		start = 0
	}
	
	var result strings.Builder
	for i := start; i < len(lines); i++ {
		line := lines[i]
		result.WriteString(line + "\n")
	}
	
	return result.String()
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
	exportContent += fmt.Sprintf("Дата: %s\n\n", time.Now().Format("2006-01-02 15:04:05"))
	
	exportContent += fmt.Sprintf("Всего файлов: %d\n", stats.TotalFiles)
	exportContent += fmt.Sprintf("Папок: %d\n", len(cfg.CustomFolders))
	exportContent += fmt.Sprintf("Расширений: %d\n\n", len(cfg.Extensions))
	
	exportContent += "По расширениям:\n"
	for ext, count := range stats.ByExtension {
		percentage := 0
		if stats.TotalFiles > 0 {
			percentage = (count * 100) / stats.TotalFiles
		}
		exportContent += fmt.Sprintf("  %s: %d файлов (%d%%)\n", ext, count, percentage)
	}
	
	filename := fmt.Sprintf("statistics_%s.txt", time.Now().Format("20060102_150405"))
	err := os.WriteFile(filename, []byte(exportContent), 0644)
	
	if err != nil {
		addLogEntry(fmt.Sprintf("Ошибка экспорта: %v", err))
	} else {
		addLogEntry(fmt.Sprintf("Статистика экспортирована в %s", filename))
	}
}
