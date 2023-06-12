//go:build integration
// +build integration

package fans_test

import (
	"testing"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kalverra/crazed-nft-fans/config"
	"github.com/kalverra/crazed-nft-fans/fans"
)

const fanCount = 5

var president *fans.President

func TestMain(m *testing.M) {
	err := config.ReadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Error reading config")
	}
	president, err = fans.NewPresident()
	if err != nil {
		log.Fatal().Err(err).Msg("Error creating president")
	}
	err = president.RecruitFans(fanCount)
	if err != nil {
		log.Fatal().Err(err).Msg("Error recruiting fans")
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

	err := president.RecruitFans(5)
	require.NoError(t, err, "Error recruiting fans")
}

func TestSearch(t *testing.T) {
	president.ActivateFans()
	fansSearching := 0
	for _, fan := range president.Fans() {
		if fan.IsSearching() {
			fansSearching++
		}
	}
	require.GreaterOrEqual(t, fansSearching, fanCount, "Expected at least %d fans to be searching", fanCount)
}

func TestStop(t *testing.T) {
	president.ActivateFans()
	president.StopFans()
	time.Sleep(100 * time.Millisecond)
	fansSearching, fansStopped := 0, 0
	for _, fan := range president.Fans() {
		if fan.IsSearching() {
			fansSearching++
		} else {
			fansStopped++
		}
	}
	require.GreaterOrEqual(t, fansStopped, fanCount, "Expected at least %d fans to be stopped, found %d searching", fanCount, fansSearching)
}

func TestActivateTimeSpan(t *testing.T) {
	president.ActivateFansTimeSpan(2 * time.Second)
	fansSearching, fansStopped := 0, 0
	for _, fan := range president.Fans() {
		if fan.IsSearching() {
			fansSearching++
		} else {
			fansStopped++
		}
	}
	assert.GreaterOrEqual(t, fansSearching, fanCount, "Expected at least %d fans to be searching, found %d stopped", fanCount, fansStopped)

	time.Sleep(4 * time.Second)
	fansSearching, fansStopped = 0, 0
	for _, fan := range president.Fans() {
		if fan.IsSearching() {
			fansSearching++
		} else {
			fansStopped++
		}
	}
	assert.GreaterOrEqual(t, fansStopped, fanCount, "Expected at least %d fans to be stopped, found %d searching", fanCount, fansSearching)
}

func TestActivateBlockSpan(t *testing.T) {
	president.ActivateFansBlockSpan(2)
	fansSearching, fansStopped := 0, 0
	for _, fan := range president.Fans() {
		if fan.IsSearching() {
			fansSearching++
		} else {
			fansStopped++
		}
	}
	assert.GreaterOrEqual(t, fansSearching, fanCount, "Expected at least %d fans to be searching, found %d stopped", fanCount, fansStopped)

	time.Sleep(4 * time.Second)
	fansSearching, fansStopped = 0, 0
	for _, fan := range president.Fans() {
		if fan.IsSearching() {
			fansSearching++
		} else {
			fansStopped++
		}
	}
	assert.GreaterOrEqual(t, fansStopped, fanCount, "Expected at least %d fans to be stopped, found %d searching", fanCount, fansSearching)
}
