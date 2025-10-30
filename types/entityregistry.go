package types

type EntityRegistry interface {
	Register(externalID, entityID string)
	Resolve(externalID string) (string, bool)
	ResolveExternalID(entityID string) (string, bool)
}
