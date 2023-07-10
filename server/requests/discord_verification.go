package requests

type DiscordVerificationRequest struct {
	Token  string `json:"platform_token"`
	UserID string `json:"platform_user_id"`
}
