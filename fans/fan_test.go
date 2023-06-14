//go:build integration
// +build integration

package fans_test

import (
	"context"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kalverra/crazed-nft-fans/config"
	"github.com/kalverra/crazed-nft-fans/fans"
)

const fanCount = 5

func TestMain(m *testing.M) {
	err := config.ReadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Error reading config")
	}
	m.Run()
}

func TestNewFan(t *testing.T) {
	t.Parallel()

	fan, err := fans.New(config.Manic)
	require.NoError(t, err, "Error creating new fan")
	require.NotNil(t, fan, "Fan should not be nil")
}

func TestRecruit(t *testing.T) {
	t.Parallel()

	president, err := fans.NewPresident()
	require.NoError(t, err, "Error creating new president")
	err = president.RecruitFans(5)
	require.NoError(t, err, "Error recruiting fans")
}

func setupPresident(t *testing.T) *fans.President {
	president, err := fans.NewPresident()
	require.NoError(t, err, "Error creating new president")
	require.NotNil(t, president, "President should not be nil")
	err = president.RecruitFans(fanCount)
	require.NoError(t, err, "Error recruiting fans")
	return president
}

func TestSearch(t *testing.T) {
	president := setupPresident(t)
	president.ActivateFans()
	fansSearching, _ := countFansStatus(t, president)
	require.GreaterOrEqual(t, fansSearching, fanCount, "Expected at least %d fans to be searching", fanCount)
}

func TestStop(t *testing.T) {
	t.Parallel()

	president := setupPresident(t)
	president.ActivateFans()
	president.StopFans()
	time.Sleep(100 * time.Millisecond)
	fansSearching, fansStopped := countFansStatus(t, president)
	require.GreaterOrEqual(t, fansStopped, fanCount, "Expected at least %d fans to be stopped, found %d searching", fanCount, fansSearching)
}

func TestActivateTimeSpan(t *testing.T) {
	t.Parallel()

	expectedDuration, failureDuration := time.Second, 2*time.Second
	president := setupPresident(t)
	startTime := time.Now()
	president.ActivateFansTimeSpan(expectedDuration)
	fansSearching, fansStopped := countFansStatus(t, president)
	assert.GreaterOrEqual(t, fansSearching, fanCount, "Expected at least %d fans to be searching, found %d stopped", fanCount, fansStopped)

	for range time.Tick(time.Millisecond * 50) {
		fansSearching, fansStopped = countFansStatus(t, president)
		if fansStopped == fanCount || time.Since(startTime) >= failureDuration {
			break
		}
	}
	require.GreaterOrEqual(t, time.Since(startTime), expectedDuration, "Expected fans to stop searching after at least %s, but they're still going after %s", failureDuration, time.Since(startTime))
	require.Less(t, time.Since(startTime), failureDuration, "Expected fans to stop searching after at least %s, but they're still going after %s", failureDuration, time.Since(startTime))
}

func TestActivateBlockSpan(t *testing.T) {
	t.Parallel()

	president := setupPresident(t)
	newHeaders := make(chan *types.Header)
	blocksSeen, expectedBlocks, maxBlocks := 0, 1, 5
	blockSub, err := president.Client.SubscribeNewHead(context.Background(), newHeaders)
	require.NoError(t, err, "Error subscribing to new blocks")
	president.ActivateFansBlockSpan(expectedBlocks)

testLoop:
	for {
		select {
		case <-time.After(5 * time.Second):
			require.Fail(t, "Timed out waiting for new blocks")
		case err := <-blockSub.Err():
			require.NoError(t, err, "Error subscribing to new blocks")
		case <-newHeaders:
			blocksSeen++
			_, fansStopped := countFansStatus(t, president)
			if fansStopped == fanCount {
				break testLoop
			}
			if blocksSeen >= maxBlocks {
				break testLoop
			}
		}
	}
	require.GreaterOrEqual(t, blocksSeen, 2, "Expected to see at least %d blocks before fans stopped, but saw %d", 2, blocksSeen)
	require.Less(t, blocksSeen, maxBlocks, "Expected to see less than %d blocks before fans stopped, but saw %d", maxBlocks, blocksSeen)
}

func countFansStatus(t *testing.T, president *fans.President) (fansSearching int, fansStopped int) {
	t.Helper()
	for _, fan := range president.Fans() {
		if fan.IsSearching() {
			fansSearching++
		} else {
			fansStopped++
		}
	}
	return
}
