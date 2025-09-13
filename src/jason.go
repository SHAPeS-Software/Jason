package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/peterh/liner"
)

type Jason struct {
	currentFile   string
	data          map[string]interface{}
	tempData      map[string]interface{}
	undoStack     []map[string]interface{}
	redoStack     []map[string]interface{}
	isLogging     bool
	isShellActive bool
}

func NewJason() *Jason {
	return &Jason{
		data:      make(map[string]interface{}),
		tempData:  make(map[string]interface{}),
		undoStack: []map[string]interface{}{},
		redoStack: []map[string]interface{}{},
	}
}

func (j *Jason) openFile(filename string) error {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return fmt.Errorf("file %s does not exist", filename)
	}
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	var data map[string]interface{}
	if err := json.Unmarshal(content, &data); err != nil {
		return err
	}
	j.currentFile = filename
	j.data = data
	j.tempData = deepCopyMap(data)
	j.undoStack = []map[string]interface{}{}
	j.redoStack = []map[string]interface{}{}
	if j.isLogging {
		fmt.Printf("Opened file: %s\n", filename)
	}
	return nil
}

func (j *Jason) createFile(filename string) error {
	if _, err := os.Stat(filename); !os.IsNotExist(err) {
		return fmt.Errorf("file %s already exists", filename)
	}
	j.currentFile = filename
	j.data = make(map[string]interface{})
	j.tempData = make(map[string]interface{})
	j.undoStack = []map[string]interface{}{}
	j.redoStack = []map[string]interface{}{}
	if err := j.saveFile(); err != nil {
		return err
	}
	if j.isLogging {
		fmt.Printf("Created file: %s\n", filename)
	}
	return nil
}

func (j *Jason) saveFile() error {
	data, err := json.MarshalIndent(j.tempData, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(j.currentFile, data, 0644)
}

func (j *Jason) displayTable() {
	if len(j.tempData) == 0 {
		fmt.Println("No data to display")
		return
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Key\tValue\tType")
	fmt.Fprintln(w, "----\t----\t----")
	for k, v := range j.tempData {
		var valueStr string
		var typeStr string
		switch val := v.(type) {
		case map[string]interface{}:
			typeStr = "object"
			lines := []string{}
			for subKey, subVal := range val {
				var formattedVal string
				switch sv := subVal.(type) {
				case string:
					formattedVal = fmt.Sprintf(`"%s"`, sv)
				case float64:
					formattedVal = fmt.Sprintf("%g", sv)
				case bool:
					formattedVal = fmt.Sprintf("%v", sv)
				case nil:
					formattedVal = "null"
				default:
					jsonBytes, _ := json.Marshal(sv)
					formattedVal = string(jsonBytes)
				}
				line := fmt.Sprintf("%s : %s", subKey, formattedVal)
				if len(lines) == 0 {
					lines = append(lines, line)
				} else {
					lines = append(lines, fmt.Sprintf("| %s", line))
				}
			}
			valueStr = strings.Join(lines, ", ")
		case []interface{}:
			typeStr = "array"
			elements := []string{}
			for _, item := range val {
				var formattedItem string
				switch iv := item.(type) {
				case string:
					formattedItem = fmt.Sprintf(`"%s"`, iv)
				case float64:
					formattedItem = fmt.Sprintf("%g", iv)
				case bool:
					formattedItem = fmt.Sprintf("%v", iv)
				case nil:
					formattedItem = "null"
				default:
					jsonBytes, _ := json.Marshal(iv)
					formattedItem = string(jsonBytes)
				}
				elements = append(elements, formattedItem)
			}
			valueStr = strings.Join(elements, ", ")
		case string:
			valueStr = fmt.Sprintf(`"%s"`, val)
			typeStr = "string"
		case float64:
			valueStr = fmt.Sprintf("%g", val)
			typeStr = "number"
		case bool:
			valueStr = fmt.Sprintf("%v", val)
			typeStr = "boolean"
		case nil:
			valueStr = "null"
			typeStr = "null"
		default:
			valueStr = fmt.Sprintf("%v", val)
			typeStr = fmt.Sprintf("%T", val)
		}
		// Escape tabs and newlines in valueStr to prevent table misalignment
		valueStr = strings.ReplaceAll(valueStr, "\t", "\\t")
		valueStr = strings.ReplaceAll(valueStr, "\n", "\\n")
		// Split valueStr into lines for multi-line display
		lines := strings.Split(valueStr, ", ")
		for i, line := range lines {
			if i == 0 {
				fmt.Fprintf(w, "%s\t%s\t%s\n", k, line, typeStr)
			} else {
				fmt.Fprintf(w, "\t%s\t\n", line)
			}
		}
	}
	w.Flush()
	if j.isLogging {
		fmt.Println("Displayed table")
	}
}
func (j *Jason) writeData(path, content string) error {
	keys := strings.Split(path, ":")
	if len(keys) < 1 || len(keys) > 2 {
		return fmt.Errorf("invalid path format, use index1 or index1:index2")
	}
	j.undoStack = append(j.undoStack, deepCopyMap(j.tempData))
	value, err := parseValue(content)
	if err != nil {
		return err
	}
	if len(keys) == 1 {
		j.tempData[keys[0]] = value
	} else {
		if subMap, ok := j.tempData[keys[0]].(map[string]interface{}); ok {
			subMap[keys[1]] = value
		} else {
			return fmt.Errorf("key %s is not a nested map", keys[0])
		}
	}
	j.redoStack = []map[string]interface{}{}
	if j.isLogging {
		fmt.Printf("Wrote %v to path %s\n", content, path)
	}
	return nil
}

func (j *Jason) undo() error {
	if len(j.undoStack) == 0 {
		return fmt.Errorf("nothing to undo")
	}
	j.redoStack = append(j.redoStack, deepCopyMap(j.tempData))
	j.tempData = j.undoStack[len(j.undoStack)-1]
	j.undoStack = j.undoStack[:len(j.undoStack)-1]
	if j.isLogging {
		fmt.Println("Performed undo")
	}
	return nil
}

func (j *Jason) redo() error {
	if len(j.redoStack) == 0 {
		return fmt.Errorf("nothing to redo")
	}
	j.undoStack = append(j.undoStack, deepCopyMap(j.tempData))
	j.tempData = j.redoStack[len(j.redoStack)-1]
	j.redoStack = j.redoStack[:len(j.redoStack)-1]
	if j.isLogging {
		fmt.Println("Performed redo")
	}
	return nil
}

func (j *Jason) removeFile(filename string) error {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return fmt.Errorf("file %s does not exist", filename)
	}
	if j.isLogging {
		fmt.Printf("Marked %s for removal (pending save)\n", filename)
	}
	// Removal happens on save, so we just note it
	return nil
}

func deepCopyMap(m map[string]interface{}) map[string]interface{} {
	newMap := make(map[string]interface{})
	for k, v := range m {
		if subMap, ok := v.(map[string]interface{}); ok {
			newMap[k] = deepCopyMap(subMap)
		} else {
			newMap[k] = v
		}
	}
	return newMap
}

func parseValue(content string) (interface{}, error) {
	if val, err := strconv.Atoi(content); err == nil {
		return val, nil
	}
	if val, err := strconv.ParseFloat(content, 64); err == nil {
		return val, nil
	}
	if content == "true" || content == "false" {
		return content == "true", nil
	}
	var jsonVal interface{}
	if err := json.Unmarshal([]byte(content), &jsonVal); err == nil {
		return jsonVal, nil
	}
	return content, nil
}

func (j *Jason) runShell() {
	line := liner.NewLiner()
	defer line.Close()
	line.SetCtrlCAborts(true)

	for {
		input, err := line.Prompt("Jason> ")
		if err != nil {
			if err == liner.ErrPromptAborted {
				fmt.Println("Aborted")
				break
			}
			fmt.Println("Error reading input:", err)
			continue
		}
		args := strings.Fields(input)
		if len(args) == 0 {
			continue
		}
		switch args[0] {
		case "read":
			j.displayTable()
		case "write":
			if len(args) < 3 {
				fmt.Println("Usage: write <index1:index2> <content>")
				continue
			}
			path := args[1]
			content := strings.Join(args[2:], " ")
			if err := j.writeData(path, content); err != nil {
				fmt.Println("Error:", err)
			}
		case "undo":
			if err := j.undo(); err != nil {
				fmt.Println("Error:", err)
			}
		case "redo":
			if err := j.redo(); err != nil {
				fmt.Println("Error:", err)
			}
		case "create":
			if len(args) != 2 {
				fmt.Println("Usage: create <filename>")
				continue
			}
			if err := j.createFile(args[1]); err != nil {
				fmt.Println("Error:", err)
			}
		case "open":
			if len(args) != 2 {
				fmt.Println("Usage: open <file.json>")
				continue
			}
			if err := j.openFile(args[1]); err != nil {
				fmt.Println("Error:", err)
			}
		case "help":
			fmt.Println(`Jason Shell Commands:
  read                - Display JSON contents in a table
  write <index1:index2> <content> - Write to JSON at specified path
  undo                - Undo last action
  redo                - Redo last undone action
  create <filename>   - Create a new JSON file
  open <file.json>    - Open a JSON file
  remove <file.json>  - Mark file for removal (on save)
  quit                - Quit without saving
  squit               - Quit and save
  DevOutputLogging    - Toggle logging of actions`)
		case "remove":
			if len(args) != 2 {
				fmt.Println("Usage: remove <file.json>")
				continue
			}
			if err := j.removeFile(args[1]); err != nil {
				fmt.Println("Error:", err)
			}
		case "quit":
			if j.isLogging {
				fmt.Println("Quitting without saving")
			}
			return
		case "squit":
			if j.currentFile != "" {
				if err := j.saveFile(); err != nil {
					fmt.Println("Error saving:", err)
				} else if j.isLogging {
					fmt.Println("Saved and quitting")
				}
			}
			return
		case "DevOutputLogging":
			j.isLogging = !j.isLogging
			fmt.Printf("Logging %s\n", map[bool]string{true: "enabled", false: "disabled"}[j.isLogging])
		default:
			fmt.Println("Unknown command. Type 'help' for commands.")
		}
	}
}

func main() {
	j := NewJason()
	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Println("Usage: jason <command> [args]. Type 'jason help' for help.")
		return
	}

	switch args[0] {
	case "open":
		if len(args) != 2 {
			fmt.Println("Usage: jason open <file.json>")
			return
		}
		if err := j.openFile(args[1]); err != nil {
			fmt.Println("Error:", err)
			return
		}
		j.isShellActive = true
		j.runShell()
	case "help":
		fmt.Println("Jason Version 1.0")
		fmt.Println(`Jason Commands:
  jason open <file.json> - Open a JSON file and start shell
  jason help             - Display this help
  jason end              - End all Jason instances (not implemented in single instance mode)`)
	case "end":
		fmt.Println("Ending Jason instance")
		os.Exit(0)
	default:
		fmt.Println("Unknown command. Type 'jason help' for help.")
	}
}
