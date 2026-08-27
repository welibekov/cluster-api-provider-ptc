package util

import (
	"fmt"
	"reflect"
)

// MergeWithOverride merges fields from src to dst, overwriting dst fields if src fields are non-nil or non-empty.
// Both dst and src must be pointers to structs of the same type.
// Currently supports *bool and string fields; other field types are ignored.
// Unexported fields are skipped.
func MergeWithOverride(dst, src interface{}) error {
	dstRef := reflect.ValueOf(dst)
	srcRef := reflect.ValueOf(src)

	// Ensure inputs are pointers
	if dstRef.Kind() != reflect.Ptr || srcRef.Kind() != reflect.Ptr {
		return fmt.Errorf("dst and src must be pointers")
	}

	dstVal := dstRef.Elem()
	srcVal := srcRef.Elem()

	// Ensure we have structs
	if dstVal.Kind() != reflect.Struct || srcVal.Kind() != reflect.Struct {
		return fmt.Errorf("dst and src must point to structs")
	}

	// Ensure that dst and src are of the same type
	if dstVal.Type() != srcVal.Type() {
		return fmt.Errorf("dst and src must have the same type")
	}

	// Loop through all fields in the struct
	for i := 0; i < dstVal.NumField(); i++ {
		dstField := dstVal.Field(i)
		srcField := srcVal.Field(i)

		// Skip unexported fields or fields that can't be set
		if !dstField.CanSet() {
			continue
		}

		// Handle *bool fields
		if dstField.Kind() == reflect.Ptr && dstField.Type().Elem().Kind() == reflect.Bool {
			if !srcField.IsNil() {
				dstField.Set(srcField)
			}
		}

		// Handle string fields
		if dstField.Kind() == reflect.String {
			if srcField.String() != "" {
				dstField.SetString(srcField.String())
			}
		}
	}

	return nil
}
