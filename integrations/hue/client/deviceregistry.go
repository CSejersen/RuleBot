package client

import "go.uber.org/zap"

// ResourceRegistry: human-readableID -> hueID
type ResourceRegistry struct {
	lights map[string]string
	rooms  map[string]string
	scenes map[string]string
}

func (r *ResourceRegistry) Resolve(typ, humanId string) (string, bool) {
	switch typ {
	case "light":
		val, ok := r.lights[humanId]
		return val, ok
	case "room":
		val, ok := r.rooms[humanId]
		return val, ok
	case "scene":
		val, ok := r.scenes[humanId]
		return val, ok
	}
	return "", false
}

// InitResourceRegistry init registry of human-readable id's
func (c *Client) InitResourceRegistry() error {
	c.ResourceRegistry = ResourceRegistry{
		lights: make(map[string]string),
		rooms:  make(map[string]string),
		scenes: make(map[string]string),
	}

	// Lights
	lights, err := c.Lights()
	if err != nil {
		return err
	}
	for _, light := range lights {
		c.ResourceRegistry.lights[light.Metadata.Name] = light.ID
	}
	c.Logger.Debug("registered lights", zap.Int("total_amount", len(lights)))

	// Rooms
	rooms, err := c.Rooms()
	if err != nil {
		return err
	}
	for _, room := range rooms {
		c.ResourceRegistry.rooms[room.Metadata.Name] = room.ID
	}
	c.Logger.Debug("registered rooms", zap.Int("total_amount", len(rooms)))

	// Scenes
	scenes, err := c.Scenes()
	if err != nil {
		return err
	}
	for _, scene := range scenes {
		c.ResourceRegistry.scenes[scene.ID] = scene.ID
	}
	c.Logger.Debug("registered scenes", zap.Int("total_amount", len(scenes)))

	return nil
}
