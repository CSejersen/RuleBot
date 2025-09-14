package types

type RoomPut struct {
	Children []ResourceIdentifier `json:"children,omitempty"` // child devices/services, optional
	Services []ResourceIdentifier `json:"services,omitempty"` // aggregated services, optional
	Type     *string              `json:"type,omitempty"`     // always room
	Metadata *RoomMetadataPut     `json:"metadata,omitempty"` // configuration object for room, optional
}

type RoomMetadataPut struct {
	Name      *string `json:"name,omitempty"`      // human-readable name, optional
	Archetype *string `json:"archetype,omitempty"` // e.g., living_room, kitchen, bedroom, etc., optional
}
