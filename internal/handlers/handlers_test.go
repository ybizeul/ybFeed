package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

const baseDir = "../../test/"
const dataDir = "./data"

func TestIndexHtml(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	RootHandlerFunc(w, req)
	res := w.Result()
	if res.StatusCode != 200 {
		t.Errorf("Expect code 200 but got %d", res.StatusCode)
	}
}

func TestCreateFeed(t *testing.T) {
	const feedName = "6ff1146b683086b8f59f550189a8f91f"

	t.Cleanup(func() {
		os.RemoveAll(path.Join(baseDir, dataDir, feedName))
	})
	api := NewApiHandler(path.Join(baseDir, dataDir))

	req := httptest.NewRequest(http.MethodGet, "/api/feed/"+feedName, nil)

	w := httptest.NewRecorder()

	api.ApiHandleFunc(w, req)

	res := w.Result()

	if res.StatusCode != 200 {
		t.Errorf("Expect code 200 but got %d", res.StatusCode)
	}

	// Read cookie
	cookies := w.Result().Cookies()
	found := false
	fmt.Println(cookies)
	for _, c := range cookies {
		found = (c.Name == "Secret")
	}

	if found == false {
		t.Errorf("Cookie is not present in reply")
	}
}

func TestGetFeedNoCredentials(t *testing.T) {
	const feedName = "test"
	api := NewApiHandler(path.Join(baseDir, dataDir))

	req := httptest.NewRequest(http.MethodGet, "/api/feed/"+feedName, nil)

	w := httptest.NewRecorder()

	api.ApiHandleFunc(w, req)

	res := w.Result()

	if res.StatusCode != 401 {
		spew.Dump(res)
		t.Errorf("Expect code 401 but got %d", res.StatusCode)
	}
}

func TestGetFeedBadCookie(t *testing.T) {
	const feedName = "test"
	const secret = "foo"
	api := NewApiHandler(path.Join(baseDir, dataDir))

	req := httptest.NewRequest(http.MethodGet, "/api/feed/"+feedName, nil)
	req.AddCookie(&http.Cookie{Name: "Secret", Value: secret})
	w := httptest.NewRecorder()

	api.ApiHandleFunc(w, req)

	res := w.Result()

	if res.StatusCode != 401 {
		t.Errorf("Expect code 401 but got %d", res.StatusCode)
	}
}

func TestGetFeedCookieAuth(t *testing.T) {
	const feedName = "test"
	const secret = "b90e516e-b256-41ff-a84e-a9e8d5b6fe30"
	api := NewApiHandler(path.Join(baseDir, dataDir))

	req := httptest.NewRequest(http.MethodGet, "/api/feed/"+feedName, nil)
	req.AddCookie(&http.Cookie{Name: "Secret", Value: secret})
	w := httptest.NewRecorder()

	api.ApiHandleFunc(w, req)

	res := w.Result()

	if res.StatusCode != 200 {
		t.Errorf("Expect code 200 but got %d", res.StatusCode)
	}
}

func TestGetFeedBadURLSecret(t *testing.T) {
	const feedName = "test"
	const secret = "foo"
	api := NewApiHandler(path.Join(baseDir, dataDir))

	req := httptest.NewRequest(http.MethodGet, "/api/feed/"+feedName+"?secret=foo", nil)
	w := httptest.NewRecorder()

	api.ApiHandleFunc(w, req)

	res := w.Result()

	if res.StatusCode != 401 {
		t.Errorf("Expect code 401 but got %d", res.StatusCode)
	}
}

func TestGetFeedURLAuth(t *testing.T) {
	const feedName = "test"
	const secret = "b90e516e-b256-41ff-a84e-a9e8d5b6fe30"
	api := NewApiHandler(path.Join(baseDir, dataDir))

	req := httptest.NewRequest(http.MethodGet, "/api/feed/"+feedName+"?secret="+secret, nil)
	w := httptest.NewRecorder()

	api.ApiHandleFunc(w, req)

	res := w.Result()

	if res.StatusCode != 200 {
		t.Errorf("Expect code 200 but got %d", res.StatusCode)
	}
}
