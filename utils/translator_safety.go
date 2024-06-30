package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/exp/slices"

	openai "github.com/sashabaranov/go-openai"
	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/shared"
)

const TARGET_FLORES_CODE = "eng_Latn"
const TARGET_LANG_SCORE_MAX = 0.88
const DETECTED_CONFIDENCE_SCORE_MIN = 0.1

type TranslatorSafetyChecker struct {
	Ctx             context.Context
	TargetFloresUrl string
	OpenaiClient    *openai.Client
	// TODO - mock better for testing, this just disables
	Disable   bool
	activeUrl int
	urls      []string
	r         http.RoundTripper
	secret    string
	client    *http.Client
	mu        sync.Mutex
	rwmu      sync.RWMutex
}

func (t *TranslatorSafetyChecker) RoundTrip(r *http.Request) (*http.Response, error) {
	r.Header.Add("Authorization", t.secret)
	r.Header.Add("Content-Type", "application/json")
	return t.r.RoundTrip(r)
}

func NewTranslatorSafetyChecker(ctx context.Context, openaiKey string, disable bool) *TranslatorSafetyChecker {
	checker := &TranslatorSafetyChecker{
		Ctx:             ctx,
		TargetFloresUrl: GetEnv().PrivateLinguaAPIUrl,
		OpenaiClient:    openai.NewClient(openaiKey),
		Disable:         disable,
		secret:          GetEnv().NllbAPISecret,
		r:               http.DefaultTransport,
	}
	checker.client = &http.Client{
		Timeout:   10 * time.Second,
		Transport: checker,
	}
	return checker
}

func (t *TranslatorSafetyChecker) UpdateURLsFromCache() {
	urls := shared.GetCache().GetNLLBUrls()
	// Compare existing slice
	if slices.Compare(urls, t.urls) != 0 {
		t.rwmu.Lock()
		defer t.rwmu.Unlock()
		t.urls = urls
	}
}

type TargetFloresCodeRequest struct {
	Inputs []string `json:"inputs"`
}

type TargetFloresCodeResponse struct {
	Outputs []string `json:"outputs"`
}

func (t *TranslatorSafetyChecker) GetTargetFloresCode(inputs []string) ([]string, error) {
	if t.Disable {
		return inputs, nil
	}
	// Make request to target flores API
	request := TargetFloresCodeRequest{
		Inputs: inputs,
	}
	reqBody, err := json.Marshal(request)
	if err != nil {
		log.Error("Error marshalling webhook body", "err", err)
		return nil, err
	}
	// Make HTTP post to target flores API
	res, postErr := http.Post(t.TargetFloresUrl, "application/json", bytes.NewBuffer(reqBody))
	if postErr != nil {
		log.Error("Error sending target flores request", "err", postErr)
		return nil, postErr
	}
	defer res.Body.Close()

	var targetFloresResponse TargetFloresCodeResponse
	decoder := json.NewDecoder(res.Body)
	decodeErr := decoder.Decode(&targetFloresResponse)
	if decodeErr != nil {
		log.Error("Error decoding target flores response", "err", decodeErr)
		return nil, decodeErr
	}

	return targetFloresResponse.Outputs, nil
}

// Translator cog types
type TranslatorCogInput struct {
	Text1                       string  `json:"text_1"`
	TextFlores1                 string  `json:"text_flores_1"`
	TargetFlores1               string  `json:"target_flores_1"`
	TargetScoreMax1             float64 `json:"target_score_max_1"`
	DetectedConfidenceScoreMin1 float64 `json:"detected_confidence_score_min_1"`
	Text2                       string  `json:"text_2"`
	TextFlores2                 string  `json:"text_flores_2"`
	TargetFlores2               string  `json:"target_flores_2"`
	TargetScoreMax2             float64 `json:"target_score_max_2"`
	DetectedConfidenceScoreMin2 float64 `json:"detected_confidence_score_min_2"`
}

type TranslatorCogRequest struct {
	Input TranslatorCogInput `json:"input"`
}

type TranslatorCogResponse struct {
	Output []string `json:"output"`
}

// Send translation to the translator cog
func (t *TranslatorSafetyChecker) TranslatePrompt(prompt string, negativePrompt string) (translatedPrompt string, translatedNegativePrompt string, err error) {
	t.UpdateURLsFromCache()

	// Prevent many calls to the translator cog
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.Disable {
		return prompt, negativePrompt, nil
	}

	var inputs []string
	if negativePrompt == "" {
		inputs = []string{prompt}
	} else {
		inputs = []string{prompt, negativePrompt}
	}

	targetCodes, err := t.GetTargetFloresCode(inputs)
	if err != nil {
		log.Error("Error getting target flores code", "err", err)
	} else {
		// Skip translator if all target codes are TARGET_FLORES_CODE
		allTargetFlores := true
		for _, floresCode := range targetCodes {
			if floresCode != TARGET_FLORES_CODE {
				allTargetFlores = false
				break
			}
		}
		if allTargetFlores {
			return prompt, negativePrompt, nil
		}
	}

	var textFlores1 string
	var textFlores2 string
	for i, floresCode := range targetCodes {
		if i == 0 {
			textFlores1 = floresCode
		} else {
			textFlores2 = floresCode
		}
	}

	// Build request for translator cog
	translatorRequest := TranslatorCogRequest{
		Input: TranslatorCogInput{
			Text1:                       prompt,
			TextFlores1:                 textFlores1,
			TargetFlores1:               TARGET_FLORES_CODE,
			TargetScoreMax1:             TARGET_LANG_SCORE_MAX,
			DetectedConfidenceScoreMin1: DETECTED_CONFIDENCE_SCORE_MIN,
			Text2:                       negativePrompt,
			TextFlores2:                 textFlores2,
			TargetFlores2:               TARGET_FLORES_CODE,
			TargetScoreMax2:             TARGET_LANG_SCORE_MAX,
			DetectedConfidenceScoreMin2: DETECTED_CONFIDENCE_SCORE_MIN,
		},
	}

	// Get URL to request to
	t.rwmu.RLock()
	url := t.urls[t.activeUrl]
	t.rwmu.RUnlock()

	// Marshal
	reqBody, err := json.Marshal(translatorRequest)
	if err != nil {
		log.Error("Error marshalling webhook body", "err", err)
		return "", "", err
	}
	// Make HTTP post to target flores API
	res, postErr := t.client.Post(fmt.Sprintf("%s/predictions", url), "application/json", bytes.NewBuffer(reqBody))
	if postErr != nil {
		log.Error("Error sending translator cog request", "err", postErr)
		return "", "", postErr
	}
	defer res.Body.Close()

	var translatorResponse TranslatorCogResponse
	decoder := json.NewDecoder(res.Body)
	decodeErr := decoder.Decode(&translatorResponse)
	if decodeErr != nil {
		log.Error("Error decoding translator cog response", "err", decodeErr)
		return "", "", decodeErr
	}

	for i, output := range translatorResponse.Output {
		if i == 0 {
			translatedPrompt = output
		} else {
			translatedNegativePrompt = output
		}
	}

	// Update URL
	t.rwmu.Lock()
	t.activeUrl = (t.activeUrl + 1) % len(t.urls)
	t.rwmu.Unlock()

	return translatedPrompt, translatedNegativePrompt, nil
}

// Safety check
func (t *TranslatorSafetyChecker) IsPromptNSFW(input string) (isNsfw bool, nsfwReason string, score float32, err error) {
	if t.Disable {
		return false, "", 0, nil
	}
	// API check
	res, err := t.OpenaiClient.Moderations(t.Ctx, openai.ModerationRequest{
		Input: input,
	})
	if err != nil {
		log.Error("Error calling openai safety check", "err", err)
		return
	}
	if len(res.Results) < 0 {
		log.Error("Error calling openai safety check", "err", "no results")
		err = errors.New("no results")
		return
	}

	isMinors := res.Results[0].Categories.SexualMinors || res.Results[0].CategoryScores.SexualMinors > 0.25
	isSexual := res.Results[0].Categories.Sexual || res.Results[0].CategoryScores.Sexual > 0.3
	isNsfw = isMinors || isSexual
	if isNsfw {
		// Populate reason
		if isMinors {
			nsfwReason = "sexual_minors"
			score = res.Results[0].CategoryScores.SexualMinors
		} else {
			nsfwReason = "sexual"
			score = res.Results[0].CategoryScores.Sexual
		}
	}
	return
}

// Write a function to check if a given array includes the given value
func includes(value string, array []string) bool {
	for _, item := range array {
		if item == value {
			return true
		}
	}
	return false
}

func removeChars(s string, charsToRemove string) string {
	// Convert the charsToRemove string to a map for O(1) lookup
	charMap := make(map[rune]bool)
	for _, c := range charsToRemove {
		charMap[c] = true
	}

	// Use a strings.Builder for efficient string manipulation
	var builder strings.Builder
	for _, c := range s {
		if !charMap[c] { // If the character is not in the map, append it to the result
			builder.WriteRune(c)
		}
	}

	return builder.String()
}
