# Git Commit Exporter

The Git Commit Exporter is a command-line tool written in Go that exports commit information from one or more Git repositories to an XML file that can be read by Excel to create a graph or table. 

## Features

- Exports commit information from one or more Git repositories within a specified time range
- Extracts ClickUp task IDs from commit messages and generates links to ClickUp tasks
- Exports commit information to an XML file that can be read by Excel to create a graph or table

## Usage

Flags:
- `-start` - Start date of the time range to export commits (format: YYYY-MM-DD)
- `-end` - End date of the time range to export commits (format: YYYY-MM-DD)
- `-repos` - Comma separated paths to the repositories you want to export from

Example usage:

`gitdocs -start 2023-01-01 -end 2023-03-10 -repos="/path/to/repo1,/path/to/repo2"`

## Dependencies

- Git
- Go 1.16+

## License

This tool is licensed under the MIT license. See the `LICENSE` file for more details.