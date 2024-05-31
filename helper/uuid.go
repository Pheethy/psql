package helper

import (
	"database/sql/driver"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/globalsign/mgo/bson"
	"github.com/gofrs/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	byteGroups = []int{8, 4, 4, 4, 12}
)

var uuidBinaryKind = bson.BinaryUUID
var bsonUUIDType = reflect.TypeOf(primitive.Binary{})

type ZeroUUID uuid.UUID

type NullUUID struct {
	UUID  ZeroUUID
	Valid bool
}

func (zu *NullUUID) Scan(src interface{}) error {
	if src == nil {
		zu.UUID, zu.Valid = ZeroUUID(uuid.UUID{}), false
		return nil
	}

	// Delegate to UUID Scan function
	zu.Valid = true
	return zu.UUID.Scan(src)
}

func (zu NullUUID) Value() (driver.Value, error) {
	if !zu.Valid {
		return nil, nil
	}
	// Delegate to UUID Value function
	return zu.UUID.Value()
}

func NewZeroUUIDFromstring(uidStr string) (ZeroUUID, error) {
	uid, err := uuid.FromString(uidStr)
	if err != nil {
		return ZeroUUID(uuid.Nil), nil
	}
	return ZeroUUID(uid), nil
}

func NewZeroUUIDFromUUID(uid *uuid.UUID) (ZeroUUID, error) {
	if uid == nil {
		return ZeroUUID(uuid.Nil), nil
	}

	return ZeroUUID(*uid), nil
}

func NewV4() ZeroUUID {
	uid, _ := uuid.NewV4()
	return ZeroUUID(uid)
}

func (zu ZeroUUID) IsZero() bool {
	if zu == ZeroUUID((uuid.UUID{})) {
		return true
	}

	return false
}

func (zu ZeroUUID) ToUUID() *uuid.UUID {
	if zu == ZeroUUID(uuid.Nil) {
		return nil
	}

	uid := uuid.UUID(zu)
	return &uid
}

func (zu ZeroUUID) ToBsonBinary() *bson.Binary {
	if zu == ZeroUUID(uuid.Nil) {
		return nil
	}

	uid := uuid.UUID(zu)
	return &bson.Binary{
		Kind: bson.BinaryUUID,
		Data: uid.Bytes(),
	}
}

func (zu ZeroUUID) NullUUID() NullUUID {
	var nullUID = NullUUID{}
	if zu == ZeroUUID((uuid.UUID{})) {
		nullUID.UUID = ZeroUUID((uuid.UUID{}))
		nullUID.Valid = false
		return nullUID
	}

	nullUID.UUID = zu
	nullUID.Valid = true
	return nullUID
}

func (zu ZeroUUID) Interface() interface{} {
	if zu == ZeroUUID((uuid.UUID{})) {
		return nil
	}

	return zu
}

func (zu ZeroUUID) MarshalJSON() ([]byte, error) {
	if zu == ZeroUUID((uuid.UUID{})) {
		return json.Marshal("")
	}
	return json.Marshal(uuid.UUID(zu).String())
}

func (zu ZeroUUID) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	if s == "" {
		return nil
	}
	_, err := uuid.FromString(s)
	if err != nil {
		return errors.New("invalid format uuid")
	}

	return nil
}

func (zu ZeroUUID) String() string {
	if zu == ZeroUUID((uuid.UUID{})) {
		return ""
	}
	return uuid.UUID(zu).String()
}

func (zu *ZeroUUID) Scan(src interface{}) error {
	switch src := src.(type) {
	case ZeroUUID: // support gorm convert from UUID to NullUUID
		*zu = src
		return nil

	case []byte:
		if len(src) == uuid.Size {
			return zu.UnmarshalBinary(src)
		}
		return zu.UnmarshalText(src)

	case string:
		return zu.UnmarshalText([]byte(src))
	}

	return fmt.Errorf("uuid: cannot convert %T to UUID", src)
}

func (zu ZeroUUID) Value() (driver.Value, error) {
	if zu == ZeroUUID(uuid.Nil) {
		return nil, nil
	}
	return zu.String(), nil
}

func (zu *ZeroUUID) UnmarshalBinary(data []byte) error {
	if len(data) != uuid.Size {
		return fmt.Errorf("uuid: UUID must be exactly 16 bytes long, got %d bytes", len(data))
	}
	copy(zu[:], data)

	return nil
}

func (zu ZeroUUID) MarshalBinary() ([]byte, error) {
	return uuid.UUID(zu).Bytes(), nil
}

func (zu *ZeroUUID) UnmarshalText(text []byte) error {
	switch len(text) {
	case 16:
		return zu.decodeBytes(text)
	case 32:
		return zu.decodeHashLike(text)
	case 36:
		return zu.decodeCanonical(text)
	default:
		return fmt.Errorf("uuid: incorrect UUID length %d in string %q", len(text), text)
	}
}

func (zu *ZeroUUID) decodeBytes(t []byte) error {
	if len(t) != 16 {
		return fmt.Errorf("uuid: incorrect length for raw byte format: %d", len(t))
	}
	copy(zu[:], t)
	return nil
}

// decodeHashLike decodes UUID strings that are using the following format:
//  "6ba7b8109dad11d180b400c04fd430c8".
func (u *ZeroUUID) decodeHashLike(t []byte) error {
	src := t[:]
	dst := u[:]

	_, err := hex.Decode(dst, src)
	return err
}

func (u *ZeroUUID) decodeCanonical(t []byte) error {
	if t[8] != '-' || t[13] != '-' || t[18] != '-' || t[23] != '-' {
		return fmt.Errorf("uuid: incorrect UUID format in string %q", t)
	}

	src := t
	dst := u[:]

	for i, byteGroup := range byteGroups {
		if i > 0 {
			src = src[1:] // skip dash
		}
		_, err := hex.Decode(dst[:byteGroup/2], src[:byteGroup])
		if err != nil {
			return err
		}
		src = src[byteGroup:]
		dst = dst[byteGroup/2:]
	}

	return nil
}

func ConvertToUUIDAndBinary(v interface{}) (*uuid.UUID, *bson.Binary) {
	var uid *uuid.UUID
	var key *bson.Binary
	if v != nil {
		if reflect.TypeOf(v).Kind() == reflect.String {
			if v.(string) != "" {
				u := uuid.FromStringOrNil(v.(string))
				uid = &u
				binary := SetUUIDBson(u)
				key = binary
			}
		} else if reflect.TypeOf(v).Kind() == reflect.TypeOf(uuid.UUID{}).Kind() {
			if v.(uuid.UUID) != (uuid.UUID{}) {
				u := v.(uuid.UUID)
				uid = &u
				binary := SetUUIDBson(u)
				key = binary
			}
		} else if reflect.ValueOf(v).Type() == reflect.ValueOf(bson.Binary{}).Type() {
			binary := v.(bson.Binary)
			key = &binary
			u := GetUUIDFromBson(&binary)
			uid = &u
		} else if reflect.TypeOf(v) == bsonUUIDType {
			id := uuid.FromBytesOrNil(v.(primitive.Binary).Data)
			uid = &id
		}
	}
	if uid != nil {
		if uid.String() == "00000000-0000-0000-0000-000000000000" {
			uid = nil
		}
	}
	return uid, key
}

func SetUUIDBson(id uuid.UUID) *bson.Binary {
	return &bson.Binary{
		Kind: uuidBinaryKind,
		Data: id.Bytes(),
	}
}

func GetUUIDFromBson(binary *bson.Binary) uuid.UUID {
	return uuid.FromBytesOrNil(binary.Data)
}

func ToUUIDBson(v interface{}) *bson.Binary {
	if v == nil {
		return nil
	}
	t := reflect.TypeOf(v).String()
	switch t {
	case "string":
		uid := uuid.FromStringOrNil(v.(string))
		return SetUUIDBson(uid)
	case "uuid.UUID":
		return SetUUIDBson(v.(uuid.UUID))
	case "*uuid.UUID":
		uid := v.(*uuid.UUID)
		return SetUUIDBson(*uid)
	}

	return nil
}

func GetBsonSlice(uuids []*uuid.UUID) []*bson.Binary {
	var bs = make([]*bson.Binary, 0)
	if uuids != nil {
		if len(uuids) > 0 {
			for _, uid := range uuids {
				key := SetUUIDBson(*uid)
				bs = append(bs, key)
			}
		}
	}
	return bs
}

func UUIDToSliceString(slice []*uuid.UUID) []string {
	var uids = make([]string, 0)
	if slice != nil {
		if len(slice) > 0 {
			for _, uid := range slice {
				if uid != nil {
					uids = append(uids, uid.String())
				}
			}
		}
	}
	return uids
}

func FindInSliceUUID(slice []*uuid.UUID, uid *uuid.UUID) (exists bool, index int) {
	if slice != nil {
		if len(slice) > 0 {
			for indexItem, item := range slice {
				if item.String() == uid.String() {
					exists = true
					index = indexItem
					break
				}
			}
		}
	}
	return
}