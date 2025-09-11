package actionexecutor

// Registry: human-readableID -> hueID
type Registry struct {
	lights map[string]string
	rooms  map[string]string
}

func (r *Registry) Resolve(typ, humanId string) (string, bool) {
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
func (e *Executor) InitRegistry() error {
	e.Registry = Registry{
		lights: make(map[string]string),
		rooms:  make(map[string]string),
	}

	// Lights
	lights, err := e.Client.Lights()
	if err != nil {
		return err
	}
	for _, light := range lights {
		e.Registry.lights[light.Metadata.Name] = light.ID
	}

	// TODO: add more entity types to the Registry
	return nil
}
