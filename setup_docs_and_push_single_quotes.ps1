# setup_docs_and_push_single_quotes.ps1

# --- Пути к файлам ---
$repoPath = 'F:\AI-LAN\MyProject' # Убедитесь, что это правильный путь к вашему репозиторию
$logFile = Join-Path $repoPath 'DEVELOPMENT_LOG.md'
$contribFile = Join-Path $repoPath 'CONTRIBUTING.md'
$readmeFile = Join-Path $repoPath 'README.md'

# --- Содержимое файлов ---
# Используем одинарные кавычки. Внутри них () и "" не интерпретируются.
$logLines = @(
    '# Development Log for AILAN Archivist',
    '',
    '## 1. Initial Setup and Goals (January 11, 2026)',
    '',
    '**Goal:** Create a file archivist that monitors directories for changes, logs them in Markdown, and integrates with Git for version control and remote storage.',
    '',
    '**Initial Steps:**',
    '- Created `build_archivist_multi.ps1` to generate Go code.',
    '- Encountered issues with PowerShell syntax when embedding Go code.',
    '- Resolved by using `Get-Location` instead of `$PSScriptRoot`.',
    '- Created `code_library.txt` and `code_library.md` for managing reusable PowerShell snippets.',
    '',
    '## 2. Integration with Git (January 11, 2026)',
    '',
    '**Goal:** Use GitHub as a remote repository for storing the archivist''s logs and code.',
    '',
    '**Steps:**',
    '- Set up a local Git repository in `F:\AI-LAN\MyProject`.',
    '- Configured global Git user info (`user.email`, `user.name`).', # <-- Теперь в одинарных кавычках
    '- Added `.go` files from `AILAN-Archivist-Go` to the repository.',
    '- Pushed changes to `https://github.com/DiscoverVLZ/MyProject-Archive.git`.',
    '',
    '**Issues:**',
    '- `fatal: not a git repository` — resolved by initializing the repo.',
    '- `error: src refspec main does not match any` — resolved by making an initial commit.',
    '- `error: failed to push some refs` — resolved by setting upstream branch with `git push -u origin master:main`.',
    '',
    '## 3. Building the Archivist (January 11, 2026)',
    '',
    '**Goal:** Compile the Go-based archivist with Git integration.',
    '',
    '**Steps:**',
    '- Attempted to compile `archivist_with_git.go` using `go build`.',
    '- Encountered dependency issues with `termdash` and `gomponents`.',
    '- Updated dependencies with `go get` and `go mod tidy`.',
    '- Fixed import paths for `maragu.dev/gomponents`.',
    '',
    '**Issues:**',
    '- API changes in `termdash` v0.20.0 broke TUI functionality.',
    '- Errors related to `textinput.Width`, `container.SplitVertical`, `t.Keyboard()`, etc.',
    '',
    '**Decision:**',
    '- Postpone TUI development.',
    '- Focus on creating a **console-only archivist** with full monitoring and Git integration.',
    '',
    '## 4. Next Steps (January 11, 2026)',
    '',
    '**Immediate Goal:** Create a console-based archivist that:',
    '- Monitors directories for changes.',
    '- Logs events to Markdown files.',
    '- Automatically commits changes to Git.',
    '',
    '**Documentation:**',
    '- Create `CONTRIBUTING.md` to guide future development.',
    '- Update `README.md` with setup instructions.',
    '',
    '**Future Work:**',
    '- Revisit TUI after stabilizing core functionality.',
    '- Add more advanced features (e.g., email notifications, GUI).',
    ''
)

$contribLines = @(
    '# Contributing to AILAN Archivist',
    '',
    'Thank you for your interest in contributing to the AILAN Archivist project!',
    '',
    '## Getting Started',
    '',
    '1. **Clone the Repository:**',
    '   ```bash',
    '   git clone https://github.com/DiscoverVLZ/MyProject-Archive.git',
    '   cd MyProject-Archive',
    '   ```',
    '',
    '2. **Set Up Your Environment:**',
    '   - Install [Go](https://go.dev/dl/) (version 1.21 or later).',
    '   - Ensure you have Git installed and configured.',
    '',
    '3. **Build the Project:**',
    '   ```bash',
    '   go build -o AILAN-Archivist-With-Git.exe archivist_with_git.go',
    '   ```',
    '',
    '## Project Structure',
    '',
    '- `archivist_with_git.go`: Main source file for the console-based archivist.',
    '- `DEVELOPMENT_LOG.md`: History of all development activities.',
    '- `README.md`: Overview and setup instructions.',
    '',
    '## How to Contribute',
    '',
    '1. **Fork the Repository** on GitHub.',
    '2. **Create a New Branch** for your feature or fix.',
    '3. **Make Your Changes** and test them locally.',
    '4. **Commit Your Changes** with clear and concise messages.',
    '5. **Push to Your Branch** and create a Pull Request.',
    '',
    '## Code Style',
    '',
    '- Follow Go conventions.',
    '- Keep functions short and focused.',
    '- Add comments for complex logic.',
    '',
    '## Reporting Issues',
    '',
    'If you encounter any problems, please open an issue with:',
    '- A clear description of the problem.',
    '- Steps to reproduce.',
    '- Expected vs. actual behavior.',
    '- Any error messages or logs.',
    ''
)

$readmeLines = @(
    '# AILAN Archivist',
    '',
    'A file archivist that monitors directories for changes and logs them with Git integration.',
    '',
    '## Features',
    '',
    '- Monitors specified directories for file changes.',
    '- Logs events in Markdown format.',
    '- Automatically commits changes to Git.',
    '',
    '## Requirements',
    '',
    '- Go 1.21+',
    '- Git',
    '',
    '## Installation',
    '',
    '1. Clone the repository:',
    '   ```bash',
    '   git clone https://github.com/DiscoverVLZ/MyProject-Archive.git',
    '   cd MyProject-Archive',
    '   ```',
    '',
    '2. Build the project:',
    '   ```bash',
    '   go build -o AILAN-Archivist-With-Git.exe archivist with git.go', # <-- Исправлено для корректного имени файла
    '   ```',
    '',
    '## Usage',
    '',
    '1. Run the executable:',
    '   ```bash',
    '   .\AILAN-Archivist-With-Git.exe',
    '   ```',
    '',
    '2. Configure the directories to monitor in `archivist_config.json`.',
    '',
    '## Configuration',
    '',
    'Edit `archivist_config.json` to specify:',
    '- `WatchedDirectories`: List of directories to monitor.',
    '- `DocsDirectory`: Directory for logs and state file.',
    '- `AllowedExtensions`: File extensions to track.',
    '',
    '## Contributing',
    '',
    'See [CONTRIBUTING.md](CONTRIBUTING.md) for details on how to contribute.',
    '',
    '## License',
    '',
    'This project is licensed under the MIT License.',
    ''
)

$logContent = $logLines -join "`n"
$contribContent = $contribLines -join "`n"
$readmeContent = $readmeLines -join "`n"

# --- Проверка существования папки репозитория ---
if (-not (Test-Path $repoPath)) {
    Write-Error 'Repository path does not exist: $repoPath'
    exit 1
}

# --- Перейти в папку репозитория ---
Set-Location -Path $repoPath

# --- Создание файлов ---
Write-Host 'Creating DEVELOPMENT_LOG.md...'
$null = New-Item -ItemType File -Path $logFile -Value $logContent -Force
Write-Host "Created: $logFile"

Write-Host 'Creating CONTRIBUTING.md...'
$null = New-Item -ItemType File -Path $contribFile -Value $contribContent -Force
Write-Host "Created: $contribFile"

Write-Host 'Creating README.md...'
$null = New-Item -ItemType File -Path $readmeFile -Value $readmeContent -Force
Write-Host "Created: $readmeFile"

# --- Git команды ---
Write-Host 'Adding files to Git index...'
& git add $logFile $contribFile $readmeFile
if ($LASTEXITCODE -ne 0) { Write-Error 'Git add failed'; exit 1 }

Write-Host 'Committing changes...'
& git commit -m 'Add documentation files for project development'
if ($LASTEXITCODE -ne 0) { Write-Error 'Git commit failed'; exit 1 }

Write-Host 'Pushing changes to GitHub...'
& git push origin master
if ($LASTEXITCODE -ne 0) { Write-Error 'Git push failed'; exit 1 }

Write-Host 'All files created and pushed successfully!' -ForegroundColor Green
Write-Host 'Now you can proceed with creating the console archivist.' -ForegroundColor Yellow
