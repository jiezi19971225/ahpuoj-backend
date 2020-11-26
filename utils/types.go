package utils

import (
	"encoding/json"
	"fmt"
	"gopkg.in/guregu/null.v4"
	"time"
)

type JSONDate time.Time
type JSONDateTime time.Time
type RelativeNullString null.String

func (t JSONDate) MarshalJSON() ([]byte, error) {
	stamp := fmt.Sprintf("\"%s\"", time.Time(t).Format("2006-01-02"))
	return []byte(stamp), nil
}

func (t JSONDateTime) MarshalJSON() ([]byte, error) {
	stamp := fmt.Sprintf("\"%s\"", time.Time(t).Format("2006-01-02 15:04:05"))
	return []byte(stamp), nil
}

func (rns RelativeNullString) MarshalJSON() ([]byte, error) {
	ans, _ := json.Marshal(ConvertTextImgUrl(rns.String))
	return []byte(ans), nil
}
