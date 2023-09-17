package qetag

import (
	"testing"
)

func TestNew(t *testing.T) {

	qetag := New()
	_, err := qetag.Write([]byte{1, 2, 3, 4, 5, 6, 7})
	if err != nil {
		t.Fatal(err)
	}

	if qetag.Etag() != "FowfKPwvSMJx1sSY8PJJzd42XFTF" {
		t.FailNow()
	}
}
