package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"

	openai "github.com/sashabaranov/go-openai"
	"github.com/stablecog/sc-go/log"
)

const TARGET_FLORES_CODE = "eng_Latn"
const TARGET_LANG_SCORE_MAX = 0.88
const DETECTED_CONFIDENCE_SCORE_MIN = 0.1

type TranslatorSafetyChecker struct {
	Ctx              context.Context
	TargetFloresUrl  string
	TranslatorCogUrl string
	OpenaiClient     *openai.Client
	// TODO - mock better for testing, this just disables
	Disable bool
	mu      sync.Mutex
}

type TBannedPair struct {
	reason string
	words  []string
}

var bannedPairs []TBannedPair = []TBannedPair{
	{
		words:  []string{"teen", "twink"},
		reason: "sexual_minors",
	},
	{
		words:  []string{"young", "twink"},
		reason: "sexual_minors",
	},
	{
		words:  []string{"twink", "mirror"},
		reason: "sexual_minors",
	},
	{
		words:  []string{"teenage", "model"},
		reason: "sexual_minors",
	},
	{
		words:  []string{"teen", "model"},
		reason: "sexual_minors",
	},
	{
		words:  []string{"teen", "model", "fitness"},
		reason: "sexual_minors",
	},
}

func NewTranslatorSafetyChecker(ctx context.Context, openaiKey string, disable bool) *TranslatorSafetyChecker {
	return &TranslatorSafetyChecker{
		Ctx:              ctx,
		TargetFloresUrl:  os.Getenv("PRIVATE_LINGUA_API_URL"),
		TranslatorCogUrl: os.Getenv("TRANSLATOR_COG_URL"),
		OpenaiClient:     openai.NewClient(openaiKey),
		Disable:          disable,
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

	// Marshal
	reqBody, err := json.Marshal(translatorRequest)
	if err != nil {
		log.Error("Error marshalling webhook body", "err", err)
		return "", "", err
	}
	// Make HTTP post to target flores API
	res, postErr := http.Post(fmt.Sprintf("%s/predictions", t.TranslatorCogUrl), "application/json", bytes.NewBuffer(reqBody))
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

	return translatedPrompt, translatedNegativePrompt, nil
}

// Safety check
func (t *TranslatorSafetyChecker) IsPromptNSFW(input string) (isNsfw bool, nsfwReason string, err error) {
	if t.Disable {
		return false, "", nil
	}
	// Word list check
	for _, pair := range bannedPairs {
		containsAll := true
		for _, word := range pair.words {
			if !strings.Contains(input, word) {
				containsAll = false
				break
			}
		}
		if containsAll {
			return true, pair.reason, nil
		}
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

	isNsfw = res.Results[0].Categories.Sexual || res.Results[0].Categories.SexualMinors || res.Results[0].CategoryScores.Sexual > 0.25 || res.Results[0].CategoryScores.SexualMinors > 0.25
	if isNsfw {
		// Populate reason
		if res.Results[0].Categories.Sexual {
			if nsfwReason != "" {
				nsfwReason += ","
			}
			nsfwReason = "sexual"
		} else if res.Results[0].Categories.SexualMinors {
			if nsfwReason != "" {
				nsfwReason += ","
			}
			nsfwReason = "sexual_minors"
		}
	}
	return
}
