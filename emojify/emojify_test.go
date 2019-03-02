package emojify

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var ts *httptest.Server

func setupEmojifyTests(delay time.Duration, timeout time.Duration) (Emojify, *httptest.ResponseRecorder) {
	ts = httptest.NewServer(
		http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			resp := `
			{
			  "Success": true,
				"Error": "",
				"Faces": [
					{}
				]
			}
			`
			time.Sleep(delay)

			rw.Write([]byte(resp))
		}),
	)

	e := NewEmojify(nil, strings.ReplaceAll(ts.URL, "http://", ""), "/", 1, timeout)

	return e, httptest.NewRecorder()
}

func TestGetFacesReturnsFaces(t *testing.T) {
	e, _ := setupEmojifyTests(0*time.Millisecond, 20*time.Millisecond)
	resp := make(chan FindFaceResponse, 1) // make a buffered channel so that we don't block
	f := bytes.NewReader([]byte("dfdfdf"))
	to := time.After(1000 * time.Millisecond)

	e.GetFaces(f, resp)

	select {
	case r := <-resp:
		assert.Nil(t, r.Error)
	case <-to:
		t.Fatal("timeout waiting for response")
	}
}

func TestGetFacesMultipleCalls(t *testing.T) {
	e, _ := setupEmojifyTests(0*time.Millisecond, 20*time.Millisecond)
	resp1 := make(chan FindFaceResponse, 1) // make a buffered channel so that we don't block
	resp2 := make(chan FindFaceResponse, 1) // make a buffered channel so that we don't block
	f1 := bytes.NewReader([]byte("dfdfdf"))
	f2 := bytes.NewReader([]byte("dfdfdf"))
	to := time.After(1000 * time.Millisecond)

	e.GetFaces(f1, resp1)
	e.GetFaces(f2, resp2)

	select {
	case r := <-resp1:
		assert.Nil(t, r.Error)
	case <-to:
		t.Fatal("timeout waiting for response")
	}

	select {
	case r := <-resp2:
		assert.Nil(t, r.Error)
	case <-to:
		t.Fatal("timeout waiting for response")
	}
}

func TestGetFacesTimesOutWaitingForWorker(t *testing.T) {
	e, _ := setupEmojifyTests(1000*time.Millisecond, 50*time.Millisecond)
	resp1 := make(chan FindFaceResponse, 1) // make a buffered channel so that we don't block
	resp2 := make(chan FindFaceResponse, 1)
	f1 := bytes.NewReader([]byte("dfdfdf"))
	f2 := bytes.NewReader([]byte("dfdfdf"))
	to := time.After(1000 * time.Millisecond)

	e.GetFaces(f1, resp1)
	time.Sleep(10 * time.Millisecond) // ensure the second call happens after the first
	e.GetFaces(f2, resp2)

	select {
	case r := <-resp2:
		assert.NotNil(t, r.Error, "Expected timeout error")
	case <-to:
		t.Fatal("timeout waiting for response")
	}
}
