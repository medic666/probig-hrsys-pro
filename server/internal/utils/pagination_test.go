package utils

import (
	"testing"
)

func TestPageRequest_Offset(t *testing.T) {
	p := PageRequest{PageNum: 1, PageSize: 20}
	if p.Offset() != 0 {
		t.Errorf("offset should be 0, got %d", p.Offset())
	}

	p.PageNum = 2
	if p.Offset() != 20 {
		t.Errorf("offset should be 20, got %d", p.Offset())
	}
}
