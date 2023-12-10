package orm

import (
	"database/sql"
	"fmt"
	"log"
	"reflect"
	"strings"
	"sync"

	"github.com/fatih/structs"
)

var TAGNAME = "db"
var TABLE_FIELD_NAME = "TableName"
var TAG_PK = "pk"
var TAG_FK = "fk"
var TAG_TYPE = "type"
var PAGINATE_COLUMN_NAME = "total_row"
var fieldSeperate = ","
var fieldFKSeperate = "+"
var fieldJoinKeyMap = "+"

func isNil(val interface{}) bool {
	if val == nil || (reflect.ValueOf(val).Kind() == reflect.Ptr && reflect.ValueOf(val).IsNil()) {
		return true
	}
	return false
}

func mustbeStruct(data interface{}) error {
	if reflect.TypeOf(data).Kind() == reflect.Struct {
		return nil
	}
	return ErrMustBeStruct
}

func getEmptySlice(structType reflect.Type, cap int) reflect.Value {
	sliceType := reflect.SliceOf(structType)

	emptySlice := reflect.MakeSlice(sliceType, 0, cap)
	return emptySlice
}

func copy(src reflect.Value) reflect.Value {
	srcValue := src.Elem()
	destPtrType := reflect.PtrTo(srcValue.Type())
	destPtr := reflect.New(destPtrType.Elem())
	destPtr.Elem().Set(srcValue)
	return destPtr
}

func getTableName(faith *structs.Struct) string {
	var tablename string

	if f, ok := faith.FieldOk(TABLE_FIELD_NAME); ok {
		tablename = f.Tag(TAGNAME)
	}

	return tablename
}

func getTagValue(faith *structs.Struct, field string, tag string) string {
	var tagVal string

	if f, ok := faith.FieldOk(field); ok {
		tagVal = strings.TrimSpace(f.Tag(tag))
	}

	return tagVal
}

func getFieldValue(faith *structs.Struct, field string) (interface{}, error) {
	f, ok := faith.FieldOk(field)
	if !ok {
		log.Println("field ", field, "not found")
		return reflect.New(nil), ErrFieldNotFound
	}

	return f.Value(), nil
}

func getFieldValues(faith *structs.Struct, fields []string) ([]interface{}, error) {
	var vals = make([]interface{}, 0)
	for _, field := range fields {
		v, err := getFieldValue(faith, field)
		if err != nil {
			return nil, err
		}
		vals = append(vals, v)
	}
	return vals, nil
}

func getFieldMetaData(faith *structs.Struct, option MapperOption) (pkFields []string, fkFields []string) {
	pkValField := getTagValue(faith, TABLE_FIELD_NAME, TAG_PK)
	pkFields = strings.Split(pkValField, fieldSeperate)
	fkFields = make([]string, 0)

	if len(option.pkFields) > 0 {
		for _, pkField := range option.pkFields {
			if pkField.faith.Name() == faith.Name() {
				pkFields = pkField.fieldName
				break
			}
		}
	}

	for _, field := range faith.Fields() {
		if field.IsEmbedded() {
			continue
		}
		if field.Tag(TAG_FK) != "" {
			fkFields = append(fkFields, field.Name())
		}
	}

	return pkFields, fkFields
}

func getStructFields(faith *structs.Struct) *sync.Map {
	var ptrColumnMap = new(sync.Map)
	fields := faith.Fields()
	for _, f := range fields {
		tagCol := f.Tag(TAGNAME)
		if tagCol != "" && tagCol != "-" {
			ptrColumnMap.Store(tagCol, f)
		}
	}

	return ptrColumnMap
}

func setFieldFromType(field *structs.Field, data interface{}) error {
	var tag = field.Tag(TAG_TYPE)
	if tag != "" && tag != "-" {
		registry, ok := GlobalRegistry[tag]
		if !ok {
			return fmt.Errorf("error: %s %s", tag, ErrRegistryNotFound.Error())
		}
		if err := registry.Bind(field, data); err != nil {
			return err
		}
	}

	return nil
}

func fillValue(ptr interface{}, columns []*sql.ColumnType, values []interface{}) error {
	faith := structs.New(ptr)
	schTableName := getTableName(faith)

	ptrColumnMap := getStructFields(faith)

	if len(values) > 0 {
		for index, col := range columns {
			orderCol := strings.ReplaceAll(col.Name(), schTableName+".", "")
			if field, ok := ptrColumnMap.Load(orderCol); ok {
				if err := setFieldFromType(field.(*structs.Field), values[index]); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func transfromIdString(faith *structs.Struct, field string, tag string, val interface{}) (string, error) {
	typeVal := getTagValue(faith, field, tag)
	registry, ok := GlobalRegistry[typeVal]
	if !ok {
		return "", ErrRegistryNotFound
	}

	return registry.RegisterPkId(val), nil
}

func getIds(faith *structs.Struct, fields []string) (string, error) {
	var ids = []string{}
	for _, field := range fields {
		val, err := getFieldValue(faith, field)
		if err != nil {
			return "", err
		}

		id, err := transfromIdString(faith, field, TAG_TYPE, val)
		if err != nil {
			return "", err
		}
		if id == "" {
			return "", nil
		}
		ids = append(ids, id)
	}
	return strings.Join(ids, "+"), nil
}

/*
Equal Value if a same type
*/
func equal(tagTypeVal string, x interface{}, y interface{}) bool {
	if !isNil(x) && !isNil(y) {
		registry := GlobalRegistry[tagTypeVal]
		return registry.Equal(x, y)
	}
	return false
}
