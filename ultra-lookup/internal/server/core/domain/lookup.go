package domain

import (
	"fmt"

	"ultra-lookup/internal/lookup"
)

const (
	LookupSourceBase  = "base"
	LookupSourceDelta = "delta"
	LookupSourceNone  = "none"
)

func ParseKind(s string) (lookup.Kind, error) {
	switch s {
	case "customer":
		return lookup.KindCustomer, nil
	case "account":
		return lookup.KindAccount, nil
	default:
		return 0, fmt.Errorf("kind must be customer or account")
	}
}

func ParseID(id string, encodedID uint64) (uint64, error) {
	if id != "" {
		return lookup.EncodeBase36ID(id)
	}
	if encodedID == 0 {
		return 0, fmt.Errorf("id or encoded_id is required")
	}
	return encodedID, nil
}
