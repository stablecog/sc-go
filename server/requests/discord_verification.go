package requests

type DiscordVerificationRequest struct {
	DiscordToken string `json:"discord_token"`
	DiscordID    string `json:"discord_id"`
}
