package utils

import (
	"reflect"
	"strings"
)

func IsStructEmpty(s interface{}) bool {
    val := reflect.ValueOf(s)
    for i := 0; i < val.NumField(); i++ {
        field := val.Field(i)
        // 文字列の場合、空白のみかどうかもチェック
        if field.Kind() == reflect.String {
            if strings.TrimSpace(field.String()) != "" {
                return false // 文字列が空白のみでない場合
            }
        } else if !field.IsZero() {
            return false // ゼロ値でない場合
        }
    }
    return true // 全てのフィールドがゼロ値または空白のみなら、trueを返す
}