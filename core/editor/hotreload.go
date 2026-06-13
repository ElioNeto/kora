package editor

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ─── File Watcher ────────────────────────────────────────────────────────────

// WatchEvent represents a file change detected by the watcher.
type WatchEvent struct {
	Path      string    // full path of the changed file
	Op        WatchOp   // what happened
	Timestamp time.Time // when it happened
}

// WatchOp describes the type of file operation.
type WatchOp int

const (
	WatchOpCreate WatchOp = iota
	WatchOpModify
	WatchOpDelete
)

// HotReloadState manages the KScript hot-reload system.
type HotReloadState struct {
	mu           sync.Mutex
	enabled      bool
	rootDir      string          // project root directory to watch
	ksFiles      map[string]bool // tracked .ks files
	lastBuild    time.Time       // last successful build time
	buildErrors  []string        // errors from last build
	compileFn    func(path string) error // callback to compile a .ks file
	onChangeFn   func(events []WatchEvent) // callback when changes are detected
	stopChan     chan struct{}
	eventLog     []LogEntry
	eventLimit   int
}

// LogEntry is a single entry in the hot-reload event log.
type LogEntry struct {
	Time    time.Time
	Message string
	IsError bool
}

// NewHotReloadState creates a new hot-reload state.
func NewHotReloadState(rootDir string) *HotReloadState {
	return &HotReloadState{
		rootDir:    rootDir,
		ksFiles:    make(map[string]bool),
		eventLimit: 50,
		stopChan:   make(chan struct{}),
	}
}

// SetCompileFn sets the function that compiles a single .ks file.
func (hr *HotReloadState) SetCompileFn(fn func(path string) error) {
	hr.mu.Lock()
	defer hr.mu.Unlock()
	hr.compileFn = fn
}

// SetOnChangeFn sets the callback for when file changes are detected.
func (hr *HotReloadState) SetOnChangeFn(fn func(events []WatchEvent)) {
	hr.mu.Lock()
	defer hr.mu.Unlock()
	hr.onChangeFn = fn
}

// Enable starts watching for KScript file changes.
func (hr *HotReloadState) Enable() {
	hr.mu.Lock()
	if hr.enabled {
		hr.mu.Unlock()
		return
	}
	hr.enabled = true
	hr.scanKScriptFiles()
	hr.mu.Unlock()

	// Start goroutine OUTSIDE the lock
	go hr.pollLoop()

	hr.addLog("Hot-reload ativado — monitorando arquivos .ks")
}

// Disable stops the file watcher.
func (hr *HotReloadState) Disable() {
	hr.mu.Lock()
	if !hr.enabled {
		hr.mu.Unlock()
		return
	}
	hr.enabled = false
	close(hr.stopChan)
	hr.stopChan = make(chan struct{})
	hr.mu.Unlock()

	hr.addLog("Hot-reload desativado")
}

// IsEnabled returns whether hot-reload is active.
func (hr *HotReloadState) IsEnabled() bool {
	hr.mu.Lock()
	defer hr.mu.Unlock()
	return hr.enabled
}

// BuildErrors returns the errors from the last build.
func (hr *HotReloadState) BuildErrors() []string {
	hr.mu.Lock()
	defer hr.mu.Unlock()
	result := make([]string, len(hr.buildErrors))
	copy(result, hr.buildErrors)
	return result
}

// EventLog returns the recent event log entries.
func (hr *HotReloadState) EventLog() []LogEntry {
	hr.mu.Lock()
	defer hr.mu.Unlock()
	result := make([]LogEntry, len(hr.eventLog))
	copy(result, hr.eventLog)
	return result
}

// ForceRecompile triggers a recompilation of all tracked .ks files.
func (hr *HotReloadState) ForceRecompile() {
	hr.mu.Lock()
	compileFn := hr.compileFn
	ksFiles := make([]string, 0, len(hr.ksFiles))
	for path := range hr.ksFiles {
		ksFiles = append(ksFiles, path)
	}
	hr.mu.Unlock()

	if compileFn == nil {
		return
	}

	hr.addLog("Recompilando todos os arquivos KScript...")
	errors := make([]string, 0)
	for _, path := range ksFiles {
		if err := compileFn(path); err != nil {
			errors = append(errors, filepath.Base(path)+": "+err.Error())
		}
	}

	hr.mu.Lock()
	hr.buildErrors = errors
	hr.lastBuild = time.Now()
	hr.mu.Unlock()

	if len(errors) > 0 {
		for _, err := range errors {
			hr.addLogErr(err)
		}
	} else {
		hr.addLog("Compilação OK — " + time.Now().Format("15:04:05"))
	}
}

// ─── Internal ────────────────────────────────────────────────────────────────

func (hr *HotReloadState) scanKScriptFiles() {
	filepath.Walk(hr.rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".ks") {
			hr.ksFiles[path] = true
		}
		return nil
	})
}

func (hr *HotReloadState) pollLoop() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	// Track file modification times
	modTimes := make(map[string]time.Time)

	// Initial scan
	filepath.Walk(hr.rootDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && strings.HasSuffix(path, ".ks") {
			modTimes[path] = info.ModTime()
		}
		return nil
	})

	for {
		select {
		case <-ticker.C:
			hr.checkForChanges(modTimes)
		case <-hr.stopChan:
			return
		}
	}
}

func (hr *HotReloadState) checkForChanges(modTimes map[string]time.Time) {
	hr.mu.Lock()
	compileFn := hr.compileFn
	onChangeFn := hr.onChangeFn
	hr.mu.Unlock()

	if compileFn == nil {
		return
	}

	events := make([]WatchEvent, 0)

	// Scan for new and modified files
	filepath.Walk(hr.rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".ks") {
			return nil
		}

		lastMod, known := modTimes[path]
		if !known {
			// New file
			modTimes[path] = info.ModTime()
			events = append(events, WatchEvent{Path: path, Op: WatchOpCreate, Timestamp: time.Now()})
			hr.mu.Lock()
			hr.ksFiles[path] = true
			hr.mu.Unlock()
		} else if info.ModTime().After(lastMod) {
			// Modified file
			modTimes[path] = info.ModTime()
			events = append(events, WatchEvent{Path: path, Op: WatchOpModify, Timestamp: time.Now()})
		}
		return nil
	})

	if len(events) == 0 {
		return
	}

	// Notify callback
	if onChangeFn != nil {
		onChangeFn(events)
	}

	// Auto-compile modified files
	errors := make([]string, 0)
	for _, ev := range events {
		if ev.Op == WatchOpDelete {
			continue
		}
		hr.addLog("Mudança detectada: " + filepath.Base(ev.Path))

		if err := compileFn(ev.Path); err != nil {
			errors = append(errors, filepath.Base(ev.Path)+": "+err.Error())
		}
	}

	hr.mu.Lock()
	hr.buildErrors = errors
	hr.lastBuild = time.Now()
	hr.mu.Unlock()

	if len(errors) > 0 {
		for _, err := range errors {
			hr.addLogErr(err)
		}
	}
}

func (hr *HotReloadState) addLog(msg string) {
	hr.mu.Lock()
	defer hr.mu.Unlock()
	hr.eventLog = append(hr.eventLog, LogEntry{Time: time.Now(), Message: msg})
	if len(hr.eventLog) > hr.eventLimit {
		hr.eventLog = hr.eventLog[len(hr.eventLog)-hr.eventLimit:]
	}
	log.Println("[HotReload]", msg)
}

func (hr *HotReloadState) addLogErr(msg string) {
	hr.mu.Lock()
	defer hr.mu.Unlock()
	hr.eventLog = append(hr.eventLog, LogEntry{Time: time.Now(), Message: msg, IsError: true})
	if len(hr.eventLog) > hr.eventLimit {
		hr.eventLog = hr.eventLog[len(hr.eventLog)-hr.eventLimit:]
	}
	log.Println("[HotReload ERROR]", msg)
}

// ─── Build Result ────────────────────────────────────────────────────────────

// BuildResult represents the result of a KScript compilation.
type BuildResult struct {
	Success  bool
	Duration time.Duration
	Errors   []string
	Files    []string
}
