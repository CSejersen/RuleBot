package events

func (p *PlugUpdate) GetType() string {
	return p.Type
}

type PlugUpdate struct {
	ID       string          `json:"id"`
	IDV1     string          `json:"id_v1,omitempty"`
	Type     string          `json:"type"` // always "plug"
	Metadata *MetadataUpdate `json:"metadata,omitempty"`
	On       *OnUpdate       `json:"on,omitempty"`
}

func (p *PlugUpdate) ResolveStateChange() string {
	on := p.SafeOn()
	if on == nil {
		return ""
	}
	return "power_mode"
}

// SafeOn returns On/off state or nil if it does not exist
func (p *PlugUpdate) SafeOn() *bool {
	if p.On != nil {
		return &p.On.On
	}
	return nil
}
