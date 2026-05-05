package lookup

import "fmt"

const (
	idWidth      = 10
	base36Radix  = 36
	maxEncoded52 = (uint64(1) << 52) - 1
)

func EncodeBase36ID(id string) (uint64, error) {
	if len(id) != idWidth {
		return 0, fmt.Errorf("lookup: id %q length=%d, want=%d", id, len(id), idWidth)
	}

	var v uint64
	for i := 0; i < len(id); i++ {
		d, ok := base36Digit(id[i])
		if !ok {
			return 0, fmt.Errorf("lookup: id %q has invalid char %q at pos=%d", id, id[i], i)
		}
		v = v*base36Radix + uint64(d)
	}

	if v > maxEncoded52 {
		return 0, fmt.Errorf("lookup: id %q encoded overflow", id)
	}
	return v, nil
}

func FormatBase36ID(v uint64) (string, error) {
	if v > maxEncoded52 {
		return "", fmt.Errorf("lookup: encoded id %d overflows base36-10", v)
	}

	var out [idWidth]byte
	for i := idWidth - 1; i >= 0; i-- {
		d := v % base36Radix
		if d < 10 {
			out[i] = byte('0' + d)
		} else {
			out[i] = byte('A' + d - 10)
		}
		v /= base36Radix
	}
	return string(out[:]), nil
}

func base36Digit(c byte) (uint8, bool) {
	switch {
	case c >= '0' && c <= '9':
		return c - '0', true
	case c >= 'A' && c <= 'Z':
		return c - 'A' + 10, true
	default:
		return 0, false
	}
}
