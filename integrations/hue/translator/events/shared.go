package events

type Owner struct {
	RID   string `json:"rid"`
	RType string `json:"rtype"`
}

type On struct {
	On bool `json:"on"`
}

type Dimming struct {
	Brightness float64 `json:"brightness"`
}

type DimmingDelta struct {
	Action          string  `json:"action"`
	BrightnessDelta float64 `json:"brightness_delta"`
}

type Metadata struct {
	Name string `json:"name"`
}
