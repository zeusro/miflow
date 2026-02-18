package ctrl

import (
	"fmt"

	"github.com/zeusro/miflow/internal/device"
)

// Controller 封装设备属性与动作操作，使用 miiot 规格常量。
type Controller struct {
	API *device.API
}

// New 创建 Controller。
func New(api *device.API) *Controller {
	return &Controller{API: api}
}

// spec 获取型号规格，未找到时返回零值。
func spec(model string) Spec {
	if s, ok := Specs[model]; ok {
		return s
	}
	return Spec{}
}

// SetOn 设置开关/插座/灯的开状态。
func (c *Controller) SetOn(did, model string, on bool) error {
	s := spec(model)
	if s.SiidSwitch == 0 && s.SiidLight == 0 {
		return fmt.Errorf("ctrl: model %s has no switch/light service", model)
	}
	siid := s.SiidSwitch
	if siid == 0 {
		siid = s.SiidLight
	}
	_, err := c.API.SetProps(did, [][3]interface{}{{siid, s.PiidOn, on}})
	return err
}

// GetOn 获取开关/插座/灯的开状态。
func (c *Controller) GetOn(did, model string) (bool, error) {
	s := spec(model)
	siid := s.SiidSwitch
	if siid == 0 {
		siid = s.SiidLight
	}
	if siid == 0 {
		return false, fmt.Errorf("ctrl: model %s has no switch/light service", model)
	}
	vals, err := c.API.GetProps(did, [][2]int{{siid, s.PiidOn}})
	if err != nil || len(vals) == 0 {
		return false, err
	}
	if b, ok := vals[0].(bool); ok {
		return b, nil
	}
	return false, nil
}

// Toggle 切换开关。
func (c *Controller) Toggle(did, model string) error {
	s := spec(model)
	if s.SiidSwitch == 0 || s.AiidToggle == 0 {
		return fmt.Errorf("ctrl: model %s has no toggle action", model)
	}
	_, err := c.API.Action(did, s.SiidSwitch, s.AiidToggle, nil)
	return err
}

// SetBrightness 设置亮度 0-100。
func (c *Controller) SetBrightness(did, model string, level int) error {
	s := spec(model)
	if s.SiidLight == 0 || s.PiidBrightness == 0 {
		return fmt.Errorf("ctrl: model %s has no brightness", model)
	}
	if level < 0 {
		level = 0
	}
	if level > 100 {
		level = 100
	}
	_, err := c.API.SetProps(did, [][3]interface{}{{s.SiidLight, s.PiidBrightness, level}})
	return err
}

// GetBrightness 获取亮度。
func (c *Controller) GetBrightness(did, model string) (int, error) {
	s := spec(model)
	if s.SiidLight == 0 || s.PiidBrightness == 0 {
		return 0, fmt.Errorf("ctrl: model %s has no brightness", model)
	}
	vals, err := c.API.GetProps(did, [][2]int{{s.SiidLight, s.PiidBrightness}})
	if err != nil || len(vals) == 0 {
		return 0, err
	}
	switch v := vals[0].(type) {
	case float64:
		return int(v), nil
	case int:
		return v, nil
	}
	return 0, nil
}

// TTS 音箱 TTS 播报。
func (c *Controller) TTS(did, model, text string) error {
	s := spec(model)
	if s.SiidVoiceAssistant == 0 || s.AiidExecuteText == 0 {
		return fmt.Errorf("ctrl: model %s has no TTS", model)
	}
	_, err := c.API.Action(did, s.SiidVoiceAssistant, s.AiidExecuteText, []interface{}{text})
	return err
}

// SetVolume 设置音量 0-100。
func (c *Controller) SetVolume(did, model string, level int) error {
	s := spec(model)
	if s.SiidSpeaker == 0 || s.PiidVolume == 0 {
		return fmt.Errorf("ctrl: model %s has no volume", model)
	}
	if level < 0 {
		level = 0
	}
	if level > 100 {
		level = 100
	}
	_, err := c.API.SetProps(did, [][3]interface{}{{s.SiidSpeaker, s.PiidVolume, level}})
	return err
}

// GetVolume 获取音量。
func (c *Controller) GetVolume(did, model string) (int, error) {
	s := spec(model)
	if s.SiidSpeaker == 0 || s.PiidVolume == 0 {
		return 0, fmt.Errorf("ctrl: model %s has no volume", model)
	}
	vals, err := c.API.GetProps(did, [][2]int{{s.SiidSpeaker, s.PiidVolume}})
	if err != nil || len(vals) == 0 {
		return 0, err
	}
	switch v := vals[0].(type) {
	case float64:
		return int(v), nil
	case int:
		return v, nil
	}
	return 0, nil
}

// SetMute 设置静音。
func (c *Controller) SetMute(did, model string, mute bool) error {
	s := spec(model)
	if s.SiidSpeaker == 0 || s.PiidMute == 0 {
		return fmt.Errorf("ctrl: model %s has no mute", model)
	}
	_, err := c.API.SetProps(did, [][3]interface{}{{s.SiidSpeaker, s.PiidMute, mute}})
	return err
}

// GetMute 获取静音状态。
func (c *Controller) GetMute(did, model string) (bool, error) {
	s := spec(model)
	if s.SiidSpeaker == 0 || s.PiidMute == 0 {
		return false, fmt.Errorf("ctrl: model %s has no mute", model)
	}
	vals, err := c.API.GetProps(did, [][2]int{{s.SiidSpeaker, s.PiidMute}})
	if err != nil || len(vals) == 0 {
		return false, err
	}
	if b, ok := vals[0].(bool); ok {
		return b, nil
	}
	return false, nil
}

// Play 播放。
func (c *Controller) Play(did, model string) error {
	s := spec(model)
	if s.SiidPlayControl == 0 || s.AiidPlay == 0 {
		return fmt.Errorf("ctrl: model %s has no play action", model)
	}
	_, err := c.API.Action(did, s.SiidPlayControl, s.AiidPlay, nil)
	return err
}

// Pause 暂停。
func (c *Controller) Pause(did, model string) error {
	s := spec(model)
	if s.SiidPlayControl == 0 || s.AiidPause == 0 {
		return fmt.Errorf("ctrl: model %s has no pause action", model)
	}
	_, err := c.API.Action(did, s.SiidPlayControl, s.AiidPause, nil)
	return err
}

// Next 下一曲。
func (c *Controller) Next(did, model string) error {
	s := spec(model)
	if s.SiidPlayControl == 0 || s.AiidNext == 0 {
		return fmt.Errorf("ctrl: model %s has no next action", model)
	}
	_, err := c.API.Action(did, s.SiidPlayControl, s.AiidNext, nil)
	return err
}

// Previous 上一曲。
func (c *Controller) Previous(did, model string) error {
	s := spec(model)
	if s.SiidPlayControl == 0 || s.AiidPrevious == 0 {
		return fmt.Errorf("ctrl: model %s has no previous action", model)
	}
	_, err := c.API.Action(did, s.SiidPlayControl, s.AiidPrevious, nil)
	return err
}

// TVTurnOff 电视关机。
func (c *Controller) TVTurnOff(did, model string) error {
	s := spec(model)
	if s.SiidTV == 0 || s.AiidTurnOff == 0 {
		return fmt.Errorf("ctrl: model %s has no turn off action", model)
	}
	_, err := c.API.Action(did, s.SiidTV, s.AiidTurnOff, nil)
	return err
}

// GetOccupancy 获取 occupancy 状态。
func (c *Controller) GetOccupancy(did, model string) (interface{}, error) {
	s := spec(model)
	if s.SiidOccupancy == 0 || s.PiidStatus == 0 {
		return nil, fmt.Errorf("ctrl: model %s has no occupancy", model)
	}
	vals, err := c.API.GetProps(did, [][2]int{{s.SiidOccupancy, s.PiidStatus}})
	if err != nil || len(vals) == 0 {
		return nil, err
	}
	return vals[0], nil
}

// SetSwitchChannel 多通道开关指定通道。
func (c *Controller) SetSwitchChannel(did, model string, channel int, on bool) error {
	s := spec(model)
	if len(s.SwitchChannels) == 0 {
		return c.SetOn(did, model, on)
	}
	if channel < 0 || channel >= len(s.SwitchChannels) {
		return fmt.Errorf("ctrl: channel %d out of range [0,%d)", channel, len(s.SwitchChannels))
	}
	siid := s.SwitchChannels[channel]
	_, err := c.API.SetProps(did, [][3]interface{}{{siid, s.PiidOn, on}})
	return err
}
