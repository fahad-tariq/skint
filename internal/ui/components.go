package ui

import (
	"fmt"
	"os"
	"strings"
	"sync/atomic"
	"time"
)

// Box draws a box around content
func Box(title string, width int) {
	if width < 10 {
		width = 40
	}

	inner := width - 2

	// Truncate title if too long
	if len(title) > inner-4 {
		title = title[:inner-7] + "..."
	}

	// Center title using raw (non-ANSI) title length
	titleLen := len(title)
	pad := (inner - titleLen) / 2
	rightPad := inner - pad - titleLen

	// Build horizontal line
	hline := strings.Repeat(Sym.BoxH, inner)

	// Print box to stderr
	fmt.Fprintf(os.Stderr, "%s%s%s\n", Sym.BoxTL, hline, Sym.BoxTR)
	fmt.Fprintf(os.Stderr, "%s%s%s%s%s\n", Sym.BoxV, strings.Repeat(" ", pad), Bold(title), strings.Repeat(" ", rightPad), Sym.BoxV)
	fmt.Fprintf(os.Stderr, "%s%s%s\n", Sym.BoxBL, hline, Sym.BoxBR)
}

// Separator draws a horizontal separator
func Separator(width int) {
	if width < 1 {
		width = 40
	}
	Dim("%s", strings.Repeat(Sym.BoxH, width))
	fmt.Fprintln(os.Stderr)
}

// ListItem prints a list item
func ListItem(checked bool, format string, a ...interface{}) {
	if checked {
		if Colors.Enabled {
			Colors.Green.Fprintf(os.Stderr, "  %s ", Sym.Check)
		} else {
			fmt.Fprintf(os.Stderr, "  %s ", Sym.Check)
		}
	} else {
		Dim("  %s ", Sym.Uncheck)
	}
	fmt.Fprintf(os.Stderr, format+"\n", a...)
}

// Prompt prints a prompt and returns user input
func Prompt(message, defaultValue string) string {
	promptText := message
	if defaultValue != "" {
		promptText = fmt.Sprintf("%s [%s]", message, defaultValue)
	}

	if Colors.Enabled {
		Colors.Cyan.Fprintf(os.Stderr, "%s: ", promptText)
	} else {
		fmt.Fprintf(os.Stderr, "%s: ", promptText)
	}

	var response string
	_, _ = fmt.Scanln(&response)

	if response == "" && defaultValue != "" {
		return defaultValue
	}

	return response
}

// Confirm asks for yes/no confirmation
func Confirm(message string, defaultYes bool) bool {
	hint := "[y/N]"
	if defaultYes {
		hint = "[Y/n]"
	}

	if Colors.Enabled {
		Colors.Cyan.Fprintf(os.Stderr, "%s %s: ", message, hint)
	} else {
		fmt.Fprintf(os.Stderr, "%s %s: ", message, hint)
	}

	var response string
	_, _ = fmt.Scanln(&response)

	if response == "" {
		return defaultYes
	}

	return strings.EqualFold(response, "y") || strings.EqualFold(response, "yes")
}

// ConfirmDanger asks for dangerous confirmation with phrase
func ConfirmDanger(action, phrase string) bool {
	fmt.Fprintln(os.Stderr)
	Box("DANGER", 40)
	fmt.Fprintln(os.Stderr)

	if Colors.Enabled {
		Colors.Red.Fprintln(os.Stderr, action)
	} else {
		fmt.Fprintln(os.Stderr, action)
	}

	fmt.Fprintln(os.Stderr)
	fmt.Fprintf(os.Stderr, "Type %s to confirm: ", Bold(phrase))

	var response string
	_, _ = fmt.Scanln(&response)

	return response == phrase
}

// Spinner is a loading spinner
type Spinner struct {
	message string
	stop    chan bool
	running atomic.Bool
}

// NewSpinner creates a new spinner
func NewSpinner(message string) *Spinner {
	return &Spinner{
		message: message,
		stop:    make(chan bool, 1),
	}
}

// Start starts the spinner
func (s *Spinner) Start() {
	if !Colors.Enabled {
		Info(s.message)
		return
	}

	s.running.Store(true)
	go func() {
		i := 0
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-s.stop:
				// Clear line
				fmt.Fprintf(os.Stderr, "\r\033[K")
				return
			case <-ticker.C:
				if Colors.Enabled {
					Colors.Blue.Fprintf(os.Stderr, "\r%s %s", Sym.Spinner[i%len(Sym.Spinner)], s.message)
				}
				i++
			}
		}
	}()
}

// Stop stops the spinner with a status
func (s *Spinner) Stop(success bool) {
	if !s.running.CompareAndSwap(true, false) {
		return
	}

	select {
	case s.stop <- true:
	default:
	}
	time.Sleep(50 * time.Millisecond) // Let the goroutine finish

	if success {
		Success("Done")
	} else {
		Error("Failed")
	}
}

// Table prints a simple table
func Table(headers []string, rows [][]string) {
	// Calculate column widths
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}

	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	// Print headers
	for i, h := range headers {
		fmt.Fprintf(os.Stderr, "%-*s  ", widths[i], Bold(h))
	}
	fmt.Fprintln(os.Stderr)

	// Print separator
	for i := range headers {
		fmt.Fprintf(os.Stderr, "%-*s  ", widths[i], strings.Repeat("-", widths[i]))
	}
	fmt.Fprintln(os.Stderr)

	// Print rows
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) {
				fmt.Fprintf(os.Stderr, "%-*s  ", widths[i], cell)
			}
		}
		fmt.Fprintln(os.Stderr)
	}
}

// MaskKey masks an API key for display.
// For short keys (12 chars or fewer), only asterisks are shown to avoid
// revealing most of the key.
func MaskKey(key string) string {
	if key == "" {
		return ""
	}
	if len(key) <= 12 {
		return "****"
	}
	return key[:4] + "****" + key[len(key)-4:]
}

// ErrorWithContext prints a detailed error with context
func ErrorWithContext(code, message, context, cause, solution string) {
	fmt.Fprintln(os.Stderr)

	if Colors.Enabled {
		Colors.Red.Fprint(os.Stderr, Bold("ERROR"))
		Dim(" [%s] ", code)
		Colors.Red.Fprintln(os.Stderr, Bold(message))
	} else {
		fmt.Fprintf(os.Stderr, "ERROR [%s] %s\n", code, message)
	}

	Dim("  Context:  %s\n", context)
	Dim("  Cause:    %s\n", cause)

	if Colors.Enabled {
		Colors.Cyan.Fprint(os.Stderr, "  Fix:      ")
		fmt.Fprintln(os.Stderr, solution)
	} else {
		fmt.Fprintf(os.Stderr, "  Fix:      %s\n", solution)
	}

	fmt.Fprintln(os.Stderr)
}

// NextSteps prints suggested next steps
func NextSteps(steps []string) {
	if !Colors.Enabled {
		fmt.Fprintln(os.Stderr, "\nNext:")
		for _, step := range steps {
			fmt.Fprintf(os.Stderr, "  %s %s\n", Sym.Arrow, step)
		}
		return
	}

	fmt.Fprintln(os.Stderr)
	Colors.Bold.Fprintln(os.Stderr, "Next:")
	for _, step := range steps {
		Colors.Cyan.Fprintf(os.Stderr, "  %s ", Sym.Arrow)
		fmt.Fprintln(os.Stderr, step)
	}
}
