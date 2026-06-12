package jobs

import (
	"context"
	"errors"
	"strings"
	"time"

	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/sethvargo/go-retry"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/database/ent/generation"
	"github.com/stablecog/sc-go/shared"
	"github.com/stablecog/sc-go/utils"
)

// ── How health is decided ────────────────────────────────────────────────────
//
// The status this job tracks is the health of the MAIN SC worker (the MQ
// path), not "is the system serving" — during a Runpod failover the system
// serves fine while the main worker may still be down.
//
// 1. Only a synthetic test generation can produce a verdict. Traffic signals
//    merely decide whether a test is needed: a high non-NSFW failure rate
//    among recent real generations, or silence (no successful generation
//    within noSuccessfulGenerationWindow), triggers a probe — real traffic
//    can fail for reasons that have nothing to do with the worker, so it is
//    never trusted as a verdict on its own.
// 2. Tester generations always force the main-worker MQ path (see
//    scworker.CreateGeneration), bypassing the RunpodActive routing flag.
//    While a failover this job initiated is active, every run probes: real
//    traffic routes through Runpod then and says nothing about the main
//    worker, so only successful probes can build the recovery streak.
// 3. Test failures are classified: 4xx responses mean the *test* is broken
//    (credits, auth, rate limit) — never the worker — and must not trigger a
//    failover. Timeouts / 5xx / transport errors count against the worker.
// 4. Failover has hysteresis: Runpod serverless is only enabled after
//    consecutiveUnhealthyChecksForFailover consecutive unhealthy runs, and is
//    automatically disabled again after consecutiveHealthyChecksForRecovery
//    consecutive healthy runs — but only if this job was the one to enable it.

// Considered failed if failed/(succeeded+failed) > maxGenerationFailWithoutNSFWRate
// among real generations completed within noSuccessfulGenerationWindow.
// NSFW failures are excluded entirely: they're user behavior, not worker health.
const maxGenerationFailWithoutNSFWRate = 0.5

// Below this many completed recent generations the failure rate is too noisy
// to judge by traffic alone — fall back to a synthetic test generation.
const minCompletedToJudgeByTraffic = 3

// Get this number of generations on each check, sorted by created_at DESC
const generationCountToCheck = 10
const successfulGenerationCountToCheck = 1

// If there is no successful generation within this window, the check is
// inconclusive from traffic alone and a synthetic test generation is run.
const noSuccessfulGenerationWindow = 3 * time.Minute

// Per-attempt HTTP timeout for the test generation request. A normal successful
// generation takes ~20s end to end, but the request rides the full production
// pipeline (queue, worker, S3, webhook) whose healthy tail can exceed 30s — a
// tight timeout here is exactly what produced false UNHEALTHY verdicts before.
const testGenerationHTTPTimeout = 90 * time.Second

// Total attempts = testGenerationRetries + 1. Retries are NOT independent
// trials: a client-abandoned attempt keeps running server-side for up to
// shared.REQUEST_COG_TIMEOUT, so piling on attempts mostly measures our own
// backlog. Two generous attempts beat five rushed ones.
const testGenerationRetries uint64 = 1
const testGenerationRetryBaseDelay = 5 * time.Second

// Wall-clock budget for the entire health check. Must accommodate the
// worst-case retry loop: 2 attempts × 90s + 5s backoff = 185s, so 200s leaves
// a small safety margin.
//
// SingletonMode on the cron schedule means a slow run that consumes most of
// this budget will simply skip the intervening 60s firings — exactly what we
// want. What this bound enforces is: a single run can't drift so far that its
// start-of-run snapshot becomes meaningless by the time the embed is posted.
const healthCheckBudget = 200 * time.Second

// Hysteresis: how many consecutive non-HEALTHY/HEALTHY verdicts are required
// before enabling/disabling Runpod serverless. A single bad run (one slow
// generation, one network blip) must never flip production to the fallback.
const consecutiveUnhealthyChecksForFailover = 2
const consecutiveHealthyChecksForRecovery = 5

const HEALTH_JOB_NAME = "HEALTH_JOB"

// CheckHealth cron job
func (j *JobRunner) CheckSCWorkerHealth(log Logger) error {
	// asOf is sampled once at the start of the run and is the reference time used
	// for every relative-time display + the embed footer. If we used time.Now()
	// at send time instead, a slow run (e.g. one stuck in test-gen retries) would
	// render its old snapshot against a fresh "now" and produce timestamps that
	// look impossible compared to neighbouring fast runs ("20m ago" sandwiched
	// between two "Just now" messages within a single minute).
	asOf := time.Now()
	ctx, cancel := context.WithTimeout(j.Ctx, healthCheckBudget)
	defer cancel()

	log.Infof("Checking health...")
	apiKey := utils.GetEnv().ScWorkerTesterApiKey

	generations, err := j.Repo.GetGenerations(generationCountToCheck)
	if err != nil || len(generations) == 0 {
		log.Errorf("Couldn't get generations %v", err)
		return err
	}

	successfulGenerations, err := j.Repo.GetSuccessfulGenerations(successfulGenerationCountToCheck)
	if err != nil {
		log.Errorf("Couldn't get successful generations %v", err)
		return err
	}

	lastGenerationTime := asOf.Add(-24 * time.Hour)
	lastSuccessfulGenerationTime := asOf.Add(-24 * time.Hour)

	if len(generations) > 0 {
		lastGenerationTime = generations[0].CreatedAt
	}
	if len(successfulGenerations) > 0 {
		lastSuccessfulGenerationTime = successfulGenerations[0].CreatedAt
	}

	// Runpod routing state is needed BEFORE the verdict: while Runpod serverless
	// is active, real traffic routes through Runpod, so traffic stats say
	// nothing about the main SC worker.
	isRunpodServerlessActive, runpodServerlessErr := j.Repo.IsRunpodServerlessActive()

	if runpodServerlessErr != nil {
		log.Errorf("🏃‍♂️‍➡️📦 🔴 Couldn't check if Runpod serverless is active: %v", runpodServerlessErr)
	}

	if isRunpodServerlessActive {
		log.Infof("🏃‍♂️‍➡️📦 🟢 Runpod serverless is active")
	}

	workerHealthStatus := shared.HEALTHY
	// The reason behind a non-HEALTHY verdict; surfaced in the Discord embed so
	// the alert itself says what actually went wrong.
	var healthErr error

	succeeded, failed := countRecentCompleted(generations, asOf, noSuccessfulGenerationWindow)
	completed := succeeded + failed

	// Decide whether a synthetic test generation is needed. Traffic signals are
	// triggers only — the probe is the verdict.
	runTestGeneration := false
	if isRunpodServerlessActive && j.runpodAutoEnabled {
		// Recovery assessment: real traffic is on Runpod, so only a probe forced
		// through the main worker can tell whether it recovered.
		runTestGeneration = true
		log.Infof("Runpod serverless failover is active, probing the main worker with a test generation...")
	} else if completed >= minCompletedToJudgeByTraffic && float64(failed)/float64(completed) > maxGenerationFailWithoutNSFWRate {
		// Real traffic is failing suspiciously often, but generations can fail
		// for user-side reasons — confirm with a probe before any verdict.
		runTestGeneration = true
		log.Infof("%d of %d real generations completed within the last %.0f minutes failed (non-NSFW), confirming with a test generation...", failed, completed, noSuccessfulGenerationWindow.Minutes())
	} else if asOf.Sub(lastSuccessfulGenerationTime) > noSuccessfulGenerationWindow {
		// Traffic is silent: probe.
		runTestGeneration = true
		log.Infof("No successful generation in the last %.0f minutes, creating a test generation...", noSuccessfulGenerationWindow.Minutes())
	}

	if runTestGeneration {
		b := retry.WithMaxRetries(testGenerationRetries, retry.NewExponential(testGenerationRetryBaseDelay))
		errTest := retry.Do(ctx, b, func(ctx context.Context) error {
			err := CreateTestGeneration(ctx, log, apiKey)
			if err != nil {
				log.Errorf("🧪 🔴 SC Worker test generation failed: %v", err)
				var tgErr *TestGenerationError
				if errors.As(err, &tgErr) && tgErr.IsNotWorkerFault() {
					// The API rejected the request (credits, auth, rate limit) or
					// the test is misconfigured — retrying won't change the answer
					// and it isn't the worker's fault.
					return err
				}
				return retry.RetryableError(err)
			}
			return nil
		})
		if errTest != nil {
			var tgErr *TestGenerationError
			if errors.As(errTest, &tgErr) && tgErr.IsNotWorkerFault() {
				// The test itself is broken, the worker state is unknown. Alert
				// loudly (Discord pings on non-healthy) but never fail over on it.
				workerHealthStatus = shared.UNKNOWN
				log.Errorf("🧪 🟡 Test generation failed for a reason that is not the worker's fault -> health UNKNOWN, not failing over: %v", errTest)
			} else {
				workerHealthStatus = shared.UNHEALTHY
				log.Infof("SC Worker test generation failed -> Assuming unhealthy")
			}
			healthErr = errTest
		}
	}

	// Hysteresis bookkeeping. UNKNOWN runs are inconclusive and freeze both
	// streaks: they neither build towards failover nor towards recovery.
	switch workerHealthStatus {
	case shared.HEALTHY:
		j.consecutiveHealthyChecks++
		j.consecutiveUnhealthyChecks = 0
	case shared.UNHEALTHY:
		j.consecutiveUnhealthyChecks++
		j.consecutiveHealthyChecks = 0
	}

	log.Infof("Done checking health in %dms", time.Since(asOf).Milliseconds())

	// Write health status to redis
	errRedis := j.Redis.SetWorkerHealth(workerHealthStatus)

	if errRedis != nil {
		log.Infof("🔴 Couldn't write SC Worker health status to Redis: %v", errRedis)
	} else {
		log.Infof("🟢 Wrote SC Worker health status to Redis: %d", workerHealthStatus)
	}

	// Failover, gated by hysteresis
	if workerHealthStatus == shared.UNHEALTHY && !isRunpodServerlessActive {
		if j.consecutiveUnhealthyChecks >= consecutiveUnhealthyChecksForFailover {
			err := j.Repo.EnableRunpodServerless()
			if err != nil {
				log.Errorf("🏃‍♂️‍➡️📦 🔴 Couldn't activate Runpod serverless: %v", err)
				runpodServerlessErr = err
			} else {
				log.Infof("🏃‍♂️‍➡️📦 🟢 Activated Runpod serverless after %d consecutive unhealthy checks", j.consecutiveUnhealthyChecks)
				isRunpodServerlessActive = true
				j.runpodAutoEnabled = true
			}
		} else {
			log.Infof("🏃‍♂️‍➡️📦 🟡 Unhealthy streak %d/%d — not failing over to Runpod serverless yet", j.consecutiveUnhealthyChecks, consecutiveUnhealthyChecksForFailover)
		}
	}

	// Auto-recovery: only undo a failover this job performed itself, so a
	// manually enabled Runpod (admin endpoint) is never fought by the cron.
	if workerHealthStatus == shared.HEALTHY && isRunpodServerlessActive && j.runpodAutoEnabled && j.consecutiveHealthyChecks >= consecutiveHealthyChecksForRecovery {
		err := j.Repo.DisableRunpodServerless()
		if err != nil {
			log.Errorf("🏃‍♂️‍➡️📦 🔴 Couldn't deactivate Runpod serverless: %v", err)
			runpodServerlessErr = err
		} else {
			log.Infof("🏃‍♂️‍➡️📦 🟢 Deactivated Runpod serverless after %d consecutive healthy checks", j.consecutiveHealthyChecks)
			isRunpodServerlessActive = false
			j.runpodAutoEnabled = false
		}
	}

	return j.Discord.SendDiscordNotificationIfNeeded(
		workerHealthStatus,
		generations,
		lastGenerationTime,
		lastSuccessfulGenerationTime,
		isRunpodServerlessActive,
		runpodServerlessErr,
		healthErr,
		asOf,
	)
}

// countRecentCompleted tallies real generations created within window of asOf
// that reached a terminal state. NSFW failures are excluded from both counts.
func countRecentCompleted(generations []*ent.Generation, asOf time.Time, window time.Duration) (succeeded int, failed int) {
	for _, g := range generations {
		if asOf.Sub(g.CreatedAt) > window {
			continue
		}
		switch g.Status {
		case generation.StatusSucceeded:
			succeeded++
		case generation.StatusFailed:
			if g.FailureReason == nil || *g.FailureReason != shared.NSFW_ERROR {
				failed++
			}
		}
	}
	return succeeded, failed
}

type RequestBody struct {
	Prompt     string `json:"prompt"`
	Width      int    `json:"width"`
	Height     int    `json:"height"`
	NumOutputs int    `json:"num_outputs"`
}

type ResponseBody struct {
	Outputs []struct {
		ID       string `json:"id"`
		URL      string `json:"url"`
		ImageURL string `json:"image_url"`
	} `json:"outputs"`
}

// TestGenerationError carries enough context to tell a broken worker apart
// from a broken test.
type TestGenerationError struct {
	// StatusCode is the HTTP status of the API response; 0 for transport-level
	// failures (timeouts, connection errors).
	StatusCode int
	// Misconfigured marks failures that happened before a request was even
	// possible (e.g. missing API key).
	Misconfigured bool
	Message       string
}

func (e *TestGenerationError) Error() string {
	if e.StatusCode > 0 {
		return fmt.Sprintf("status %d: %s", e.StatusCode, e.Message)
	}
	return e.Message
}

// IsNotWorkerFault reports whether the failure is attributable to the test
// itself (missing key, or a 4xx rejection: insufficient credits, auth, rate
// limit, nsfw) rather than the worker pipeline. These must never count as
// UNHEALTHY or trigger a Runpod failover.
func (e *TestGenerationError) IsNotWorkerFault() bool {
	return e.Misconfigured || (e.StatusCode >= 400 && e.StatusCode < 500)
}

// Response bodies are propagated into logs and the Discord embed; keep them
// under Discord's 1024-char embed field limit.
const maxTestGenErrorBodyLen = 500

func truncateForLog(b []byte) string {
	s := strings.TrimSpace(string(b))
	if len(s) > maxTestGenErrorBodyLen {
		return s[:maxTestGenErrorBodyLen] + "…"
	}
	return s
}

func CreateTestGeneration(ctx context.Context, log Logger, apiKey string) error {
	log.Infof("🧪 Creating test generation to check SC Worker health...")

	if apiKey == "" {
		log.Errorf("🧪 🔴 SC Worker tester API key not found")
		return &TestGenerationError{Misconfigured: true, Message: "SC Worker tester API key not found"}
	}

	url := "https://api.stablecog.com/v1/image/generation/create"
	prompt := "Mavi renkli bir bina"
	width := 1024
	height := 1024
	numOutputs := 1

	requestBody := RequestBody{
		Prompt:     prompt,
		Width:      width,
		Height:     height,
		NumOutputs: numOutputs,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		log.Errorf("🧪 🔴 Couldn't marshal json %v", err)
		debugString := fmt.Sprintf("%+v", requestBody)
		log.Errorf("🧪 🔴 Request body: %s", debugString)
		return &TestGenerationError{Misconfigured: true, Message: fmt.Sprintf("couldn't marshal request body: %v", err)}
	}

	// Pin a per-attempt deadline so a stalled API call can't drag the cron job
	// out beyond healthCheckBudget. The outer ctx is still honored — whichever
	// expires first wins.
	reqCtx, cancel := context.WithTimeout(ctx, testGenerationHTTPTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Errorf("🧪 🔴 Couldn't create request %v", err)
		return &TestGenerationError{Misconfigured: true, Message: fmt.Sprintf("couldn't create request: %v", err)}
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: testGenerationHTTPTimeout}
	resp, err := client.Do(req)
	if err != nil {
		log.Errorf("🧪 🔴 Couldn't send request %v", err)
		return &TestGenerationError{Message: fmt.Sprintf("request failed: %v", err)}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("🧪 🔴 Couldn't read response body %v", err)
		return &TestGenerationError{StatusCode: resp.StatusCode, Message: fmt.Sprintf("couldn't read response body: %v", err)}
	}

	// A non-2xx body is the actual diagnosis (insufficient_credits, nsfw_prompt,
	// rate limit, worker error...) — propagate it instead of swallowing it as
	// "no outputs in response".
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Errorf("🧪 🔴 Test generation request returned status %d: %s", resp.StatusCode, truncateForLog(body))
		return &TestGenerationError{StatusCode: resp.StatusCode, Message: truncateForLog(body)}
	}

	var responseBody ResponseBody
	err = json.Unmarshal(body, &responseBody)
	if err != nil {
		log.Errorf("🧪 🔴 Couldn't unmarshal response body %v: %s", err, truncateForLog(body))
		return &TestGenerationError{StatusCode: resp.StatusCode, Message: fmt.Sprintf("couldn't unmarshal response body: %v: %s", err, truncateForLog(body))}
	}

	if len(responseBody.Outputs) == 0 {
		log.Errorf("🧪 🔴 No outputs in response: %s", truncateForLog(body))
		return &TestGenerationError{StatusCode: resp.StatusCode, Message: fmt.Sprintf("no outputs in response: %s", truncateForLog(body))}
	}

	log.Infof("🧪 🟢 SC Worker test generation created: %s", responseBody.Outputs[0].ImageURL)

	return nil
}
