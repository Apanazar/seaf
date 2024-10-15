package ui

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type GameResult struct {
	Success bool
	Message string
}

type model struct {
	cursorX         int
	cursorY         int
	sequence        []string
	message         string
	grid            [][]string
	isVertical      bool
	firstPick       bool
	targetSequences [][]string
	success         bool
}

var possibleValues = []string{"BD", "1C", "E9", "55", "7A"}

// // random Sequence generates a random sequence of a given maximum length
func randomSequence(maxLength int) []string {
	rand.Seed(time.Now().UnixNano())
	length := rand.Intn(maxLength-1) + 1
	sequence := make([]string, length)
	for i := range sequence {
		sequence[i] = possibleValues[rand.Intn(len(possibleValues))]
	}
	return sequence
}

// // initial Model initializes the game model
func initialModel() model {
	grid := make([][]string, 4)
	for i := range grid {
		grid[i] = make([]string, 4)
		for j := range grid[i] {
			grid[i][j] = possibleValues[rand.Intn(len(possibleValues))]
		}
	}

	targetSequences := [][]string{
		randomSequence(3),
		randomSequence(4),
		randomSequence(5),
	}

	return model{
		cursorX:         0,
		cursorY:         0,
		sequence:        []string{},
		message:         "",
		grid:            grid,
		isVertical:      true,
		firstPick:       true,
		targetSequences: targetSequences,
		success:         false,
	}
}

// Init initializes the game launch command
func (m model) Init() tea.Cmd {
	return nil
}

// Update handles keyboard events and updates the game state
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "up":
			if m.firstPick || m.isVertical {
				if m.cursorY > 0 {
					m.cursorY--
					// Skipping empty cells
					for m.cursorY > 0 && m.grid[m.cursorY][m.cursorX] == "  " {
						m.cursorY--
					}
				}
			}

		case "down":
			if m.firstPick || m.isVertical {
				if m.cursorY < len(m.grid)-1 {
					m.cursorY++
					// Skipping empty cells
					for m.cursorY < len(m.grid)-1 && m.grid[m.cursorY][m.cursorX] == "  " {
						m.cursorY++
					}
				}
			}

		case "left":
			if m.firstPick || !m.isVertical {
				if m.cursorX > 0 {
					m.cursorX--
					// Skipping empty cells
					for m.cursorX > 0 && m.grid[m.cursorY][m.cursorX] == "  " {
						m.cursorX--
					}
				}
			}

		case "right":
			if m.firstPick || !m.isVertical {
				if m.cursorX < len(m.grid[0])-1 {
					m.cursorX++
					// Skipping empty cells
					for m.cursorX < len(m.grid[0])-1 && m.grid[m.cursorY][m.cursorX] == "  " {
						m.cursorX++
					}
				}
			}

		case "enter":
			if m.grid[m.cursorY][m.cursorX] != "  " {
				m.sequence = append(m.sequence, m.grid[m.cursorY][m.cursorX])
				m.grid[m.cursorY][m.cursorX] = "  " // Emptying the selected cell
				if m.firstPick {
					m.firstPick = false // Switch to a strict sequence of movement after the first selection
				}
				m.isVertical = !m.isVertical // // Switching to another advance mode

				if m.checkSequences() {
					if m.success {
						// Success, we finish the game
						return m, tea.Quit
					} else {
						// // Losing, clearing the screen and completing the attempt
						return m, tea.Batch(
							clearScreenCmd(),
							tea.Quit,
						)
					}
				}
			}
		}
	}

	return m, nil
}

func clearScreenCmd() tea.Cmd {
	return func() tea.Msg {
		var cmd *exec.Cmd
		switch runtime.GOOS {
		case "windows":
			cmd = exec.Command("cmd", "/c", "cls")
		default:
			cmd = exec.Command("clear")
		}
		cmd.Stdout = os.Stdout
		cmd.Run()
		return nil
	}
}

// checkSequences checks whether the current sequence matches one of the targets
func (m *model) checkSequences() bool {
	// Check for a match with any of the target sequences
	for _, targetSeq := range m.targetSequences {
		if len(m.sequence) == len(targetSeq) {
			match := true
			for i := 0; i < len(m.sequence); i++ {
				if m.sequence[i] != targetSeq[i] {
					match = false
					break
				}
			}
			if match {
				m.message = "Success! You have put together one of the combinations."
				m.success = true
				return true
			}
		}
	}

	// Check if the current sequence is a prefix of any of the targets
	for _, targetSeq := range m.targetSequences {
		if len(m.sequence) < len(targetSeq) {
			prefix := true
			for i := 0; i < len(m.sequence); i++ {
				if m.sequence[i] != targetSeq[i] {
					prefix = false
					break
				}
			}
			if prefix {
				return false // We continue the game, give a new attempt
			}
		}
	}

	// If the current sequence does not match and is not a prefix, it is a loss
	m.message = "A loss! The combination does not match any of the targets."
	m.success = false
	return true
}

// View returns a visual representation of the game
func (m model) View() string {
	var b strings.Builder
	// ANSI color codes
	blue := "\033[1;34m"
	yellow := "\033[1;33m"
	green := "\033[1;32m"
	reset := "\033[0m"

	for y, row := range m.grid {
		for x, val := range row {
			var formattedVal string
			if x == m.cursorX && y == m.cursorY {
				formattedVal = fmt.Sprintf("%s[%s]%s ", reset, val, reset)
			} else {
				formattedVal = fmt.Sprintf("%s %s  %s", blue, val, reset)
			}
			fmt.Fprintf(&b, "%s", formattedVal)
		}
		b.WriteString("\n")
	}

	// Display target sequences each on a new line in yellow
	b.WriteString("\n")
	for _, seq := range m.targetSequences {
		fmt.Fprintf(&b, "%s%v%s\n", yellow, seq, reset)
	}

	// Display current sequence and message
	fmt.Fprintf(&b, "\n%s%v%s\n%s", green, m.sequence, reset, m.message)
	return b.String()
}

// RunGame starts the game and returns the result
func RunGame() GameResult {
	p := tea.NewProgram(initialModel())

	// Using StartReturningModel to get the final model
	finalModel, err := p.StartReturningModel()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Game launch error: %v\n", err)
		return GameResult{Success: false, Message: "Game launch error"}
	}

	// Check that the final model has the right type
	m, ok := finalModel.(model)
	if !ok {
		return GameResult{Success: false, Message: "Incorrect state of the game model"}
	}

	return GameResult{
		Success: m.success,
		Message: m.message,
	}
}
