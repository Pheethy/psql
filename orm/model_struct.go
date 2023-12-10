package orm

import (
	"reflect"
	"strings"
	"sync"

	"github.com/fatih/structs"
)

type modelStruct struct {
	fieldname        string        // fieldname ใช้สำหรับ reference ของ pk
	name             string        // ชื่อ Type model
	model            interface{}   // ค่าของ model เช่น new(models.User)
	modelType        reflect.Type  // reflect.TypeOf(new(models.User))
	modelSlice       reflect.Value // value reflect.Value( []*models.User)
	pkM              *sync.Map     // map สำหรับทำ duplicate PK
	pkFields         []string
	refFields        []string // binding modelRef -> main
	isReferenceModel bool
	subRefModel      []modelStruct
}

type modelStructs []modelStruct

func newMainModelStruct(model interface{}, option MapperOption) modelStruct {
	var modelType = reflect.TypeOf(model)
	var ptrs = getEmptySlice(modelType, 10000)
	ms := modelStruct{
		name:             modelType.String(),
		model:            model,
		modelType:        modelType,
		modelSlice:       ptrs,
		pkM:              new(sync.Map),
		isReferenceModel: false,
	}
	if ms.IsMainModel() {
		faith := structs.New(model)
		pkFields, fkFields := getFieldMetaData(faith, option)

		ms.pkFields = pkFields
		ms.refFields = fkFields
	}

	return ms
}
func newRefModelStruct(model interface{}, fieldName string, refFields []string) modelStruct {
	var modelType = reflect.TypeOf(model)
	var ptrs = getEmptySlice(modelType, 0)
	ms := modelStruct{
		fieldname:        fieldName,
		name:             modelType.String(),
		model:            model,
		modelType:        modelType,
		modelSlice:       ptrs,
		pkM:              new(sync.Map),
		isReferenceModel: true,
		refFields:        refFields,
		subRefModel:      make([]modelStruct, 0),
	}

	return ms
}

func newModelStruct(model interface{}, options MapperOption) ([]modelStruct, error) {
	if err := validateModel(model); err != nil {
		return nil, err
	}
	faithModel := structs.New(model)
	_, fkFields := getFieldMetaData(faithModel, options)
	mainModelStruct := newMainModelStruct(model, options)

	var ms = make([]modelStruct, 0)
	ms = append(ms, mainModelStruct)

	/* add fk model */
	if len(fkFields) > 0 && options.autobinding {
		for _, field := range fkFields {
			val, err := getFieldValue(faithModel, field)
			if err != nil {
				return nil, err
			}
			var elem reflect.Value

			if reflect.ValueOf(val).Kind() == reflect.Ptr {
				/* pointer Object */
				types := reflect.ValueOf(val).Type()
				elem = reflect.New(types.Elem())
			} else {
				/* slice */
				elemType := reflect.TypeOf(val).Elem()
				elem = reflect.New(elemType.Elem())
			}

			if err := validateModel(elem.Interface()); err != nil {
				return nil, err
			}

			if tagVal := getTagValue(faithModel, field, TAG_FK); tagVal != "" {
				fk := newForeignKeyFromTag(tagVal)
				if err := fk.Validate(); err != nil {
					return nil, err
				}

				ms = append(ms, newRefModelStruct(elem.Interface(), field, fk.fkField2))
			}
		}
	}

	return ms, nil
}

func (m modelStruct) IsZero() bool {
	return m.name == ""
}

func (m modelStruct) IsMainModel() bool {
	return !m.isReferenceModel
}

func (m modelStructs) GetMainModel() modelStruct {
	for index := range m {
		if m[index].IsMainModel() {
			return m[index]
		}
	}
	return modelStruct{}
}

func (m modelStructs) GetListReferenceModelIndex() []int {
	var ms = make([]int, 0)
	for index := range m {
		if !m[index].IsMainModel() {
			ms = append(ms, index)
		}
	}
	return ms
}

func (m modelStructs) GetRefModelByFieldName(fieldName string) modelStruct {
	if len(m) > 0 {
		for index := range m {
			if strings.EqualFold(m[index].fieldname, fieldName) {
				return m[index]
			}
		}
	}
	return modelStruct{}
}
