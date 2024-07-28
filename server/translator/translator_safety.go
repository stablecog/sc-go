package translator

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/exp/slices"

	openai "github.com/sashabaranov/go-openai"
	"github.com/stablecog/sc-go/database"
	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/shared"
	"github.com/stablecog/sc-go/utils"
)

const TARGET_FLORES_CODE = "eng_Latn"
const TARGET_LANG_SCORE_MAX = 0.88
const DETECTED_CONFIDENCE_SCORE_MIN = 0.1
const TRANSLATOR_SYSTEM_MESSAGE = "You are a helpful translator. You always translate the entire message to English. If the message is already in English, just respond with the message untouched. Only answer with the translation."

type TranslatorSafetyChecker struct {
	Ctx             context.Context
	TargetFloresUrl string
	OpenaiClient    *openai.Client
	Redis           *database.RedisWrapper
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

func NewTranslatorSafetyChecker(ctx context.Context, openaiKey string, disable bool, redis *database.RedisWrapper) *TranslatorSafetyChecker {
	checker := &TranslatorSafetyChecker{
		Ctx:             ctx,
		TargetFloresUrl: utils.GetEnv().PrivateLinguaAPIUrl,
		OpenaiClient:    openai.NewClient(openaiKey),
		Disable:         disable,
		secret:          utils.GetEnv().NllbAPISecret,
		r:               http.DefaultTransport,
		Redis:           redis,
	}
	checker.client = &http.Client{
		Timeout:   10 * time.Second,
		Transport: checker,
	}
	return checker
}

func (t *TranslatorSafetyChecker) UpdateURLsFromCache() {
	urls := shared.GetCache().GetNLLBUrls()
	if len(urls) == 0 {
		return
	}
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
	if t.Disable {
		return prompt, negativePrompt, nil
	}
	// See if we can get the translation from cache
	var promptCacheKey string
	var negativePromptCacheKey string

	promptCacheKey = utils.Sha256(prompt)
	if negativePrompt != "" {
		negativePromptCacheKey = utils.Sha256(negativePrompt)
	}

	var promptCacheRes string
	var negativePromptCacheRes string
	var promptCacheErr error
	var negativePromptCacheErr error

	promptCacheRes, promptCacheErr = t.Redis.GetTranslation(t.Ctx, promptCacheKey)
	if negativePromptCacheKey != "" {
		negativePromptCacheRes, negativePromptCacheErr = t.Redis.GetTranslation(t.Ctx, negativePromptCacheKey)
	}

	// Cache hit for prompt and negative prompt
	if promptCacheErr == nil && negativePromptCacheErr == nil {
		log.Infof("🈳🟢 Cache hit for prompt and negative prompt, returning: %s • %s /// %s • %s", prompt, promptCacheRes, negativePrompt, negativePromptCacheRes)
		return promptCacheRes, negativePromptCacheRes, nil
	}
	// Cache hit for prompt, no negative prompt
	if promptCacheErr == nil && (negativePrompt == "") {
		log.Infof("🈳🟢 Cache hit for prompt, no negative prompt, returning: %s • %s", prompt, promptCacheRes)
		return promptCacheRes, negativePrompt, nil
	}

	if promptCacheErr == nil {
		log.Infof("🈳🟠 Partial cache hit for prompt: %s • %s", prompt, promptCacheRes)
		translatedPrompt = promptCacheRes
	}
	if negativePromptCacheErr == nil {
		log.Infof("🈳🟠 Partial cache hit for negative prompt: %s • %s", negativePrompt, negativePromptCacheRes)
		translatedNegativePrompt = negativePromptCacheRes
	}

	var inputs []string
	if promptCacheErr != nil {
		inputs = append(inputs, prompt)
	}
	if negativePrompt != "" && negativePromptCacheErr != nil {
		inputs = append(inputs, negativePrompt)
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
			if promptCacheErr == nil {
				return translatedPrompt, negativePrompt, nil
			}
			if negativePromptCacheErr == nil {
				return prompt, translatedNegativePrompt, nil
			}
			return prompt, negativePrompt, nil
		}
	}

	if promptCacheErr != nil {
		promptRes, promptErr := t.OpenaiClient.CreateChatCompletion(
			t.Ctx,
			openai.ChatCompletionRequest{
				Model: openai.GPT4oMini,
				Messages: []openai.ChatCompletionMessage{
					{
						Role:    openai.ChatMessageRoleSystem,
						Content: TRANSLATOR_SYSTEM_MESSAGE,
					},
					{
						Role:    openai.ChatMessageRoleUser,
						Content: prompt,
					},
				},
			},
		)
		if promptErr != nil {
			log.Error("Error calling OpenAI translator for prompt", "err", promptErr)
			return prompt, negativePrompt, promptErr
		} else {
			translatedPrompt = promptRes.Choices[0].Message.Content
			log.Infof("🈳✅ Translated prompt: %s • %s", prompt, translatedPrompt)
			// Update cache
			err = t.Redis.CacheTranslation(t.Ctx, promptCacheKey, translatedPrompt)
			if err != nil {
				log.Error("Error caching translated prompt", "err", err)
			}
		}
	}

	if negativePrompt != "" && negativePromptCacheErr != nil {
		negativePromptRes, negativePromptErr := t.OpenaiClient.CreateChatCompletion(
			t.Ctx,
			openai.ChatCompletionRequest{
				Model: openai.GPT4oMini,
				Messages: []openai.ChatCompletionMessage{
					{
						Role:    openai.ChatMessageRoleSystem,
						Content: TRANSLATOR_SYSTEM_MESSAGE,
					},
					{
						Role:    openai.ChatMessageRoleUser,
						Content: negativePrompt,
					},
				},
			},
		)
		if negativePromptErr != nil {
			log.Error("Error calling OpenAI translator for negative prompt", "err", negativePromptErr)
			return prompt, negativePrompt, negativePromptErr
		} else {
			translatedNegativePrompt = negativePromptRes.Choices[0].Message.Content
			log.Infof("🈳✅ Translated negative prompt: %s • %s", negativePrompt, translatedNegativePrompt)
			// Update cache
			if negativePrompt != "" {
				err = t.Redis.CacheTranslation(t.Ctx, negativePromptCacheKey, translatedNegativePrompt)
				if err != nil {
					log.Error("Error caching translated negative prompt", "err", err)
				}
			}
		}
	}

	if negativePrompt == "" {
		translatedNegativePrompt = negativePrompt
	}

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
