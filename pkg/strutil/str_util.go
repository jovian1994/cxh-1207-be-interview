package strutil

import (
	"encoding/base64"
	"fmt"
	"reflect"
	"strconv"
	"unicode"
	"unsafe"
)

func IncludeLetter(str string) bool {
	runes := []rune(str)
	for _, r := range runes {
		if unicode.IsLetter(r) && !unicode.Is(unicode.Scripts["Han"], r) {
			fmt.Println("r", r)
			return true
		}
	}
	return false
}

func IsDigit(str string) bool {
	for _, x := range []rune(str) {
		if !unicode.IsDigit(x) {
			return false
		}
	}
	return true
}

func Int64ToString(num int64) string {
	return strconv.FormatInt(num, 10)
}

func BytesToString(data *[]byte) string {
	return *(*string)(unsafe.Pointer(data))
}

func StringToBytes(data string) (b []byte) {
	*(*string)(unsafe.Pointer(&b)) = data
	(*reflect.SliceHeader)(unsafe.Pointer(&b)).Cap = len(data)
	return
}

func BytesForJson(data []byte) []byte {
	base64Data := base64.StdEncoding.EncodeToString(data)
	return StringToBytes(base64Data)
}

func Base64stringToBytes(data string) ([]byte, error) {
	decodeData, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		fmt.Println("Base64 decoding error:", err)
		return make([]byte, 0), err
	}
	return decodeData, nil
}
