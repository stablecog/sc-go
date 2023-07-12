package responses

type DiscordRedisStreamMessage struct {
	DiscordId string `json:"user_id"`
	Username  string `json:"username"`
}
