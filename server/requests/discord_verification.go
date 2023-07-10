package requests

type DiscordVerificationRequest struct {
	Token  string `json:"token"`
	UserID string `json:"user_id"`
}
