package server

import (
	"testing"
)

type some struct {
	Id   int64
	Data string
}

func TestNew(t *testing.T) {
	srv, err := New[some]("test")
	if err != nil {
		t.Fatal(err.Error())
	}

	s := srv.GetFasthttpServer()
	s.ListenAndServe(":80")
}
