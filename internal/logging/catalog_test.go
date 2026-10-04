// internal/logging/catalog_test.go

package logging

import (
	"reflect"
	"testing"
)

// TestEventCatalog_noZeroValueEntries ensures every declared event has a
// definition. An uninitialized EventDef still compiles and can be emitted,
// producing a log entry with empty event identity and severity.
func TestEventCatalog_noZeroValueEntries(t *testing.T) {
	walkEventDefs(t, reflect.ValueOf(Events), "Events")
}

func walkEventDefs(t *testing.T, v reflect.Value, path string) {
	t.Helper()

	if v.Kind() != reflect.Struct {
		return
	}

	if v.Type() == reflect.TypeOf(EventDef{}) {

		def := v.Interface().(EventDef)

		if def.Component == "" || def.Code == "" || def.Level == "" || def.Message == "" {
			t.Errorf(
				"%s is an incomplete EventDef (%+v) — every field in catalog.go's struct definitions must have a matching entry filled in under var Events",
				path, def,
			)
		}

		return
	}

	for i := 0; i < v.NumField(); i++ {
		walkEventDefs(t, v.Field(i), path+"."+v.Type().Field(i).Name)
	}
}
