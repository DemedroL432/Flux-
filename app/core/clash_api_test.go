package core

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchClashConnections_BearerMatchesSecret(t *testing.T) {
	wantSecret := "test-secret-token"
	var sawAuth string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/connections" {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
		sawAuth = r.Header.Get("Authorization")
		if sawAuth != "Bearer "+wantSecret {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = io.WriteString(w, `Unauthorized`)
			return
		}
		_, _ = io.WriteString(w, `{"connections":[]}`)
	}))
	defer ts.Close()

	prev := clashAPIHTTPBase
	clashAPIHTTPBase = ts.URL
	defer func() { clashAPIHTTPBase = prev }()

	rows, err := FetchClashConnections(wantSecret)
	if err != nil {
		t.Fatal(err)
	}
	if sawAuth != "Bearer "+wantSecret {
		t.Fatalf("Authorization header: got %q", sawAuth)
	}
	if rows == nil {
		t.Fatal("expected non-nil slice")
	}
}

func TestPutClashProxyGroupSelection_PUTProxiesGlobal(t *testing.T) {
	wantSecret := "put-secret"
	var sawAuth string
	var sawBody string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/proxies/GLOBAL" || r.Method != http.MethodPut {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		sawAuth = r.Header.Get("Authorization")
		if sawAuth != "Bearer "+wantSecret {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		b, _ := io.ReadAll(io.LimitReader(r.Body, 512))
		sawBody = string(b)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	prev := clashAPIHTTPBase
	clashAPIHTTPBase = ts.URL
	defer func() { clashAPIHTTPBase = prev }()

	if err := PutClashProxyGroupSelection(wantSecret, "GLOBAL", "proxy"); err != nil {
		t.Fatal(err)
	}
	if sawBody != `{"name":"proxy"}` {
		t.Fatalf("body: got %q", sawBody)
	}
}
