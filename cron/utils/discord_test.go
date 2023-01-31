package utils

import (
	"testing"
)

// var MockDiscordHealthTracker *DiscordHealthTracker

// func TestMain(m *testing.M) {
// 	os.Exit(testMainWrapper(m))
// }

// func testMainWrapper(m *testing.M) int {
// 	// Setup
// 	os.Setenv("MOCK_REDIS", "true")
// 	defer os.Unsetenv("MOCK_REDIS")
// 	os.Setenv("DISCORD_WEBHOOK_URL", "http://localhost:123456")
// 	defer os.Unsetenv("DISCORD_WEBHOOK_URL")

// 	redis, err := database.NewRedis(context.Background())
// 	if err != nil {
// 		panic(err)
// 	}
// 	MockDiscordHealthTracker = NewDiscordHealthTracker(context.Background(), redis)

// 	return m.Run()
// }

func TestSendDiscordNotificationIfNeeded(t *testing.T) {
	// // Mock logger
	// orgKlogInfof := klogInfof
	// defer func() { klogInfof = orgKlogInfof }()

	// // Write log output to string
	// logs := []string{}
	// klogInfof = func(format string, args ...interface{}) {
	// 	logs = append(logs, fmt.Sprintf(format, args...))
	// }

	// // Mock generations
	// generations := []*ent.Generation{}
	// failedStatus := generation.StatusFailed
	// startedStatus := generation.StatusStarted
	// successStatus := generation.StatusSucceeded
	// nsfw := "NSFW"
	// generations = append(generations, &ent.Generation{
	// 	ID:            uuid.New(),
	// 	FailureReason: &nsfw,
	// 	Status:        &failedStatus,
	// })
	// generations = append(generations, &ent.Generation{
	// 	ID:            uuid.New(),
	// 	FailureReason: nil,
	// 	Status:        &failedStatus,
	// })
	// generations = append(generations, &ent.Generation{
	// 	ID:            uuid.New(),
	// 	FailureReason: nil,
	// 	Status:        &startedStatus,
	// })
	// generations = append(generations, &ent.Generation{
	// 	ID:            uuid.New(),
	// 	FailureReason: nil,
	// 	Status:        &successStatus,
	// })

	// // ! Test invalid address
	// err := MockDiscordHealthTracker.SendDiscordNotificationIfNeeded("test", "test", generations, time.Now(), time.Now())
	// assert.NotNil(t, err)
	// assert.Equal(t, "Post \"http://localhost:123456\": dial tcp: address 123456: invalid port", err.Error())

	// // ! Test notification not needed
	// err = MockDiscordHealthTracker.SendDiscordNotificationIfNeeded("test", "unknown", generations, time.Now(), time.Now())
	// assert.Nil(t, err)
	// assert.Equal(t, "Skipping Discord notification, not needed", logs[0])

	// MockDiscordHealthTracker.lastNotificationTime = time.Now()
	// err = MockDiscordHealthTracker.redis.Set(MockDiscordHealthTracker.ctx, lastHealthyKey, MockDiscordHealthTracker.lastNotificationTime.Format(time.RFC3339), rTTL).Err()
	// assert.Nil(t, err)
	// err = MockDiscordHealthTracker.redis.Set(MockDiscordHealthTracker.ctx, lastUnhealthyKey, MockDiscordHealthTracker.lastNotificationTime.Format(time.RFC3339), rTTL).Err()
	// assert.Nil(t, err)

	// err = MockDiscordHealthTracker.SendDiscordNotificationIfNeeded("unhealthy", "unhealthy", generations, time.Now(), time.Now())
	// assert.Nil(t, err)
	// assert.Equal(t, "Skipping Discord notification, not needed", logs[1])

	// // Cleanup redis keys
	// err = MockDiscordHealthTracker.redis.Del(MockDiscordHealthTracker.ctx, lastHealthyKey).Err()
	// assert.Nil(t, err)
	// err = MockDiscordHealthTracker.redis.Del(MockDiscordHealthTracker.ctx, lastUnhealthyKey).Err()
	// assert.Nil(t, err)

	// // ! Test notification needed
	// // Setup httpmock
	// httpmock.Activate()
	// defer httpmock.DeactivateAndReset()

	// httpmock.RegisterResponder("POST", "http://localhost:123456",
	// 	func(req *http.Request) (*http.Response, error) {
	// 		var request models.DiscordWebhookBody
	// 		err := json.NewDecoder(req.Body).Decode(&request)
	// 		assert.Nil(t, err)
	// 		assert.Equal(t, 11437547, request.Embeds[0].Color)
	// 		assert.Equal(t, "```🟢👌🟢```", request.Embeds[0].Fields[0].Value)
	// 		assert.Equal(t, "```🌶️🔴🟡🟢```", request.Embeds[0].Fields[1].Value)
	// 		assert.Equal(t, "```Just now```", request.Embeds[0].Fields[2].Value)

	// 		resp, err := httpmock.NewJsonResponse(200, map[string]interface{}{
	// 			"status": "ok",
	// 		})
	// 		return resp, err
	// 	},
	// )

	// err = MockDiscordHealthTracker.SendDiscordNotificationIfNeeded("healthy", "unhealthy", generations, time.Now(), time.Now())
	// assert.Nil(t, err)
}
