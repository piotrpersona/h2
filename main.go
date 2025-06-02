package main

import (
	"flag"
	"fmt"

	"log/slog"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/schollz/progressbar/v3"
	"golang.org/x/sync/errgroup"
)

func exit(err error) {
	if err != nil {
		slog.Error("terminating...", "err", err)
		os.Exit(1)
	}
}

func main() {
	inputDir := flag.String("input", "", "input directory with heic files")
	outputDir := flag.String("output", "", "output directory with png files")
	desiredExtension := flag.String("ext", "png", "desired extension to convert heic to")
	flag.Parse()
	slog.SetLogLoggerLevel(slog.LevelDebug)

	if inputDir == nil || *inputDir == "" {
		slog.Warn("input dir must be provided")
		os.Exit(1)
	}
	if outputDir == nil || *outputDir == "" {
		slog.Warn("output dir must be provided")
		os.Exit(1)
	}

	heicExtension := ".heic"

	inputFiles := make([]string, 0)

	err := filepath.Walk(*inputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), heicExtension) {
			inputFiles = append(inputFiles, path)
		}
		return nil
	})
	if err != nil {
		slog.Warn("error while reading heic files")
		os.Exit(1)
	}

	pbar := progressbar.NewOptions(len(inputFiles), progressbar.OptionSetDescription("Converting"))

	var eg errgroup.Group
	eg.SetLimit(4)

	for _, inputFile := range inputFiles {
		eg.Go(func() error {
			defer pbar.Add(1)
			return convertHeicToPng(inputFile, *outputDir, *desiredExtension)
		})
	}
	if err := eg.Wait(); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	slog.Info("done!")
}

func convertHeicToPng(heicFilePath, outputDir, desiredExtension string) error {
	baseFile := path.Base(heicFilePath)
	baseName := strings.TrimSuffix(baseFile, filepath.Ext(heicFilePath))

	outputFileName := fmt.Sprintf("%s.%s", baseName, desiredExtension)
	outputFilePath := path.Join(outputDir, outputFileName)

	slog.Debug("converting heic image", "extension", desiredExtension, "heic", heicFilePath, "output", outputFilePath)

	return exec.Command("magick", heicFilePath, outputFilePath).Run()
}
