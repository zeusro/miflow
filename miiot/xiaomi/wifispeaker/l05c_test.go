package wifispeaker

import (
	"testing"
)

// TestL05C_GetVolume 测试场景：获取设备当前音量，校验返回值在 [0,100] 范围内。
func TestL05C_GetVolume(t *testing.T) {
	api, dev := setupAPIForModel(t, ModelL05C)
	if api == nil || dev == nil {
		return
	}
	testGetVolume(t, api, dev)
}

// TestL05C_SetVolumeGetVolume 测试场景：设置音量为 50 后读取，校验 SetVolume 与 GetVolume 一致性。
func TestL05C_SetVolumeGetVolume(t *testing.T) {
	api, dev := setupAPIForModel(t, ModelL05C)
	if api == nil || dev == nil {
		return
	}
	testSetVolumeGetVolume(t, api, dev)
}

// TestL05C_SetVolumeBoundary 测试场景：设置边界音量 0 和 100，校验设备对边界值的处理。
func TestL05C_SetVolumeBoundary(t *testing.T) {
	api, dev := setupAPIForModel(t, ModelL05C)
	if api == nil || dev == nil {
		return
	}
	testSetVolumeBoundary(t, api, dev)
}

// TestL05C_GetMute 测试场景：获取设备当前静音状态。
func TestL05C_GetMute(t *testing.T) {
	api, dev := setupAPIForModel(t, ModelL05C)
	if api == nil || dev == nil {
		return
	}
	testGetMute(t, api, dev)
}

// TestL05C_SetMuteGetMute 测试场景：切换静音状态后读取，校验 SetMute 与 GetMute 一致性。
func TestL05C_SetMuteGetMute(t *testing.T) {
	api, dev := setupAPIForModel(t, ModelL05C)
	if api == nil || dev == nil {
		return
	}
	testSetMuteGetMute(t, api, dev)
}

// TestL05C_TTS 测试场景：调用 TTS 播报指定文本，校验语音播报接口可用。
func TestL05C_TTS(t *testing.T) {
	api, dev := setupAPIForModel(t, ModelL05C)
	if api == nil || dev == nil {
		return
	}
	testTTS(t, api, dev)
}

// TestL05C_Play 测试场景：发送播放命令，校验 Play 接口可用。
func TestL05C_Play(t *testing.T) {
	api, dev := setupAPIForModel(t, ModelL05C)
	if api == nil || dev == nil {
		return
	}
	testPlay(t, api, dev)
}

// TestL05C_Pause 测试场景：发送暂停命令，校验 Pause 接口可用。
func TestL05C_Pause(t *testing.T) {
	api, dev := setupAPIForModel(t, ModelL05C)
	if api == nil || dev == nil {
		return
	}
	testPause(t, api, dev)
}

// TestL05C_PlayPauseSequence 测试场景：依次执行 Play→Pause→Play，校验播放控制序列。
func TestL05C_PlayPauseSequence(t *testing.T) {
	api, dev := setupAPIForModel(t, ModelL05C)
	if api == nil || dev == nil {
		return
	}
	testPlayPauseSequence(t, api, dev)
}

// TestL05C_Next 测试场景：发送下一曲命令，校验 Next 接口可用。
func TestL05C_Next(t *testing.T) {
	api, dev := setupAPIForModel(t, ModelL05C)
	if api == nil || dev == nil {
		return
	}
	testNext(t, api, dev)
}

// TestL05C_Previous 测试场景：发送上一曲命令，校验 Previous 接口可用。
func TestL05C_Previous(t *testing.T) {
	api, dev := setupAPIForModel(t, ModelL05C)
	if api == nil || dev == nil {
		return
	}
	testPrevious(t, api, dev)
}
