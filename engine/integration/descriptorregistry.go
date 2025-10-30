package integration

const (
	CapabilityAudio     = "audio"
	CapabilityLighting  = "lighting"
	CapabilityDiscovery = "discovery"
	CapabilityControl   = "control"
)

type IntegrationDescRegistry struct {
	Available map[string]IntegrationDescriptor
}

func NewIntegrationRegistry() *IntegrationDescRegistry {
	return &IntegrationDescRegistry{
		Available: make(map[string]IntegrationDescriptor),
	}
}

func (r *IntegrationDescRegistry) Register(descriptor IntegrationDescriptor) {
	r.Available[descriptor.Name] = descriptor
}

func (r *IntegrationDescRegistry) List() []IntegrationDescriptor {
	result := make([]IntegrationDescriptor, 0, len(r.Available))
	for _, d := range r.Available {
		result = append(result, d)
	}
	return result
}

func (r *IntegrationDescRegistry) Get(name string) (IntegrationDescriptor, bool) {
	d, ok := r.Available[name]
	return d, ok
}
