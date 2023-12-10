package helper

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"

	"github.com/gofrs/uuid"
)

func ValidateKeyExists(keys []string, params map[string]interface{}) map[string]error {
	if len(keys) == 0 || params == nil {
		return nil
	}
	var errs = make(map[string]error, 0)

	for _, key := range keys {
		if _, ok := params[key]; !ok {
			message := fmt.Sprintf("key '%s' not exists", key)
			errs[key] = errors.New(message)
		}
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

/*
	Validate Rule
*/

func ValidateNotSpace(val interface{}) error {
	if err := ValidateTypeString(val); err != nil {
		return err
	}

	if val.(string) == " " {
		return errors.New(fmt.Sprintf(`can not be space only`))
	}

	return nil
}

func ValidateOnlyThaiLetterNumeric(val interface{}) error {
	if err := ValidateTypeString(val); err != nil {
		return err
	}

	thaiRegex := `[\x{0E00}-\x{0E7Fa}0-9]`
	engRegex := `[a-zA-Z]`
	exp := regexp.MustCompile(thaiRegex)
	enExp := regexp.MustCompile(engRegex)

	if !exp.MatchString(val.(string)) || enExp.MatchString(val.(string)) {
		return errors.New("letter can be Thai letters and digits only")
	}

	return nil
}

func ValidateUUIDOrIDZero(val interface{}) error {
	if err := ValidateTypeString(val); err != nil {
		return err
	}

	if val.(string) == "0" {
		return nil
	}

	if _, err := uuid.FromString(val.(string)); err != nil {
		return errors.New("is not uuid")
	}

	return nil
}

func ValidateTypeUUID(val interface{}) error {
	if err := ValidateTypeString(val); err != nil {
		return err
	}

	if _, err := uuid.FromString(val.(string)); err != nil {
		return errors.New("is not uuid")
	}

	return nil
}

func ValidateTypeString(val interface{}) error {
	rf := reflect.ValueOf(val)
	if rf.Kind() != reflect.String {
		return errors.New(fmt.Sprintf("is not type string"))
	}
	return nil
}

func ValidateTypeInt(val interface{}) error {
	rf := reflect.ValueOf(val)
	if rf.Kind() == reflect.Float64 {
		stringValue := fmt.Sprintf("%d", int(val.(float64)))
		_, err := strconv.ParseInt(stringValue, 10, 32)
		if err == nil {
			return nil
		}
	}
	if rf.Kind() != reflect.Int {
		return errors.New(fmt.Sprintf("is not type int"))
	}
	return nil
}

func ValidateTypeFloat(val interface{}) error {
	rf := reflect.ValueOf(val)
	if rf.Kind() != reflect.Float64 {
		return errors.New(fmt.Sprintf("is not type float"))
	}
	return nil
}

func ValidateTypeMap(val interface{}) error {
	rf := reflect.ValueOf(val)
	if rf.Kind() != reflect.Map {
		return errors.New(fmt.Sprintf("is not type map"))
	}
	return nil
}

func ValidateTypeSlice(val interface{}) error {
	rf := reflect.ValueOf(val)
	if rf.Kind() != reflect.Slice {
		return errors.New(fmt.Sprintf("is not type array"))
	}
	return nil
}

func ValidateTypeBool(v interface{}) error {
	if reflect.TypeOf(v).Kind() == reflect.Bool {
		return nil
	}
	return errors.New("is not type bool")
}

func ValidateTypeBoolString(v interface{}) error {
	if reflect.TypeOf(v).Kind() == reflect.Bool {
		return nil
	}
	if reflect.TypeOf(v).Kind() == reflect.String {
		if _, err := strconv.ParseBool(v.(string)); err == nil {
			return nil
		}
	}
	return errors.New("is not type bool")
}

/* validate function with null */

func ValidateTypeMapWithNull(val interface{}) error {
	if val == nil {
		return nil
	}
	rf := reflect.ValueOf(val)
	if rf.Kind() != reflect.Map {
		return errors.New(fmt.Sprintf("value must be null or type map"))
	}
	return nil
}

func reverseString(rawString string) string {
	if rawString == "" {
		return rawString
	}
	var reverseString string

	for i := len([]rune(rawString)); i > 0; i-- {
		var str = rawString[i-1]
		reverseString += string(str)
	}

	return reverseString
}

func ValidCitizenId(citizen string) bool {
	if len(citizen) != 13 {
		return false
	}

	var revString = reverseString(citizen)
	var total float64
	for index := 1; index < 13; index++ {
		var mul = index + 1
		var num, _ = strconv.Atoi(string([]rune(revString)[index]))
		var count = num * mul
		total = total + float64(count)
	}
	var mod = int(total) % 11
	var sub = 11 - mod
	var checkDigit = sub % 10

	var lastCitizen, _ = strconv.Atoi(string([]rune(revString)[0]))
	if lastCitizen == checkDigit {
		return true
	}

	return false
}

func IsCompany(citizen string) bool {
	if !ValidCitizenId(citizen) {
		return false
	}

	var num, err = strconv.Atoi(string([]rune(citizen)[0]))
	if err != nil {
		return false
	}
	if num == 0 {
		return true
	}

	return false
}