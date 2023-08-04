package shared

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/pemistahl/lingua-go"
	"github.com/stablecog/sc-go/utils"
)

// Config
const targetLangFlores = "eng_Latn"
const targetLang = lingua.English
const targetLangMaxScore float64 = 0.88
const detectedConfidenceScoreMin = 0.1
const maxTextLength = 150

type LanguageDetector struct {
	LanguageDetector lingua.LanguageDetector
}

func NewLanguageDetector() *LanguageDetector {
	ld := lingua.NewLanguageDetectorBuilder().FromAllLanguages().Build()
	// Prime
	ld.ComputeLanguageConfidenceValues("hello world")

	// Return
	return &LanguageDetector{
		LanguageDetector: ld,
	}
}

func (ld *LanguageDetector) GetFloresCodes(inputs []string) (outputs []string) {
	s := time.Now()
	outputs = make([]string, len(inputs))
	for i, input := range inputs {
		outputs[i] = ld.GetFloresCode(input)
	}
	log.Printf("-- GetFloresCodes took: %v --", fmt.Sprint(time.Since(s).Milliseconds(), "ms"))
	return outputs
}

func (ld *LanguageDetector) GetFloresCode(text string) string {
	if text == "" {
		return targetLangFlores
	}

	if len(text) > maxTextLength {
		text = text[:maxTextLength]
	}

	confidenceValues := ld.LanguageDetector.ComputeLanguageConfidenceValues(text)
	if len(confidenceValues) < 1 {
		return targetLangFlores
	}

	var detectedLang *lingua.Language
	var detectedLangScore *float64
	var targetLangScore *float64
	for i, curr := range confidenceValues {
		if i == 0 {
			detectedLang = utils.ToPtr(curr.Language())
			detectedLangScore = utils.ToPtr(curr.Value())
		}
		if strings.ToUpper(curr.Language().String()) == "ENGLISH" {
			targetLangScore = utils.ToPtr(curr.Value())
		}
	}

	langToFlores, _ := LANG_TO_FLORES[strings.ToUpper(detectedLang.String())]

	if detectedLang != nil && *detectedLang != lingua.English &&
		*detectedLang != targetLang && *detectedLangScore > detectedConfidenceScoreMin &&
		(targetLangScore == nil || *targetLangScore < targetLangMaxScore) &&
		langToFlores != "" {
		return langToFlores
	}

	return targetLangFlores
}

var LANG_TO_FLORES = map[string]string{
	"AFRIKAANS":   "afr_Latn",
	"ALBANIAN":    "als_Latn",
	"ARABIC":      "arb_Arab",
	"ARMENIAN":    "hye_Armn",
	"AZERBAIJANI": "azj_Latn",
	"BASQUE":      "eus_Latn",
	"BELARUSIAN":  "bel_Cyrl",
	"BENGALI":     "ben_Beng",
	"BOKMAL":      "nob_Latn",
	"BOSNIAN":     "bos_Latn",
	"CATALAN":     "cat_Latn",
	"CHINESE":     "zho_Hans",
	"CROATIAN":    "hrv_Latn",
	"CZECH":       "ces_Latn",
	"DANISH":      "dan_Latn",
	"DUTCH":       "nld_Latn",
	"ENGLISH":     "eng_Latn",
	"ESPERANTO":   "epo_Latn",
	"ESTONIAN":    "est_Latn",
	"FINNISH":     "fin_Latn",
	"FRENCH":      "fra_Latn",
	"GANDA":       "lug_Latn",
	"GEORGIAN":    "kat_Geor",
	"GERMAN":      "deu_Latn",
	"GREEK":       "ell_Grek",
	"GUJARATI":    "guj_Gujr",
	"HEBREW":      "heb_Hebr",
	"HINDI":       "hin_Deva",
	"HUNGARIAN":   "hun_Latn",
	"ICELANDIC":   "isl_Latn",
	"INDONESIAN":  "ind_Latn",
	"IRISH":       "gle_Latn",
	"ITALIAN":     "ita_Latn",
	"JAPANESE":    "jpn_Jpan",
	"KAZAKH":      "kaz_Cyrl",
	"KOREAN":      "kor_Hang",
	"LATVIAN":     "lvs_Latn",
	"LITHUANIAN":  "lit_Latn",
	"MACEDONIAN":  "mkd_Cyrl",
	"MALAY":       "zsm_Latn",
	"MAORI":       "mri_Latn",
	"MARATHI":     "mar_Deva",
	"MONGOLIAN":   "khk_Cyrl",
	"NYNORSK":     "nno_Latn",
	"PERSIAN":     "pes_Arab",
	"POLISH":      "pol_Latn",
	"PORTUGUESE":  "por_Latn",
	"PUNJABI":     "pan_Guru",
	"ROMANIAN":    "ron_Latn",
	"RUSSIAN":     "rus_Cyrl",
	"SERBIAN":     "srp_Cyrl",
	"SHONA":       "sna_Latn",
	"SLOVAK":      "slk_Latn",
	"SLOVENE":     "slv_Latn",
	"SOMALI":      "som_Latn",
	"SOTHO":       "nso_Latn",
	"SPANISH":     "spa_Latn",
	"SWAHILI":     "swh_Latn",
	"SWEDISH":     "swe_Latn",
	"TAGALOG":     "tgl_Latn",
	"TAMIL":       "tam_Taml",
	"TELUGU":      "tel_Telu",
	"THAI":        "tha_Thai",
	"TSONGA":      "tso_Latn",
	"TURKISH":     "tur_Latn",
	"UKRAINIAN":   "ukr_Cyrl",
	"URDU":        "urd_Arab",
	"VIETNAMESE":  "vie_Latn",
	"XHOSA":       "xho_Latn",
	"YORUBA":      "yor_Latn",
	"ZULU":        "zul_Latn",
}
