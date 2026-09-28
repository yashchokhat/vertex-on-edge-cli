package ui

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// Spinner represents an animated progress indicator running in a goroutine.
type Spinner struct {
	message string
	active  bool
	stopCh  chan struct{}
	doneCh  chan struct{}
	mu      sync.Mutex
}

// SpinnerStart starts a spinner and returns a handle to stop it.
func SpinnerStart(message string) *Spinner {
	s := &Spinner{
		message: message,
		active:  true,
		stopCh:  make(chan struct{}),
		doneCh:  make(chan struct{}),
	}
	go s.run()
	return s
}

func (s *Spinner) run() {
	defer close(s.doneCh)

	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	frameIdx := 0
	ticker := time.NewTicker(80 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.mu.Lock()
			msg := s.message
			s.mu.Unlock()
			ClearLine()
			fmt.Fprintf(os.Stdout, "    %s  %s",
				spinnerAccent.Render(frames[frameIdx]),
				textStyle.Render(msg),
			)
			frameIdx = (frameIdx + 1) % len(frames)
		}
	}
}

// Update changes the spinner's message.
func (s *Spinner) Update(message string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.message = message
}

// Stop halts the spinner and shows a success or failure mark.
func (s *Spinner) Stop(success bool) {
	s.mu.Lock()
	if !s.active {
		s.mu.Unlock()
		return
	}
	s.active = false
	msg := s.message
	s.mu.Unlock()

	close(s.stopCh)
	<-s.doneCh // Wait for goroutine to finish

	mark := successMark
	if !success {
		mark = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true).Render("✕")
	}

	ClearLine()
	fmt.Fprintf(os.Stdout, "    %s  %s\n", mark, textStyle.Render(msg))
}
