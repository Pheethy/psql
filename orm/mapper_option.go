package orm

import (
	"fmt"

	"github.com/fatih/structs"
)

type IterationTypes string

const (
	ITERATION_TYPE_LIST     IterationTypes = "array_list"
	ITERATION_TYPE_HASH_MAP IterationTypes = "hash_map"
)

type MapperOption struct {
	autobinding       bool
	pkFields          []MapperOptionPkField
	copyIntoIteration bool
	iterTypes         IterationTypes
	pkIterMapKeys     []string // pk for map column with rows.Next()
	mapStoreColumn    []string // choosestoreColumns
}

type MapperOptionPkField struct {
	model     interface{}
	fieldName []string
	faith     *structs.Struct
}

func NewMapperOption() MapperOption {
	return MapperOption{
		autobinding:    true,
		pkFields:       make([]MapperOptionPkField, 0),
		pkIterMapKeys:  make([]string, 0),
		mapStoreColumn: make([]string, 0),
	}
}

func NewMapperOptionPKField(model interface{}, fieldname []string) MapperOptionPkField {
	return MapperOptionPkField{
		model:     model,
		fieldName: fieldname,
		faith:     structs.New(model),
	}
}

func (m MapperOption) SetDisableBinding() MapperOption {
	m.autobinding = false
	return m
}

func (m MapperOption) SetOverridePKField(fields ...MapperOptionPkField) MapperOption {
	m.pkFields = fields
	return m
}

// Copy rows.Next() -> ArrayList
func (m MapperOption) SetIterationList() MapperOption {
	m.copyIntoIteration = true
	m.iterTypes = ITERATION_TYPE_LIST
	return m
}

// pkColumnName is column from query to get pk if data is null will be skip that rows |
// storeOnlyColumn is optional choose store column
// example []string{"orders.id"}, nil
func (m MapperOption) SetIterationHashMapWithColumns(pkColumnName []string, storeColumns []string) MapperOption {
	m.copyIntoIteration = true
	m.iterTypes = ITERATION_TYPE_HASH_MAP
	m.pkIterMapKeys = pkColumnName
	if len(storeColumns) > 0 {
		m.mapStoreColumn = storeColumns
	}
	return m
}

/*
model is column from query to get pk if data is null will be skip that rows |
storeOnlyColumn is optional choose store column
example new(Order), nil
*/
func (m MapperOption) SetIterationHashMapWithModel(model interface{}, storeColumns []string) MapperOption {
	faith := structs.New(model)
	tablename := getTableName(faith)
	pkFields, _ := getFieldMetaData(faith, m)
	pkColumnNames := []string{}
	for index := range pkFields {
		faith.Field(pkFields[index])
		tagVal := getTagValue(faith, pkFields[index], TAGNAME)
		if tagVal != "" && tagVal != "-" {
			pkColumnNames = append(pkColumnNames, fmt.Sprintf("%s.%s", tablename, tagVal))
		}
	}

	m.copyIntoIteration = true
	m.iterTypes = ITERATION_TYPE_HASH_MAP
	m.pkIterMapKeys = pkColumnNames
	if len(storeColumns) > 0 {
		m.mapStoreColumn = storeColumns
	}
	return m
}
