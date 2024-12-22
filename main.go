package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"gopkg.in/gographics/imagick.v3/imagick"
)

func main() {
	inputDir := flag.String("input", "", "input directory to monitor")
	outputDir := flag.String("output", "", "output directory to monitor")
	flag.Parse()

	logsFileName := time.Now().Format("2006-01-02")
	logsFileName = fmt.Sprintf("/tmp/h2img_%s.log", logsFileName)

	if inputDir == nil || *inputDir == "" {
		log.Println("input dir must be provided")
		os.Exit(1)
	}
	if outputDir == nil || *outputDir == "" {
		log.Println("output dir must be provided")
		os.Exit(1)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	err = watcher.Add(*inputDir)
	if err != nil {
		log.Fatal(err)
	}

	imagick.Initialize()
	mw := imagick.NewMagickWand()
	defer mw.Destroy()

	doneCh := make(chan struct{})

	logFile, err := os.Create(logsFileName)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(logFile)
	go func() {
		defer close(doneCh)
		log.Printf("watching directory '%s", *inputDir)
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if !event.Has(fsnotify.Create) {
					continue
				}
				if strings.ToLower(filepath.Ext(event.Name)) != ".heic" {
					continue
				}
				heicFile := event.Name
				err := mw.ReadImage(heicFile)
				if err != nil {
					log.Fatal(err)
				}
				err = mw.SetImageFormat("png")
				if err != nil {
					log.Fatalf("Failed to set image format: %v", err)
				}

				baseFile := path.Base(heicFile)
				baseName := strings.TrimSuffix(baseFile, filepath.Ext(heicFile))
				outputFileName := baseName + ".png"
				outputFilePath := path.Join(*outputDir, outputFileName)

				err = mw.WriteImage(outputFilePath)
				if err != nil {
					log.Fatalf("Failed to write PNG image: %v", err)
				}
				log.Printf("converted '%s' to '%s'", heicFile, outputFilePath)
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	<-doneCh
}
