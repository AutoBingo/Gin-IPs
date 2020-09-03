package test

import "testing"

func TestAPI_Request(t *testing.T) {
	url := "127.0.0.1:8080"
	ak := "A00001"
	sk := "SECRET-A00001"
	api := NewAPI(url, ak, sk)
	params := map[string]interface{}{
		"ip": "10.1.162.18",
	}
	if result, err  := api.Request("/", "GET", params); err != nil {
		t.Fatal(err)
	} else {
		t.Log(result)
	}
}
