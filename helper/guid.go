package helper

import (
	"database/sql/driver"
	"fmt"

	"github.com/Microsoft/go-winio/pkg/guid"
)

type GUIDWrapper struct {
	guid.GUID
}

// Implement Valuer interface for GUIDWrapper
func (g GUIDWrapper) Value() (driver.Value, error) {
	return g.String(), nil
}

func NewGUIDWrapperFromString(s string) GUIDWrapper {
	id, _ := guid.FromString(s)
	return GUIDWrapper{id}
}

// Implement Scanner interface for GUIDWrapper
func (g *GUIDWrapper) Scan(value interface{}) error {
	if value == nil {
		return fmt.Errorf("NULL GUID")
	}

	switch v := value.(type) {
	case []byte:
		if len(v) != 16 {
			return fmt.Errorf("invalid byte length for GUID: %d", len(v))
		}
		// Convert to GUID string representation
		g.GUID = guid.GUID{
			Data1: uint32(v[0]) | uint32(v[1])<<8 | uint32(v[2])<<16 | uint32(v[3])<<24,
			Data2: uint16(v[4]) | uint16(v[5])<<8,
			Data3: uint16(v[6]) | uint16(v[7])<<8,
			Data4: [8]byte{v[8], v[9], v[10], v[11], v[12], v[13], v[14], v[15]},
		}
		return nil
	case string:
		parsedGUID, err := guid.FromString(v)
		if err != nil {
			return err
		}
		g.GUID = parsedGUID
	default:
		return fmt.Errorf("unsupported type for GUID: %T", value)
	}

	return nil
}
