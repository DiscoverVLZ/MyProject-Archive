# AILAN Archivist

A file archivist that monitors directories for changes and logs them with Git integration.

## Features

- Monitors specified directories for file changes.
- Logs events in Markdown format.
- Automatically commits changes to Git.

## Requirements

- Go 1.21+
- Git

## Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/DiscoverVLZ/MyProject-Archive.git
   cd MyProject-Archive
   ```

2. Build the project:
   ```bash
   go build -o AILAN-Archivist-With-Git.exe archivist with git.go
   ```

## Usage

1. Run the executable:
   ```bash
   .\AILAN-Archivist-With-Git.exe
   ```

2. Configure the directories to monitor in `archivist_config.json`.

## Configuration

Edit `archivist_config.json` to specify:
- `WatchedDirectories`: List of directories to monitor.
- `DocsDirectory`: Directory for logs and state file.
- `AllowedExtensions`: File extensions to track.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for details on how to contribute.

## License

This project is licensed under the MIT License.
