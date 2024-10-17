package watcher

import (
	"log/slog"
	"os"
	"time"

	"github.com/fsnotify/fsnotify"
)

func WatchFile(file string, action func() error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		slog.Error("Error creating watcher", "err", err)
		return
	}
	defer watcher.Close()

	err = watcher.Add(file)
	if err != nil {
		slog.Error("Error adding file to watcher", "err", err)
		return
	}
	slog.Info("Watching file", "file", file)

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			slog.Debug("Watcher event", "op", event.Op, "file", event.Name)

			if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create || event.Op&fsnotify.Rename == fsnotify.Rename {
				// Delay before re-adding to handle rename issues (apprently vim renames the file and overwrites somehing? not sure, but this seems to work)
				if event.Op&fsnotify.Rename == fsnotify.Rename {
					time.Sleep(100 * time.Millisecond)

					// Ensure file exists before re-adding to watcher
					if _, err := os.Stat(file); err == nil {
						err = watcher.Add(file)
						if err != nil {
							slog.Error("Error re-adding file to watcher", "err", err)
						}
					} else {
						slog.Error("File does not exist after rename; cannot re-add", "file", file)
					}
				}

				slog.Info("Watched file changed", "file", event.Name)
				err := action()
				if err != nil {
					slog.Error("Error executing action on file change", "err", err)
				} else {
					slog.Info("Action on fiel changes executed successfully", "file", file)
				}
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			slog.Error("Watcher error", "err", err)
		}
	}
}
