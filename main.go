package main

import (
	"flag"
	"log/slog"
	"os"

	"github.com/fritojay/ai-music-tagger/internal/ai"
	"github.com/fritojay/ai-music-tagger/internal/audio"
)

func main() {
	dir := flag.String("dir", ".", "Directory to search for MP3 files")
	dryRun := flag.Bool("dryRun", false, "Whether to modify the tags or not")
	flag.Parse()
	files, err := audio.ListMP3Files(*dir)
	if err != nil {
		slog.Error("issue listing files", "error", err)
		os.Exit(1)
	}
	tracksNeedYear := make([]*audio.AudioFile, 0)
	for _, file := range files {
		audioFile, err := audio.ParseMP3(file)
		if err != nil {
			slog.Error("error reading tags", "error", err, "file", file)
			continue
		}
		if audioFile.Year == "" {
			tracksNeedYear = append(tracksNeedYear, audioFile)
		}
	}

	for _, audioFile := range tracksNeedYear {
		slog.Info("file", "audio", audioFile)
	}
	if *dryRun {
		slog.Info("Skipping AI Lookup..")
		return
	}
	client, err := ai.NewGeminiClient()
	if err != nil {
		slog.Error("unable to create gemini client", "error", err)
		os.Exit(1)
	}

	modified := client.QueryForTags(tracksNeedYear)

	for _, file := range modified {
		if file.Year != "" {
			err := audio.ModifyYear(file.File, file.Year)
			if err != nil {
				slog.Error("unable to add year to file", "file", file)
			}
		}
	}
}
