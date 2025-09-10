package apiclient

type Registry struct {
	lights map[string]string
	groups map[string]string
}

func (r *Registry) Resolve(typ, id string) (string, bool) {
	switch typ {
	case "hue.light":
		val, ok := r.lights[id]
		return val, ok
	case "hue.group":
		val, ok := r.groups[id]
		return val, ok
	}
	return "", false
}

// TODO: implement this
func (c *ApiClient) InitRegistry() {
}
