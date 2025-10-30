package types

type RoomGet struct {
	ID       string               `json:"id"`              // UUID pattern
	IDV1     string               `json:"id_v1,omitempty"` // optional Clip v1 identifier
	Children []ResourceIdentifier `json:"children"`        // child devices/services
	Services []ResourceIdentifier `json:"services"`        // aggregated services
	Type     string               `json:"type"`            // always "room"
	Metadata RoomMetadataGet      `json:"metadata"`        // configuration object for room
}

type RoomMetadataGet struct {
	Name      string `json:"name"`      // human-readable name
	Archetype string `json:"archetype"` // e.g., living_room, kitchen, bedroom, etc.
}

func (r *RoomGet) GetType() string {
	return r.Type
}

func (r *RoomGet) GetID() string {
	return r.ID
}

func (r *RoomGet) GetGroupedLightID() (string, bool) {
	for _, svc := range r.Services {
		if svc.RType == "grouped_light" {
			return svc.RID, true
		}
	}
	return "", false
}
