package translator

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	openai "github.com/sashabaranov/go-openai"
	"github.com/stablecog/sc-go/database"
	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/utils"
)

const TARGET_FLORES_CODE = "eng_Latn"
const TARGET_LANG_SCORE_MAX = 0.88
const DETECTED_CONFIDENCE_SCORE_MIN = 0.1
const OPENAI_TRANSLATOR_SYSTEM_MESSAGE = "You are a helpful translator. You always translate the entire message to English. If the message is already in English, just respond with the message untouched. Only answer with the translation."
const OPENAI_TRANSLATOR_MAX_TOKENS = 500

type TranslatorSafetyChecker struct {
	Ctx             context.Context
	TargetFloresUrl string
	OpenaiClient    *openai.Client
	Redis           *database.RedisWrapper
	// TODO - mock better for testing, this just disables
	Disable    bool
	activeUrl  int
	urls       []string
	HTTPClient *http.Client
}

func NewTranslatorSafetyChecker(ctx context.Context, openaiKey string, disable bool, redis *database.RedisWrapper) *TranslatorSafetyChecker {
	return &TranslatorSafetyChecker{
		Ctx:             ctx,
		TargetFloresUrl: utils.GetEnv().PrivateLinguaAPIUrl,
		OpenaiClient:    openai.NewClient(openaiKey),
		Disable:         disable,
		Redis:           redis,
		HTTPClient:      &http.Client{}, // Initialize a reusable HTTP client
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
	req, err := http.NewRequest("POST", t.TargetFloresUrl, bytes.NewBuffer(reqBody))
	if err != nil {
		log.Error("Error creating request", "err", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := t.HTTPClient.Do(req)
	if err != nil {
		log.Error("Error sending request", "err", err)
		return nil, err
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

	hasPromptCache := promptCacheErr == nil
	hasNegativePromptCache := negativePromptCacheKey != "" && negativePromptCacheErr == nil

	if hasPromptCache && hasNegativePromptCache {
		log.Infof("<> 🟢 Cache hit for prompt and negative prompt, returning: %s • %s /// %s • %s", prompt, promptCacheRes, negativePrompt, negativePromptCacheRes)
		return promptCacheRes, negativePromptCacheRes, nil
	}
	if hasPromptCache && (negativePrompt == "") {
		log.Infof("<> 🟢 Cache hit for prompt, no negative prompt, returning: %s • %s", prompt, promptCacheRes)
		return promptCacheRes, negativePrompt, nil
	}

	if hasPromptCache {
		log.Infof("<> 🟠 Partial cache hit for prompt: %s • %s", prompt, promptCacheRes)
		translatedPrompt = promptCacheRes
	}
	if hasNegativePromptCache {
		log.Infof("<> 🟠 Partial cache hit for negative prompt: %s • %s", negativePrompt, negativePromptCacheRes)
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
			var promptToReturn string
			var negativePromptToReturn string
			if hasPromptCache {
				promptToReturn = promptCacheRes
			} else {
				promptToReturn = prompt
			}
			if hasNegativePromptCache {
				negativePromptToReturn = negativePromptCacheRes
			} else {
				negativePromptToReturn = negativePrompt
			}
			return promptToReturn, negativePromptToReturn, nil
		}
	}

	if !hasPromptCache {
		promptRes, promptErr := TranslateViaOpenAI(prompt, t.OpenaiClient, t.Ctx)
		if promptErr != nil {
			log.Error("Error calling OpenAI translator for prompt", "err", promptErr)
			return prompt, negativePrompt, promptErr
		} else {
			translatedPrompt = promptRes
			log.Infof("<> ✅ Translated prompt: %s • %s", prompt, translatedPrompt)
			// Update cache
			err = t.Redis.CacheTranslation(t.Ctx, promptCacheKey, translatedPrompt)
			if err != nil {
				log.Error("Error caching translated prompt", "err", err)
			}
		}
	}

	if negativePrompt != "" && negativePromptCacheErr != nil {
		negativePromptRes, negativePromptErr := TranslateViaOpenAI(negativePrompt, t.OpenaiClient, t.Ctx)
		if negativePromptErr != nil {
			log.Error("Error calling OpenAI translator for negative prompt", "err", negativePromptErr)
			return prompt, negativePrompt, negativePromptErr
		} else {
			translatedNegativePrompt = negativePromptRes
			log.Infof("<> ✅ Translated negative prompt: %s • %s", negativePrompt, translatedNegativePrompt)
			// Update cache
			err = t.Redis.CacheTranslation(t.Ctx, negativePromptCacheKey, translatedNegativePrompt)
			if err != nil {
				log.Error("Error caching translated negative prompt", "err", err)
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
		Model: "omni-moderation-latest",
		Input: input,
	})

	if err != nil {
		log.Error("Error calling openai safety check", "err", err)
		return true, "", 0, err
	}

	if len(res.Results) < 1 {
		log.Error("Error calling openai safety check", "err", "no results")
		err = errors.New("no results")
		return true, "", 0, err
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

func TranslateViaOpenAI(prompt string, client *openai.Client, ctx context.Context) (string, error) {
	res, err := client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model:     openai.GPT4oMini,
			MaxTokens: OPENAI_TRANSLATOR_MAX_TOKENS,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: OPENAI_TRANSLATOR_SYSTEM_MESSAGE,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
		},
	)
	if err != nil {
		return "", err
	}
	return res.Choices[0].Message.Content, err
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
