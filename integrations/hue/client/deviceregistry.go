package client

// DeviceRegistry: human-readableID -> hueID
type DeviceRegistry struct {
	lights map[string]string
	rooms  map[string]string
}

func (r *DeviceRegistry) Resolve(typ, humanId string) (string, bool) {
	switch typ {
	case "light":
		val, ok := r.lights[humanId]
		return val, ok
	case "room":
		val, ok := r.rooms[humanId]
		return val, ok
	}
	return "", false
}

// InitRegistry init registry of human-readable id's
func (c *Client) InitRegistry() error {
	c.DeviceRegistry = DeviceRegistry{
		lights: make(map[string]string),
		rooms:  make(map[string]string),
	}

	// Lights
	lights, err := c.Lights()
	if err != nil {
		return err
	}
	for _, light := range lights {
		c.DeviceRegistry.lights[light.Metadata.Name] = light.ID
	}

	// TODO: add more entity types to the DeviceRegistry
	return nil
}
