package orm

import "strings"

type foreignKey struct {
	fkField1 []string
	fkField2 []string
}

func newForeignKeyFromTag(tag string) foreignKey {
	vals := strings.Split(tag, fieldSeperate)
	fkField1 := []string{}
	fkField2 := []string{}
	var fkKey1 = "fk_field1"
	var fkKey2 = "fk_field2"
	var getFkField = func(fkVal string) []string {
		data := strings.Split(fkVal, ":")

		return strings.Split(data[1], fieldFKSeperate)
	}
	for _, val := range vals {
		if strings.Contains(val, fkKey1) {
			fkField1 = getFkField(val)
		}
		if strings.Contains(val, fkKey2) {
			fkField2 = getFkField(val)
		}
	}
	return foreignKey{
		fkField1: fkField1,
		fkField2: fkField2,
	}
}

func (f foreignKey) Validate() error {
	if len(f.fkField1) == 0 || len(f.fkField2) == 0 {
		return ErrNotIdentifyFkField
	}
	return nil
}
