package ui

import (
	"sync"

	"github.com/gosuri/uiprogress"
)

// ProgressUI manages the progress bar
type ProgressUI struct {
	bar *uiprogress.Bar
	mu  sync.Mutex
}

// NewProgressUI initializes the progress bar with the specified total number of steps
func NewProgressUI(total int) *ProgressUI {
	uiprogress.Start()
	bar := uiprogress.AddBar(total).AppendCompleted().PrependElapsed()
	bar.Width = 50
	return &ProgressUI{bar: bar}
}

// Increment increases the progress bar by one step
func (p *ProgressUI) Increment() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.bar.Incr()
}

// Finish stops the progress bar
func (p *ProgressUI) Finish() {
	uiprogress.Stop()
}
