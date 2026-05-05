package dto

type AppendRequest struct {
	Kind      string `json:"kind"`
	ID        string `json:"id"`
	EncodedID uint64 `json:"encoded_id"`
	Shard     uint16 `json:"shard"`
}

type LookupResponse struct {
	Found  bool   `json:"found"`
	Shard  uint8  `json:"shard,omitempty"`
	Source string `json:"source"`
}

type AppendResponse struct {
	Added        bool `json:"added"`
	DeltaEntries int  `json:"delta_entries"`
}
