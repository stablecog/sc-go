package enttypes

type SourceType string

const (
	SourceTypeWebUI    SourceType = "web-ui"
	SourceTypeAPI      SourceType = "api"
	SourceTypeDiscord  SourceType = "discord"
	SourceTypeInternal SourceType = "internal"
)

// Values provides list valid values for Enum.
func (SourceType) Values() (kinds []string) {
	for _, s := range []SourceType{SourceTypeWebUI, SourceTypeAPI, SourceTypeDiscord, SourceTypeInternal} {
		kinds = append(kinds, string(s))
	}
	return
}
