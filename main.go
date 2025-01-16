package main

import (
	"context"
	"flag"
	"fmt"
	"h2img/internal/imageutils"
	"image/png"
	"log/slog"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
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

			for {
				select {
				case <-ctx.Done():
					slog.Warn("stopping work", "reason", ctx.Err())
					return
				case event, ok := <-watcher.Events:
					if !ok {
						return
					}
					logger.Debug("received watcher event", "event", event)
					if !proceed(event) {
						continue
					}
					heicFilePath := event.Name
					if err := convertHeicToPng(heicFilePath, *outputDir); err != nil {
						slog.Error("cannot convert HEIC to PNG", "path", heicFilePath, "err", err)
						continue
					}
					logger.Debug("converted HEIC to PNG", "outputDir", *outputDir)
				case err, ok := <-watcher.Errors:
					if !ok {
						return
					}
					slog.Error("watcher error", "err", err)
				}
			}
		}()
	}

	wg.Wait()
}

func proceed(event fsnotify.Event) bool {
	return !event.Has(fsnotify.Create) &&
		strings.ToLower(filepath.Ext(event.Name)) != ".heic"
}

func convertHeicToPng(heicFilePath, outputDir string) error {
	heicImage, err := imageutils.NewHeicDecoder().DecodeFromFile(heicFilePath)
	if err != nil {
		return errors.Wrapf(err, "cannot decode HEIC image from file '%s'", heicFilePath)
	}

	baseFile := path.Base(heicFilePath)
	baseName := strings.TrimSuffix(baseFile, filepath.Ext(heicFilePath))
	outputFileName := baseName + ".png"
	outputFilePath := path.Join(outputDir, outputFileName)

	outputFile, err := os.Create(outputFilePath)
	if err != nil {
		return errors.Wrapf(err, "cannot create output file '%s'", outputFilePath)
	}
	defer outputFile.Close()

	if err := png.Encode(outputFile, heicImage); err != nil {
		return errors.Wrapf(err, "cannot save HEIC image to output file '%s'", outputFilePath)
	}
	return nil
}
