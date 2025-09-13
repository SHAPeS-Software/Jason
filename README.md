# Jason

Jason is a lightweight CLI-based JSON viewer and editor written in Go. It lets you view JSON files in a neat, SQL-like table format and edit them with undo/redo capabilities. Perfect for quick JSON tweaks in the terminal!

## Building

Here’s how to build and run Jason on your system.

### Prerequisites
- **Go**: Version 1.16 or higher (You got that)
- **Git**: Optional, for cloning the repository.
- A terminal or command prompt.

### Build Steps
1. **Clone the Repository** (if using Git):
   ```bash
   git clone https://github.com/SHAPeS-Software/Jason.git
   cd Jason
Or download and extract the source code to a directory named Jason.

Install the Dependency:
Jason uses the github.com/peterh/liner package for its interactive shell prompt. So initlize a project and get it:
go mod init example.com
go mod tidy
go get github.com/peterh/liner

Ensure the Source Code:
Make sure the jason.go file is in the Jason directory.
Build the Program:
Compile the Go code to create the jason executable:
go build jason.go
This creates a jason binary (or jason.exe on Windows) in the current directory.
Add to PATH (Optional):
To run jason from any directory, move the binary to a directory in your PATH:

Linux/macOS:
mv jason /usr/local/bin/

Windows: Move jason.exe to a directory like C:\Program Files\Jason and add it to your system PATH via environment variables.

# Commands

Non-Shell Commands:

jason open <file.json>: Opens a JSON file and starts the shell.
jason help: Shows available non-shell commands.
jason end: Exits the program.


Shell Commands:

read: Displays JSON in a table.
write <index1:index2> <content>: Edits a field (e.g., write user:name Jane Doe).
undo: Reverts the last change.
redo: Reapplies an undone change.
create <filename>: Creates a new JSON file.
open <file.json>: Opens another JSON file.
remove <file.json>: Marks a file for removal (pending save).
quit: Exits without saving.
squit: Saves changes and exits.
DevOutputLogging: Toggles verbose logging.



That’s It!
Jason’s simple, so it's just for messing with JSON files in the terminal. Modify the code, play around, and have fun! If you run into issues, check that your JSON files are valid and the jason binary is accessible. Feel free to post a Issue to.
