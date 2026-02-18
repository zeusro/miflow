package wifispeaker

import (
	"testing"
)

func TestL05C(t *testing.T) {
	api, dev := setupAPIForModel(t, ModelL05C)
	if api == nil || dev == nil {
		return
	}
	t.Run("GetVolume", func(t *testing.T) { testGetVolume(t, api, dev) })
	t.Run("SetVolumeGetVolume", func(t *testing.T) { testSetVolumeGetVolume(t, api, dev) })
	t.Run("SetVolumeBoundary", func(t *testing.T) { testSetVolumeBoundary(t, api, dev) })
	t.Run("GetMute", func(t *testing.T) { testGetMute(t, api, dev) })
	t.Run("SetMuteGetMute", func(t *testing.T) { testSetMuteGetMute(t, api, dev) })
	t.Run("TTS", func(t *testing.T) { testTTS(t, api, dev) })
	t.Run("Play", func(t *testing.T) { testPlay(t, api, dev) })
	t.Run("Pause", func(t *testing.T) { testPause(t, api, dev) })
	t.Run("PlayPauseSequence", func(t *testing.T) { testPlayPauseSequence(t, api, dev) })
	t.Run("Next", func(t *testing.T) { testNext(t, api, dev) })
	t.Run("Previous", func(t *testing.T) { testPrevious(t, api, dev) })
}
