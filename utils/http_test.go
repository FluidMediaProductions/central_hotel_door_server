package utils

import (
	"testing"
	"net/http/httptest"
	"net/http"
	"fmt"
)

func TestGetJson(t *testing.T)  {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "{\"foo\": \"bar\"}")
	}))
	defer ts.Close()

	req, err := http.NewRequest("GET", ts.URL, nil)
	if err != nil {
		t.Fatalf("Failed to make request for testing GetJson: %v", err)
	}

	res, err := GetJson(req)
	if err != nil {
		t.Fatalf("Got error from GetJson: %v", err)
	}

	switch res["foo"].(type) {
	case string:
		if res["foo"].(string) != "bar" {
			t.Errorf("Expected foo: bar but got foo: %s", res["foo"].(string))
		}
		break
	default:
		t.Errorf("Expected foo of type string but got type %T", res["foo"])
	}
}