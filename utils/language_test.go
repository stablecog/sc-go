package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetFloresCode(t *testing.T) {
	// Empty input is english
	assert.Equal(t, targetLangFlores, GetFloresCode(""))
	// Low confidence value is english
	assert.Equal(t, targetLangFlores, GetFloresCode("$@!"))
	// English
	assert.Equal(t, targetLangFlores, GetFloresCode("A portrait of a cat by van gogh"))
	// French
	assert.Equal(t, "fra_Latn", GetFloresCode("Un portrait de chat par van gogh"))
	// German
	assert.Equal(t, "deu_Latn", GetFloresCode("Ein Porträt eines Katers von van gogh"))
	// Chinese
	assert.Equal(t, "zho_Hans", GetFloresCode("一幅猫的肖像由梵高"))
}
