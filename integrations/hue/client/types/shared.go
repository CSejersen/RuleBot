package types

type PutResponse struct {
	Errors []ApiError           `json:"errors"`
	Data   []ResourceIdentifier `json:"data"`
}

type ApiError struct {
	Description string `json:"description"`
}

type ResourceIdentifier struct {
	RID   string `json:"rid"`   // UUID of referenced resource
	RType string `json:"rtype"` // type of resource, e.g., device, grouped_light, etc.
}

type XY struct {
	X float64 `json:"x"` // 0-1
	Y float64 `json:"y"` // 0-1
}
