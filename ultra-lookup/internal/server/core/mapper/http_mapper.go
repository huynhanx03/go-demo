package mapper

import (
	"net/http"
	"strconv"

	"ultra-lookup/internal/server/core/domain"
	"ultra-lookup/internal/server/core/dto"
	"ultra-lookup/internal/server/core/entity"
)

func LookupInputFromRequest(r *http.Request) (kind string, id string, encodedID uint64) {
	q := r.URL.Query()
	return q.Get("kind"), q.Get("id"), parseOptionalUint64(q.Get("encoded_id"))
}

func EntryFromAppendRequest(req dto.AppendRequest) (entity.Entry, error) {
	kind, err := domain.ParseKind(req.Kind)
	if err != nil {
		return entity.Entry{}, err
	}
	id, err := domain.ParseID(req.ID, req.EncodedID)
	if err != nil {
		return entity.Entry{}, err
	}
	return entity.Entry{
		Kind:  kind,
		ID:    id,
		Shard: uint8(req.Shard),
	}, nil
}

func LookupResponseFromResult(result entity.LookupResult) dto.LookupResponse {
	return dto.LookupResponse{
		Found:  result.Found,
		Shard:  result.Shard,
		Source: result.Source,
	}
}

func parseOptionalUint64(s string) uint64 {
	if s == "" {
		return 0
	}
	v, _ := strconv.ParseUint(s, 10, 64)
	return v
}
