package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
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

func NewOpenAIService(cfg *config.Config) *OpenAIService {
	if cfg.OpenAIKey == "" {
		log.Fatal("OpenAI API key is not set")
	}
	return &OpenAIService{
		apiKey: cfg.OpenAIKey,
	}
}

func (s *OpenAIService) detectLanguage(text string) (string, error) {
	request := OpenAIRequest{
		Model: "gpt-3.5-turbo",
		Messages: []Message{
			{Role: "user", Content: fmt.Sprintf("What language is this text written in? Just respond with the language name in English. Text: %s", text)},
		},
		Temperature: 0.3, // Lower temperature for more consistent results
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to prepare request: %v", err)
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	var response OpenAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to parse response: %v", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI")
	}

	detectedLanguage := response.Choices[0].Message.Content
	// Clean up and standardize the language name
	detectedLanguage = strings.TrimSpace(detectedLanguage)
	detectedLanguage = strings.Title(strings.ToLower(detectedLanguage))

	return detectedLanguage, nil
}

func (s *OpenAIService) Paraphrase(text, language, style string) (string, error) {
	var detectedLanguage string
	var err error

	if language == "auto" {
		detectedLanguage, err = s.detectLanguage(text)
		if err != nil {
			log.Printf("Failed to detect language: %v", err)
			return "", fmt.Errorf("failed to detect language: %v", err)
		}
		language = detectedLanguage
	}

	prompt := fmt.Sprintf(
		"Paraphrase the following text in %s language using a %s style. Text: %s",
		language,
		style,
		text,
	)

	request := OpenAIRequest{
		Model: "gpt-3.5-turbo",
		Messages: []Message{
			{Role: "user", Content: prompt},
		},
		Temperature: 0.7,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		log.Printf("Error marshaling request: %v", err)
		return "", fmt.Errorf("failed to prepare request: %v", err)
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error making request: %v", err)
		return "", fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
		return "", fmt.Errorf("failed to read response: %v", err)
	}

	var response OpenAIResponse
	if err := json.Unmarshal(body, &response); err != nil {
		log.Printf("Error unmarshaling response: %v", err)
		return "", fmt.Errorf("failed to parse response: %v", err)
	}

	// Check for OpenAI API error
	if response.Error != nil {
		log.Printf("OpenAI API error: %s", response.Error.Message)
		return "", fmt.Errorf("OpenAI API error: %s", response.Error.Message)
	}

	if len(response.Choices) == 0 {
		log.Printf("No choices in response")
		return "", fmt.Errorf("no response from OpenAI")
	}

	return response.Choices[0].Message.Content, nil
}

func (s *OpenAIService) ParaphraseText(text string) (string, error) {
	// Implement OpenAI API call logic here
	// For now, return a placeholder
	return "Paraphrased: " + text, nil
}

// Add this method to get the detected language
func (s *OpenAIService) GetDetectedLanguage(text string) (string, error) {
	return s.detectLanguage(text)
}
