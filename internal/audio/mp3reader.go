package audio

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/bogem/id3v2"
)

type AudioFile struct {
	Title  string `json:"title"`
	Artist string `json:"artist"`
	Year   string `json:"year"`
	Album  string `json:"album"`
	File   string `json:"-"`
}

func (f *AudioFile) ToJSONQuery() ([]byte, error) {
	return json.Marshal(f)
}

func ParseMP3(file string) (*AudioFile, error) {
	tag, err := id3v2.Open(file, id3v2.Options{Parse: true})
	if err != nil {
		// Handle ID3 tag errors. For example, the file might not have a tag.
		// You might want to log the error or just skip the file.
		return nil, err // Skip the file
	}
	defer tag.Close()
	return &AudioFile{
		Year:   tag.Year(),
		Artist: tag.Artist(),
		Title:  tag.Title(),
		Album:  tag.Album(),
		File:   file,
	}, nil
}

func ModifyYear(file string, year string) error {
	tag, err := id3v2.Open(file, id3v2.Options{Parse: true})
	if err != nil {
		return err
	}
	defer tag.Close()
	tag.SetYear(year)
	return tag.Save()
}

// TODO: MP3 for now. Create interface for supporting other music tags.
func ListMP3Files(dir string) ([]string, error) {
	info, err := os.Stat(dir)
	if err != nil {
		return nil, fmt.Errorf("unable to stat directory %s", dir)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", dir)
	}

	var mp3Files []string

	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err // Handle errors during traversal
		}
		if !info.IsDir() && filepath.Ext(path) == ".mp3" {
			mp3Files = append(mp3Files, path)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking directory: %w", err)
	}

	return mp3Files, nil
}
