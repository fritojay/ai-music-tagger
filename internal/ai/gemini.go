package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"reflect"
	"strings"

	"github.com/fritojay/ai-music-tagger/internal/audio"
	"github.com/google/generative-ai-go/genai"
	"github.com/googleapis/gax-go/v2/apierror"
	"google.golang.org/api/option"
)

const GeminiQueryPrefix = "You are an API that returns the year a song or album was released. I will give you a json object that contains the artist name and the title or album name of the year I need. Respond in the format without any markdown or additional characters. JSON Only. {'year':'2025-06-06'}."

type GeminiClient struct {
	client *genai.Client
}

type TagResponse struct {
	Year string `json:"year"`
}

func NewGeminiClient() (*GeminiClient, error) {
	ctx := context.Background()
	geminiAPIKey := os.Getenv("GEMINI_API_KEY")
	if geminiAPIKey == "" {
		return nil, errors.New("unable to get api key")
	}
	client, err := genai.NewClient(ctx, option.WithAPIKey(geminiAPIKey))
	if err != nil {
		return nil, err
	}
	return &GeminiClient{
		client: client,
	}, nil
}

func (c *GeminiClient) QueryForTags(audioFiles []*audio.AudioFile) {
	ctx := context.Background()
	model := c.client.GenerativeModel("gemini-2.0-flash")
	model.SystemInstruction = &genai.Content{
		Parts: []genai.Part{
			genai.Text(GeminiQueryPrefix),
		},
	}
	for _, file := range audioFiles {
		b, err := file.ToJSONQuery()
		if err != nil {
			continue
		}
		resp, err := model.GenerateContent(ctx, genai.Text(string(b)))
		if err != nil {
			slog.Error("unable to query gemini", "error", err, "type", reflect.TypeOf(err))
			if apiErr, ok := err.(*apierror.APIError); ok && apiErr.HTTPCode() == 429 {
				return
			}
			continue
		}
		tagResponse, err := extractMetaFromLLM(resp)
		if err != nil {
			slog.Error("unable to parse meta", "error", err, "file", file.File)
			continue
		}
		file.Year = tagResponse.Year
	}

}

func extractMetaFromLLM(resp *genai.GenerateContentResponse) (*TagResponse, error) {
	if len(resp.Candidates) == 0 {
		return nil, errors.New("no candidates returned")
	}
	if len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, errors.New("no parts returned")
	}
	content := resp.Candidates[0].Content.Parts[0]
	textPart, ok := content.(genai.Text)
	if !ok {
		return nil, errors.New("no text returned")
	}
	trimmedString := strings.TrimSpace(string(textPart))
	var tagResponse TagResponse
	err := json.Unmarshal([]byte(trimmedString), &tagResponse)
	if err != nil {
		return nil, fmt.Errorf("no json returned %s: %v", string(trimmedString), err)
	}
	return &tagResponse, nil
}
