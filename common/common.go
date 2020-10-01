package common

import (
	"errors"
	"strconv"
)

var ErrNameIsExisted = errors.New("昵称已存在")

const (
	None uint32 = iota
	Join
	Aite
	Quit
)

func InterfaceSlice2String(params []interface{}, comma string) string {
	var str string
	for _, param := range params {
		switch v := param.(type) {
		case int:
			str += strconv.Itoa(v) + comma
		case int64:
			str += strconv.FormatInt(v, 10) + comma
		case string:
			str += v + comma
		case float32:
			str += strconv.FormatFloat(float64(v), 'f', 6, 64) + comma
		case float64:
			str += strconv.FormatFloat(v, 'f', 6, 64) + comma
		}
	}
	if len(str) > 1 {
		return str[:len(str)-1]
	} else {
		return str
	}
}

func StrSlice2Str(params []string, comma string) string {
	var str string
	for _, param := range params {
		str += param + comma
	}
	if len(str) > 1 {
		return str[:len(str)-1]
	} else {
		return str
	}
}
