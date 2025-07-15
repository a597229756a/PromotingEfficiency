// monitor.go
package file

import (
	"os"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

type FileMonitor struct {
	watchDir string
	watcher  *fsnotify.Watcher
	lastFile string
	lastMod  time.Time
	mu       sync.Mutex
}

func NewFileMonitor(dir string) (*FileMonitor, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	if err := watcher.Add(dir); err != nil {
		return nil, err
	}

	return &FileMonitor{
		watchDir: dir,
		watcher:  watcher,
	}, nil
}

func (m *FileMonitor) Watch(handler func(string)) error {
	for {
		select {
		case event, ok := <-m.watcher.Events:
			if !ok {
				return nil
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				info, err := os.Stat(event.Name)
				if err != nil {
					continue
				}

				m.mu.Lock()
				if info.ModTime().After(m.lastMod) {
					m.lastMod = info.ModTime()
					m.lastFile = event.Name
					go handler(event.Name)
				}
				m.mu.Unlock()
			}
		case err, ok := <-m.watcher.Errors:
			if !ok {
				return nil
			}
			return err
		}
	}
}
