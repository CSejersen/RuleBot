package client

type Light struct {
	Type     string        `json:"type"`
	ID       string        `json:"id"`
	Metadata LightMetadata `json:"metadata"`
	On       LightOn       `json:"on"`
	Dimming  LightDimming  `json:"dimming"`
}

type LightStepBrightnessRequest struct {
	DimmingDelta DimmingDelta `json:"dimming_delta"`
}

type LightDimming struct {
	Brightness float64 `json:"brightness"`
}

type DimmingDelta struct {
	Action          string  `json:"action"`
	BrightnessDelta float64 `json:"brightness_delta"`
}

type LightOn struct {
	On bool
}

type LightMetadata struct {
	Name string `json:"name"`
}

type getLightsResponse struct {
	Data []Light `json:"data"`
}

type apiResponse struct {
	Errors []SetBrightnessError `json:"errors"`
}

type SetBrightnessError struct {
	Error string `json:"error"`
}
