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
	
	// Create Gemini client (API key from environment)
	client, err := genai.NewClient(ctx, nil)
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
		// Fallback response if AI fails
		return c.getFallbackChatResponse(zodiacSign), nil
	}
	
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
	
	return fmt.Sprintf(`You are a %s AI companion with these personality traits: %s.

Respond to the user's message with empathy, wisdom, and understanding characteristic of %s.
Be supportive, insightful, and help them reflect on their feelings.

User message: %s

Respond in a warm, compassionate tone (max 150 words). Speak in Bahasa Indonesia.`, 
		zodiacSign, traits, zodiacSign, userMessage)
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
		"Aries":       "Saya mendengarkan Anda. Keberanian Anda untuk berbagi ini menunjukkan kekuatan sejati. Teruslah maju dengan percaya diri.",
		"Taurus":      "Terima kasih telah berbagi. Kesabaran dan keteguhan Anda akan membawa Anda melewati ini. Percayalah pada prosesnya.",
		"Gemini":      "Saya memahami perspektif Anda. Kemampuan adaptasi Anda adalah kekuatan. Teruslah terbuka terhadap kemungkinan baru.",
		"Cancer":      "Perasaan Anda valid dan penting. Intuisi Anda membimbing Anda dengan baik. Percayalah pada diri sendiri.",
		"Leo":         "Saya menghargai keterbukaan Anda. Kekuatan dan kreativitas Anda akan membantu Anda menemukan jalan. Tetaplah bersinar.",
		"Virgo":       "Terima kasih atas kepercayaan Anda. Analisis dan dedikasi Anda akan membawa kejelasan. Teruslah berusaha.",
		"Libra":       "Saya mendengarkan dengan penuh perhatian. Keseimbangan dan kebijaksanaan Anda akan membantu menemukan harmoni. Tetaplah adil pada diri sendiri.",
		"Scorpio":     "Keberanian Anda untuk menghadapi ini menginspirasi. Kekuatan batin Anda luar biasa. Percayalah pada transformasi.",
		"Sagittarius": "Optimisme Anda adalah hadiah. Teruslah mencari makna dan pertumbuhan. Petualangan ini akan mengajarkan banyak hal.",
		"Capricorn":   "Disiplin dan tanggung jawab Anda patut dihormati. Teruslah bergerak maju dengan tujuan yang jelas. Anda akan berhasil.",
		"Aquarius":    "Perspektif unik Anda berharga. Teruslah berinovasi dan berpikir bebas. Perubahan dimulai dari dalam.",
		"Pisces":      "Empati dan kreativitas Anda adalah kekuatan. Percayalah pada intuisi artistik Anda. Anda tidak sendirian.",
	}
	
	if response, ok := fallbacks[zodiacSign]; ok {
		return response
	}
	
	return "Saya mendengarkan Anda. Terima kasih telah berbagi perasaan Anda. Anda berani dan kuat."
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
