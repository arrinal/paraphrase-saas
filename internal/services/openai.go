package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/arrinal/paraphrase-saas/internal/config"
)

type OpenAIService struct {
	apiKey string
}

type OpenAIRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type ParaphraseResponse struct {
	DetectedLanguage string
	ParaphrasedText  string
}

func NewOpenAIService(cfg *config.Config) *OpenAIService {
	if cfg.OpenAIKey == "" {
		log.Fatal("OpenAI API key is not set")
	}
	return &OpenAIService{
		apiKey: cfg.OpenAIKey,
	}
}

func (s *OpenAIService) Paraphrase(text, language, style string) (*ParaphraseResponse, error) {
	// Add style-specific instructions
	styleGuide := ""
	switch style {
	case "standard":
		styleGuide = `
Additional style guide:
- Use clear and straightforward language
- Maintain a balanced tone that's neither too formal nor too casual
- Focus on clarity and precision
- Keep sentences well-structured but not overly complex`

	case "formal":
		styleGuide = `
Additional style guide:
- Maintain a professional and respectful tone
- Use precise vocabulary and avoid colloquialisms
- Avoid contractions and slang
- Follow conventional business writing structure
- Ensure clarity while maintaining professionalism`

	case "academic":
		styleGuide = `
Additional style guide:
- Use objective and analytical tone
- Incorporate specialized terminology appropriately
- Focus on evidence-based statements
- Maintain scholarly conventions and formal structure
- Avoid personal opinions and emotional language
- Use precise and technical vocabulary where appropriate`

	case "casual":
		styleGuide = `
Additional style guide:
- Use a friendly and conversational tone
- Include simple and straightforward language
- Write as if talking to a friend
- Feel free to use common expressions
- Keep sentences shorter and more direct
- Make the text relatable and easy to understand`

	case "creative":
		styleGuide = `
Additional style guide:
- Use expressive and imaginative language
- Incorporate figurative language and metaphors where appropriate
- Vary sentence rhythm and structure for effect
- Create vivid descriptions and engaging narrative flow
- Use unique and evocative vocabulary
- Focus on creating memorable and impactful expressions`
	}

	var prompt string
	if language == "auto" {
		prompt = fmt.Sprintf(`
You are an expert writer specializing in text paraphrasing and language detection.
First, detect the language of the text and then paraphrase it in the same language.

1. First line of your response must be "DETECTED_LANGUAGE: [language name in English]"
2. Second line must be empty
3. From third line onwards, provide the paraphrased text following these rules:

Now you are an expert writer specializing in text paraphrasing. Rewrite the following text in the detected language using a %s style. 
Make sure to make substantial structural changes by following these steps:

1. Reorder the sequence of ideas and rearrange paragraphs or sections.
2. Split long sentences into shorter ones, and combine short sentences into more complex structures.
3. Vary the sentence starters and use different transitions to change the flow of the text.
4. Alter the sentence patterns and adjust the logical flow of information.
5. Every quote from a person (commonly marked with double quotes, or someone said) keep it as it is, don't paraphrase it.
6. Never ignore line part that have sentence like chapter, subchapter, title, subtitle, etc.
7. If the original text structure based on chapter, subchapter, title, subtitle, etc, do above steps for each chapter, subchapter, title, subtitle, etc.

%s

The goal is to create a rewritten version that looks structurally different, with minimal overlap in wording, especially if analyzed with a diff checker, and apply similar structural changes. 
Only return the rewritten text without quotation marks at the beginning and end.

Text: %s`, style, styleGuide, text)
	} else {
		// Use existing prompt for known language
		prompt = fmt.Sprintf(`
Now you are an expert writer specializing in text paraphrasing. Rewrite the following text in %s language using a %s style. 
Make sure to make substantial structural changes by following these steps:

1. Reorder the sequence of ideas and rearrange paragraphs or sections.
2. Split long sentences into shorter ones, and combine short sentences into more complex structures.
3. Vary the sentence starters and use different transitions to change the flow of the text.
4. Alter the sentence patterns and adjust the logical flow of information.
5. Every quote from a person (commonly marked with double quotes, or someone said) keep it as it is, don't paraphrase it.
6. Never ignore line part that have sentence like chapter, subchapter, title, subtitle, etc.
7. If the original text structure based on chapter, subchapter, title, subtitle, etc, do above steps for each chapter, subchapter, title, subtitle, etc.

%s

The goal is to create a rewritten version that looks structurally different, with minimal overlap in wording, especially if analyzed with a diff checker, and apply similar structural changes. 
Only return the rewritten text without quotation marks at the beginning and end.

Text: %s`, language, style, styleGuide, text)
	}

	request := OpenAIRequest{
		Model: "chatgpt-4o-latest",
		Messages: []Message{
			{Role: "user", Content: prompt},
		},
		Temperature: 0.7,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare request: %v", err)
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	var response OpenAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	if response.Error != nil {
		return nil, fmt.Errorf("OpenAI API error: %s", response.Error.Message)
	}

	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI")
	}

	// Parse response for auto-detect case
	if language == "auto" {
		lines := strings.Split(response.Choices[0].Message.Content, "\n")
		if len(lines) < 3 {
			return nil, fmt.Errorf("invalid response format")
		}

		// Extract detected language
		langLine := strings.TrimPrefix(lines[0], "DETECTED_LANGUAGE: ")
		detectedLanguage := strings.TrimSpace(langLine)

		// Get paraphrased text (everything after the second line)
		paraphrasedText := strings.Join(lines[2:], "\n")

		return &ParaphraseResponse{
			DetectedLanguage: detectedLanguage,
			ParaphrasedText:  paraphrasedText,
		}, nil
	}

	// For non-auto cases, return the specified language
	return &ParaphraseResponse{
		DetectedLanguage: language,
		ParaphrasedText:  response.Choices[0].Message.Content,
	}, nil
}

func (s *OpenAIService) ParaphraseText(text string) (string, error) {
	// Implement OpenAI API call logic here
	// For now, return a placeholder
	return "Paraphrased: " + text, nil
}

// GetDetectedLanguage returns the detected language from the paraphrased response
func (s *OpenAIService) GetDetectedLanguage(text string) (string, error) {
	// Use the same Paraphrase function with auto-detect
	response, err := s.Paraphrase(text, "auto", "standard")
	if err != nil {
		return "", err
	}

	// Extract language from the first line
	lines := strings.Split(response.ParaphrasedText, "\n")
	if len(lines) < 1 {
		return "", fmt.Errorf("invalid response format")
	}

	langLine := strings.TrimPrefix(lines[0], "DETECTED_LANGUAGE: ")
	return strings.TrimSpace(langLine), nil
}
