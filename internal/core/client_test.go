package core

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_State(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(`{"ok":true,"volume":12,"current_app_id":"netflix"}`))
	}))
	defer srv.Close()

	got, err := New(srv.URL).State(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if got["volume"].(float64) != 12 {
		t.Errorf("got volume=%v", got["volume"])
	}
}

func TestClient_ErrorStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(502)
		w.Write([]byte(`{"detail":"TV unreachable"}`))
	}))
	defer srv.Close()

	if err := New(srv.URL).PowerOff(context.Background()); err == nil {
		t.Fatal("expected error")
	}
}
