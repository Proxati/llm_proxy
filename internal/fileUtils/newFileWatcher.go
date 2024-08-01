package fileUtils

import (
	"errors"
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/fsnotify/fsnotify"
)

// NewFileWatcher creates a new fsnotify.Watcher and adds the specified directory to it.
//
// Example usage:
//
//	watch, err := NewFileWatcher(dirName)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	err = WaitForFile(watch, time.Second*5) // err is set if timeout was reached
//	if err != nil {
//	    log.Fatal(err)
//	}
func NewFileWatcher(logger *slog.Logger, watchDir string) (*fsnotify.Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	err = watcher.Add(watchDir)
	if err != nil {
		watcher.Close()
		return nil, err
	}
	logger.Debug("Watching", "watchDir", watchDir)
	return watcher, nil
}

// WaitForFile waits for a file event to occur in the watcher
func WaitForFile(logger *slog.Logger, watcher *fsnotify.Watcher, timeout time.Duration) error {
	defer watcher.Close()
	logger = logger.WithGroup("fileUtils")
	logger.Debug("Waiting for file event", "timeout", timeout)

	// Channel to signal when a file event occurs
	eventOccurred := make(chan string)
	errorChannel := make(chan error)

	// Goroutine to handle fsnotify events
	go func() {
		defer close(eventOccurred)
		defer close(errorChannel)

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					logger.Debug("Watcher events channel closed")
					return
				}
				log.Printf("Event received: %v", event)
				if event.Op&(fsnotify.Create|fsnotify.Write|fsnotify.Rename|fsnotify.Remove) != 0 {
					logger.Debug("Relevant event received", "event.Name", event.Name)
					eventOccurred <- event.Name
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				logger.Debug("Watcher error", "error", err)
				errorChannel <- err
			}
		}
	}()

	// Wait for a relevant file event or timeout
	select {
	case eventName := <-eventOccurred:
		logger.Debug("Relevant file event received", "eventName", eventName)
		return nil
	case err := <-errorChannel:
		return fmt.Errorf("error from watcher: %w", err)
	case <-time.After(timeout):
		return errors.New("timeout waiting for file event")
	}
}
