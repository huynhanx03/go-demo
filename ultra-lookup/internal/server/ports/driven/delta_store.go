package driven

import "ultra-lookup/internal/server/core/entity"

type DeltaStore interface {
	Lookup(entry entity.Entry) (uint8, bool)
	Append(entry entity.Entry) (bool, error)
	Len() int
}
