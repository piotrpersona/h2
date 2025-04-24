package main

import (
	"context"
	"flag"
	"fmt"

	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

func exit(err error) {
	if err != nil {
		slog.Error("terminating...", "err", err)
		os.Exit(1)
	}
}

func main() {
	inputDir := flag.String("input", "", "input directory to monitor")
	outputDir := flag.String("output", "", "output directory to monitor")
	workersCount := flag.Int("workers", 8, "number of workers watching the directory")
	flag.Parse()
	slog.SetLogLoggerLevel(slog.LevelDebug)

	logsFileName := time.Now().Format("2006-01-02")
	logsFileName = fmt.Sprintf("/tmp/h2img_%s.log", logsFileName)

	if inputDir == nil || *inputDir == "" {
		slog.Warn("input dir must be provided")
		os.Exit(1)
	}
	if outputDir == nil || *outputDir == "" {
		slog.Warn("output dir must be provided")
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	watcher, err := fsnotify.NewWatcher()
	exit(err)
	defer watcher.Close()

	err = watcher.Add(*inputDir)
	exit(err)

	doneCh := make(chan struct{})

	logFile, err := os.Create(logsFileName)
	exit(err)
	defer logFile.Close()

	logger := slog.New(slog.NewTextHandler(logFile, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	logger = logger.With(
		slog.String("path", *inputDir),
	)

	defer close(doneCh)
	logger.Info("watching directory", "n_workers", *workersCount)

	var wg sync.WaitGroup

	for range *workersCount {
		wg.Add(1)
		go func() {
			defer wg.Done()

			startWorker(ctx, watcher, *outputDir)
		}()
	}

	wg.Wait()
}

func startWorker(ctx context.Context, watcher *fsnotify.Watcher, outputDir string) {
	for {
		select {
		case <-ctx.Done():
			slog.Warn("stopping work", "reason", ctx.Err())
			return
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			slog.Debug("received watcher event", "event", event)
			if !proceed(event) {
				continue
			}
			heicFilePath := event.Name
			if err := convertHeicToPng(heicFilePath, outputDir); err != nil {
				slog.Error("cannot convert HEIC to PNG", "path", heicFilePath, "err", err)
				continue
			}
			slog.Debug("converted HEIC to PNG", "outputDir", outputDir)
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			slog.Error("watcher error", "err", err)
		}
	}

}

func proceed(event fsnotify.Event) bool {
	return !event.Has(fsnotify.Create) &&
		strings.ToLower(filepath.Ext(event.Name)) != ".heic"
}

func convertHeicToPng(heicFilePath, outputDir string) error {
	baseFile := path.Base(heicFilePath)
	baseName := strings.TrimSuffix(baseFile, filepath.Ext(heicFilePath))

	outputFileName := baseName + ".png"
	outputFilePath := path.Join(outputDir, outputFileName)

	return exec.Command("magick", heicFilePath, outputFilePath).Err
}
