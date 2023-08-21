package handlers

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
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

func TestGetFeedItemNoCredentials(t *testing.T) {
	const feedName = "test"
	const secret = "b90e516e-b256-41ff-a84e-a9e8d5b6fe30"
	const item = "Pasted Image 1.png"

	api := NewApiHandler(path.Join(baseDir, dataDir))

	req := httptest.NewRequest(http.MethodGet, "/api/feed/"+feedName+"/"+url.QueryEscape(item), nil)
	w := httptest.NewRecorder()

	api.ApiHandleFunc(w, req)

	res := w.Result()

	if res.StatusCode != 401 {
		t.Errorf("Expect code 401 but got %d", res.StatusCode)
	}
}

func TestGetFeedItem(t *testing.T) {
	const feedName = "test"
	const secret = "b90e516e-b256-41ff-a84e-a9e8d5b6fe30"
	const item = "Pasted Image 1.png"

	api := NewApiHandler(path.Join(baseDir, dataDir))

	req := httptest.NewRequest(http.MethodGet, "/api/feed/"+feedName+"/"+url.QueryEscape(item), nil)
	req.AddCookie(&http.Cookie{Name: "Secret", Value: secret})
	w := httptest.NewRecorder()

	api.ApiHandleFunc(w, req)

	res := w.Result()

	if res.StatusCode != 200 {
		t.Errorf("Expect code 200 but got %d", res.StatusCode)
	}
}

func TestSetPinNoCredentials(t *testing.T) {
	const feedName = "test"
	const secret = "b90e516e-b256-41ff-a84e-a9e8d5b6fe30"
	const pin = "1234"

	api := NewApiHandler(path.Join(baseDir, dataDir))
	buf := bytes.NewBuffer([]byte(pin))
	req := httptest.NewRequest(http.MethodPatch, "/api/feed/"+feedName, buf)

	w := httptest.NewRecorder()

	api.ApiHandleFunc(w, req)

	res := w.Result()

	if res.StatusCode != 401 {
		t.Errorf("Expect code 401 but got %d", res.StatusCode)
	}
}

func TestSetPin(t *testing.T) {

	const feedName = "test"
	const secret = "b90e516e-b256-41ff-a84e-a9e8d5b6fe30"
	const pin = "1234"
	pinPath := path.Join(baseDir, dataDir, feedName, "pin")

	t.Cleanup(func() {
		os.Remove(pinPath)
	})

	api := NewApiHandler(path.Join(baseDir, dataDir))
	buf := bytes.NewBuffer([]byte(pin))
	req := httptest.NewRequest(http.MethodPatch, "/api/feed/"+feedName, buf)

	req.AddCookie(&http.Cookie{Name: "Secret", Value: secret})
	w := httptest.NewRecorder()

	api.ApiHandleFunc(w, req)

	res := w.Result()

	if res.StatusCode != 200 {
		t.Errorf("Expect code 200 but got %d", res.StatusCode)
	}

	b, err := os.ReadFile(pinPath)
	if err != nil {
		t.Error(err.Error())
	}

	pin_written := string(b)

	if pin_written != pin {
		t.Errorf("Expected PIN %s but got %s", pin, pin_written)
	}
}

func TestAddAndRemoveContent(t *testing.T) {

	const feedName = "test"
	const secret = "b90e516e-b256-41ff-a84e-a9e8d5b6fe30"
	filePath := path.Join(baseDir, dataDir, feedName, "Pasted Image 1.png")
	newFilePath := path.Join(baseDir, dataDir, feedName, "Pasted Image 2.png")

	t.Cleanup(func() {
		os.Remove(newFilePath)
	})

	api := NewApiHandler(path.Join(baseDir, dataDir))
	reader, err := os.Open(filePath)

	if err != nil {
		t.Error(err.Error())
	}

	req := httptest.NewRequest(http.MethodPost, "/api/feed/"+feedName, reader)

	req.AddCookie(&http.Cookie{Name: "Secret", Value: secret})
	req.Header.Add("Content-type", "image/png")
	w := httptest.NewRecorder()

	api.ApiHandleFunc(w, req)

	res := w.Result()

	if res.StatusCode != 200 {
		t.Errorf("Expect code 200 but got %d", res.StatusCode)
	}

	_, err = os.Stat(newFilePath)
	if err != nil {
		t.Error(err.Error())
	}

	// Delete request
	req = httptest.NewRequest(http.MethodDelete, "/api/feed/"+feedName+"/"+url.QueryEscape("Pasted Image 2.png"), nil)

	req.AddCookie(&http.Cookie{Name: "Secret", Value: secret})
	req.Header.Add("Content-type", "image/png")
	w = httptest.NewRecorder()

	api.ApiHandleFunc(w, req)

	res = w.Result()

	if res.StatusCode != 200 {
		t.Errorf("Expect code 200 but got %d", res.StatusCode)
	}

	_, err = os.Stat(newFilePath)
	if err == nil {
		t.Error("Expected file to be delete and it isn't")
	}
}
