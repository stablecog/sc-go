package requests

type DiscordVerificationRequest struct {
	Token     string `json:"token"`
	DiscordID string `json:"discord_id"`
}
