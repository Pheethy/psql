package helper

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"4d63.com/tz"
)

const (
	TimestampLayout = "2006-01-02 15:04:05" // 2006-01-02 15:04:05, 2006-01-02T15:04:05
)

type Timestamp time.Time

/*
------------------------
Timestamp Function
------------------------
*/

func NewTimestampFromString(dateString string) Timestamp {
	if dateString == "" {
		return Timestamp(time.Time{})
	}
	loc, _ := tz.LoadLocation("Asia/Bangkok")
	d, err := time.ParseInLocation(TimestampLayout, dateString, loc)
	if err != nil {
		panic(err)
	}
	return Timestamp(d)
}

func NewTimestampFromTime(t time.Time) Timestamp {
	loc := time.FixedZone("UTC+7", 7*60*60)
	d, err := time.Parse(TimestampLayout, t.UTC().Format(TimestampLayout))
	if err != nil {
		panic(err)
	}
	d = d.In(loc)
	return Timestamp(d)
}

func (t Timestamp) ToUnix() int64 {
	tt := time.Time(t)
	return tt.Unix()
}

func (j *Timestamp) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	t, err := time.Parse(TimestampLayout, s)
	if err != nil {
		return err
	}
	*j = Timestamp(t)
	return nil
}

func (j Timestamp) MarshalJSON() ([]byte, error) {
	return json.Marshal(j.Format(TimestampLayout))
}

func (j Timestamp) Format(s string) string {
	t := time.Time(j)
	return t.Format(s)
}

func (j Timestamp) YearDay() int {
	t := time.Time(j)
	return t.YearDay()
}

func (j Timestamp) String() string {
	return j.Format(TimestampLayout)
}

func (j *Timestamp) Interface() interface{} {
	if j == nil {
		return nil
	}

	return j.Format(TimestampLayout)
}

func (j *Timestamp) GetBSON() (interface{}, error) {
	if j == nil {
		return nil, nil
	}
	if (*j) == Timestamp(time.Time{}) {
		return nil, nil
	}
	loc, _ := tz.LoadLocation("Asia/Bangkok")
	t := time.Time(*j)
	d, err := time.ParseInLocation(TimestampLayout, t.Format(TimestampLayout), loc)
	if err != nil {
		return nil, err
	}

	return d, nil
}

func (j Timestamp) Value() (driver.Value, error) {
	if j == (Timestamp{}) {
		return nil, nil
	}
	// Delegate to UUID Value function
	return j.String(), nil
}

func (j Timestamp) ValueOrZero() string {
	if j == (Timestamp{}) {
		return ""
	}
	return j.String()
}

func (j Timestamp) ToTime() time.Time {
	return time.Time(j)
}

func (t *Timestamp) Scan(value interface{}) error {
	if value == nil {
		*t = Timestamp(time.Time{})
		return nil
	}
	switch v := value.(type) {
	case time.Time:
		*t = Timestamp(v)
		return nil
	case string:
		loc, _ := tz.LoadLocation("Asia/Bangkok")
		parsed, err := time.ParseInLocation(TimestampLayout, v, loc)
		if err != nil {
			return err
		}
		*t = Timestamp(parsed)
		return nil
	case []byte:
		loc, _ := tz.LoadLocation("Asia/Bangkok")
		parsed, err := time.ParseInLocation(TimestampLayout, string(v), loc)
		if err != nil {
			return err
		}
		*t = Timestamp(parsed)
		return nil
	default:
		return fmt.Errorf("cannot scan type %T into Timestamp", value)
	}
}

