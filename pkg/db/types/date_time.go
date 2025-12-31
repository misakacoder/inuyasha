package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type DateTime time.Time

func (dateTime *DateTime) Scan(v any) error {
	newTime, ok := v.(time.Time)
	if !ok {
		return fmt.Errorf("不能将%T转换为DateTime", v)
	}
	*dateTime = DateTime(newTime)
	return nil
}

func (dateTime *DateTime) Value() (driver.Value, error) {
	if dateTime != nil {
		return dateTime.Time(), nil
	}
	return nil, nil
}

func (dateTime *DateTime) MarshalJSON() ([]byte, error) {
	if dateTime != nil {
		return json.Marshal(dateTime.Time().Format(time.DateTime))
	}
	return nil, nil
}

func (dateTime *DateTime) UnmarshalJSON(data []byte) error {
	var timeString string
	err := json.Unmarshal(data, &timeString)
	if err != nil {
		return err
	}
	newTime, err := time.ParseInLocation(time.DateTime, timeString, time.Local)
	if err != nil {
		return err
	}
	*dateTime = DateTime(newTime)
	return err
}

func (dateTime *DateTime) Time() time.Time {
	if dateTime != nil {
		return time.Time(*dateTime)
	}
	return time.Time{}
}

func (dateTime *DateTime) String() string {
	if dateTime != nil {
		return dateTime.Time().Format(time.DateTime)
	}
	return ""
}

func DateTimeNow() *DateTime {
	return DateTimeFrom(time.Now())
}

func DateTimeFrom(tm time.Time) *DateTime {
	dateTime := DateTime(tm)
	return &dateTime
}
