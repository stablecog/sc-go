package utils

import (
	"os"
	"sync"
)

type TranslatorSafetyChecker struct {
	TranslatorCogUrl string
	openaiUrl        string
	openaiKey        string
	mu               sync.Mutex
}

func NewTranslatorSafetyChecker(openaiKey string) *TranslatorSafetyChecker {
	return &TranslatorSafetyChecker{
		TranslatorCogUrl: os.Getenv("TRANSLATOR_COG_URL"),
	}
}

func (t *TranslatorSafetyChecker) TranslatePrompt(prompt string) (string, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	return "", nil
}
