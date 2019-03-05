package id

import (
	"testing"
)

func TestGetTimeFromUuid(t *testing.T) {
	if GetTimeFromUuid(4176607997370268) != 1500359236 {
		t.Error("GetTimeFromUuid Error")
	}
}
