package lookup

import "testing"

func TestEncodeBase36ID(t *testing.T) {
	v, err := EncodeBase36ID("000000000A")
	if err != nil {
		t.Fatalf("EncodeBase36ID(): %v", err)
	}
	if v != 10 {
		t.Fatalf("encoded=%d want=10", v)
	}

	if _, err := EncodeBase36ID("000000000a"); err == nil {
		t.Fatal("lowercase id must fail")
	}
	if _, err := EncodeBase36ID("ABC"); err == nil {
		t.Fatal("short id must fail")
	}
}

func TestFormatBase36ID(t *testing.T) {
	id, err := FormatBase36ID(10)
	if err != nil {
		t.Fatalf("FormatBase36ID(): %v", err)
	}
	if id != "000000000A" {
		t.Fatalf("id=%q want=%q", id, "000000000A")
	}

	encoded, err := EncodeBase36ID(id)
	if err != nil {
		t.Fatalf("EncodeBase36ID(): %v", err)
	}
	if encoded != 10 {
		t.Fatalf("encoded=%d want=10", encoded)
	}
}

func TestBuildAndLookupStringID(t *testing.T) {
	records := []Record{
		{Kind: KindCustomer, ID: "0000000001", Shard: 7},
		{Kind: KindCustomer, ID: "0000000002", Shard: 9},
		{Kind: KindAccount, ID: "000000000A", Shard: 30},
		{Kind: KindAccount, ID: "000000000B", Shard: 31},
	}

	tests := []struct {
		kind      Kind
		id        string
		want      uint8
		wantFound bool
	}{
		{kind: KindCustomer, id: "0000000001", want: 7, wantFound: true},
		{kind: KindCustomer, id: "0000000002", want: 9, wantFound: true},
		{kind: KindAccount, id: "000000000A", want: 30, wantFound: true},
		{kind: KindAccount, id: "000000000B", want: 31, wantFound: true},
		{kind: KindCustomer, id: "000000000A", wantFound: false},
		{kind: KindAccount, id: "0000000001", wantFound: false},
		{kind: KindCustomer, id: "00000000ZZ", wantFound: false},
		{kind: KindCustomer, id: "bad-id", wantFound: false},
	}

	table, err := Build(records, DefaultConfig())
	if err != nil {
		t.Fatalf("Build(): %v", err)
	}
	for _, tc := range tests {
		got, ok := table.Lookup(tc.kind, tc.id)
		if ok != tc.wantFound {
			t.Fatalf("Lookup(%v, %q): found=%v want=%v", tc.kind, tc.id, ok, tc.wantFound)
		}
		if ok && got != tc.want {
			t.Fatalf("Lookup(%v, %q): shard=%d want=%d", tc.kind, tc.id, got, tc.want)
		}
	}
}

func TestDuplicateRecordIsRejected(t *testing.T) {
	records := []Record{
		{Kind: KindCustomer, ID: "0000000001", Shard: 7},
		{Kind: KindCustomer, ID: "0000000001", Shard: 8},
	}

	if _, err := Build(records, DefaultConfig()); err == nil {
		t.Fatal("Build() succeeded for duplicate customer ID; want error")
	}
}

func TestBuildEncodedAndMemory(t *testing.T) {
	records := []EncodedRecord{
		{Kind: KindCustomer, ID: 1, Shard: 1},
		{Kind: KindCustomer, ID: 2, Shard: 2},
		{Kind: KindAccount, ID: 9, Shard: 3},
	}

	table, err := BuildEncoded(records, DefaultConfig())
	if err != nil {
		t.Fatalf("BuildEncoded(): %v", err)
	}

	stats := table.MemoryStats()
	if stats.ValueBytes < 3 {
		t.Fatalf("ValueBytes=%d want>=3", stats.ValueBytes)
	}
	if stats.KeyBytes == 0 {
		t.Fatal("KeyBytes must be > 0")
	}
}
