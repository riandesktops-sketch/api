package client

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"google.golang.org/genai"
)

var (
	ErrAIUnavailable = errors.New("AI service unavailable")
	ErrInvalidPrompt = errors.New("invalid prompt")
)

// GeminiClient handles Gemini AI interactions
// Implements retry strategy with exponential backoff
type GeminiClient struct {
	client      *genai.Client
	model       string
	maxRetries  int
	baseDelay   time.Duration
}

// NewGeminiClient creates a new Gemini AI client
func NewGeminiClient(apiKey string) (*GeminiClient, error) {
	ctx := context.Background()
	
	// Create Gemini client with API key
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: apiKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	return &GeminiClient{
		client:     client,
		model:      "gemini-2.0-flash-exp",
		maxRetries: 3,
		baseDelay:  time.Second,
	}, nil
}

// GenerateContent generates content with retry strategy
// Retry strategy: 3 attempts with exponential backoff (1s, 2s, 4s)
// Reference: Pragmatic Programmer - Fail gracefully
func (c *GeminiClient) GenerateContent(ctx context.Context, prompt string) (string, error) {
	if prompt == "" {
		return "", ErrInvalidPrompt
	}

	var lastErr error
	
	for attempt := 0; attempt < c.maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff: 1s, 2s, 4s
			delay := c.baseDelay * time.Duration(1<<uint(attempt-1))
			log.Printf("Retry attempt %d after %v delay", attempt+1, delay)
			time.Sleep(delay)
		}

		// Set timeout for this attempt
		attemptCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		
		result, err := c.client.Models.GenerateContent(
			attemptCtx,
			c.model,
			genai.Text(prompt),
			nil,
		)
		cancel()

		if err == nil {
			return result.Text(), nil
		}

		lastErr = err
		log.Printf("Gemini API error (attempt %d/%d): %v", attempt+1, c.maxRetries, err)
	}

	// All retries failed, return fallback
	log.Printf("All Gemini API retries failed: %v", lastErr)
	return "", ErrAIUnavailable
}

// GenerateChatResponse generates AI chat response with zodiac persona
func (c *GeminiClient) GenerateChatResponse(ctx context.Context, zodiacSign, userMessage string) (string, error) {
	prompt := c.buildChatPrompt(zodiacSign, userMessage)
	
	response, err := c.GenerateContent(ctx, prompt)
	if err != nil {
		// Log the error for debugging
		log.Printf("⚠️ Gemini API failed, using fallback. Error: %v", err)
		// Fallback response if AI fails
		return c.getFallbackChatResponse(zodiacSign), nil
	}
	
	log.Printf("✅ Gemini API success for zodiac: %s", zodiacSign)
	return response, nil
}

// GenerateInsight generates life lesson insight from chat history
func (c *GeminiClient) GenerateInsight(ctx context.Context, chatHistory string) (string, error) {
	prompt := c.buildInsightPrompt(chatHistory)
	
	response, err := c.GenerateContent(ctx, prompt)
	if err != nil {
		// Fallback insight if AI fails
		return c.getFallbackInsight(), nil
	}
	
	return response, nil
}

// buildChatPrompt builds prompt for chat response
func (c *GeminiClient) buildChatPrompt(zodiacSign, userMessage string) string {
	traits := getZodiacTraits(zodiacSign)
	
	return fmt.Sprintf(`Kamu adalah AI companion yang ramah dan bisa diajak ngobrol santai.
Kamu punya sedikit karakteristik zodiak %s (%s), tapi jangan terlalu berlebihan atau alay.

Respon dengan natural seperti teman yang ngobrol biasa:
- Jangan terlalu formal atau kaku
- Jangan terlalu dramatis atau puitis
- Fokus pada apa yang user tanyakan/ceritakan
- Kasih respon yang relevan dengan pesan mereka
- Boleh santai tapi tetap supportive

Pesan user: %s

Respon dalam bahasa Indonesia yang natural dan casual (max 100 kata).`, 
		zodiacSign, traits, userMessage)
}

// buildInsightPrompt builds prompt for insight generation
func (c *GeminiClient) buildInsightPrompt(chatHistory string) string {
	return fmt.Sprintf(`Analyze this conversation and extract a profound life lesson or insight.
Create a short, inspirational message (max 200 words) that could help others facing similar situations.

Conversation:
%s

Generate a wisdom-filled insight that:
1. Identifies the core emotional theme
2. Offers a universal life lesson
3. Provides hope and encouragement
4. Is relatable to others

Format: A single paragraph of wisdom in Bahasa Indonesia. Make it profound and shareable.`, 
		chatHistory)
}

// getFallbackChatResponse returns fallback response if AI fails
func (c *GeminiClient) getFallbackChatResponse(zodiacSign string) string {
	fallbacks := map[string]string{
		"Aries":       "Halo! Maaf nih, lagi ada gangguan sebentar. Tapi aku di sini kok, siap dengerin kamu.",
		"Taurus":      "Hai! Ada yang bisa aku bantu? Cerita aja, aku dengerin.",
		"Gemini":      "Halo! Gimana kabarnya? Ada yang mau diobrolin?",
		"Cancer":      "Hai! Aku di sini kalau kamu mau cerita atau butuh temen ngobrol.",
		"Leo":         "Halo! Ada yang bisa aku bantu hari ini?",
		"Virgo":       "Hai! Cerita aja kalau ada yang mau dibahas, aku siap dengerin.",
		"Libra":       "Halo! Gimana hari ini? Ada yang mau diceritain?",
		"Scorpio":     "Hai! Aku di sini kalau kamu butuh temen ngobrol.",
		"Sagittarius": "Halo! Ada yang mau dibahas? Cerita aja santai.",
		"Capricorn":   "Hai! Gimana kabarnya? Aku siap dengerin kalau ada yang mau diceritain.",
		"Aquarius":    "Halo! Ada yang bisa aku bantu? Ngobrol aja santai.",
		"Pisces":      "Hai! Aku di sini kalau kamu butuh temen cerita.",
	}
	
	if response, ok := fallbacks[zodiacSign]; ok {
		return response
	}
	
	return "Halo! Ada yang bisa aku bantu? Cerita aja santai."
}

// getFallbackInsight returns fallback insight if AI fails
func (c *GeminiClient) getFallbackInsight() string {
	return "Setiap percakapan adalah cerminan dari perjalanan hidup kita. Dalam berbagi cerita dan perasaan, kita menemukan kekuatan untuk terus maju. Ingatlah bahwa setiap tantangan adalah kesempatan untuk tumbuh, dan setiap emosi yang kita rasakan adalah bagian dari kemanusiaan kita. Teruslah berbicara, teruslah berbagi, dan teruslah percaya bahwa hari esok membawa harapan baru."
}

// getZodiacTraits returns personality traits for zodiac signs
func getZodiacTraits(sign string) string {
	traits := map[string]string{
		"Aries":       "passionate, confident, determined, and courageous",
		"Taurus":      "reliable, patient, devoted, and practical",
		"Gemini":      "adaptable, outgoing, intelligent, and curious",
		"Cancer":      "intuitive, emotional, protective, and nurturing",
		"Leo":         "creative, passionate, generous, and warm-hearted",
		"Virgo":       "loyal, analytical, hardworking, and practical",
		"Libra":       "diplomatic, gracious, fair-minded, and social",
		"Scorpio":     "resourceful, brave, passionate, and determined",
		"Sagittarius": "generous, idealistic, great sense of humor, and adventurous",
		"Capricorn":   "responsible, disciplined, self-controlled, and ambitious",
		"Aquarius":    "progressive, original, independent, and humanitarian",
		"Pisces":      "compassionate, artistic, intuitive, and gentle",
	}
	
	if trait, ok := traits[sign]; ok {
		return trait
	}
	
	return "empathetic and understanding"
}

// Close closes the Gemini client
func (c *GeminiClient) Close() error {
	// Gemini client doesn't have explicit close method
	return nil
}
