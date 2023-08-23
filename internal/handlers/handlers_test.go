package handlers

import (
	"bytes"
	"fmt"
	"io"
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

const testFeedName = "test"

type APITestRequest struct {
	method      string
	feed        string
	item        string
	body        io.Reader
	contentType string

	cookieAuthType AuthType
	queryAuthType  AuthType
}
type AuthType int

const (
	AuthTypeNone = iota
	AuthTypeAuth
	AuthTypeFail
)

func (t APITestRequest) performRequest() *http.Response {
	const goodSecret = "b90e516e-b256-41ff-a84e-a9e8d5b6fe30"
	const badSecret = "foo"

	api := NewApiHandler(path.Join(baseDir, dataDir))

	authQuery := ""
	switch t.queryAuthType {
	case AuthTypeAuth:
		authQuery = "?secret=" + goodSecret
	case AuthTypeFail:
		authQuery = "?secret=" + badSecret
	}

	path := "/api/feed/"
	if t.feed == "" {
		path = path + testFeedName
	} else {
		path = path + t.feed
	}

	if t.item != "" {
		path = path + "/" + url.QueryEscape(t.item)
	}

	req := httptest.NewRequest(t.method, path+authQuery, t.body)

	switch t.cookieAuthType {
	case AuthTypeAuth:
		req.AddCookie(&http.Cookie{Name: "Secret", Value: goodSecret})
	case AuthTypeFail:
		req.AddCookie(&http.Cookie{Name: "Secret", Value: badSecret})
	}

	if t.contentType != "" {
		req.Header.Add("Content-type", t.contentType)
	}
	w := httptest.NewRecorder()

	api.ApiHandleFunc(w, req)

	return w.Result()
}

func TestCreateFeed(t *testing.T) {
	const feedName = "6ff1146b683086b8f59f550189a8f91f"

	t.Cleanup(func() {
		os.RemoveAll(path.Join(baseDir, dataDir, feedName))
	})

	res := APITestRequest{
		method: http.MethodGet,
		feed:   feedName,
		body:   nil,
	}.performRequest()

	if res.StatusCode != 200 {
		t.Errorf("Expect code 200 but got %d", res.StatusCode)
	}

	// Read cookie
	cookies := res.Cookies()
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
	res := APITestRequest{
		method: http.MethodGet,
		body:   nil,
	}.performRequest()

	if res.StatusCode != 401 {
		spew.Dump(res)
		t.Errorf("Expect code 401 but got %d", res.StatusCode)
	}
}

func TestGetFeedBadCookie(t *testing.T) {
	res := APITestRequest{
		method:         http.MethodGet,
		body:           nil,
		cookieAuthType: AuthTypeFail,
	}.performRequest()

	if res.StatusCode != 401 {
		t.Errorf("Expect code 401 but got %d", res.StatusCode)
	}
}

func TestGetFeedCookieAuth(t *testing.T) {

	res := APITestRequest{
		method:         http.MethodGet,
		feed:           testFeedName,
		body:           nil,
		cookieAuthType: AuthTypeAuth,
	}.performRequest()

	if res.StatusCode != 200 {
		t.Errorf("Expect code 200 but got %d", res.StatusCode)
	}
}

func TestGetFeedBadQuerySecret(t *testing.T) {
	res := APITestRequest{
		method:        http.MethodGet,
		body:          nil,
		queryAuthType: AuthTypeFail,
	}.performRequest()

	if res.StatusCode != 401 {
		t.Errorf("Expect code 401 but got %d", res.StatusCode)
	}
}

func TestGetFeedURLAuth(t *testing.T) {
	res := APITestRequest{
		method:        http.MethodGet,
		body:          nil,
		queryAuthType: AuthTypeAuth,
	}.performRequest()

	if res.StatusCode != 200 {
		t.Errorf("Expect code 200 but got %d", res.StatusCode)
	}
}

func TestGetFeedItemNoCredentials(t *testing.T) {
	const item = "Pasted Image 1.png"

	res := APITestRequest{
		method: http.MethodGet,
		item:   item,
		body:   nil,
	}.performRequest()

	if res.StatusCode != 401 {
		t.Errorf("Expect code 401 but got %d", res.StatusCode)
	}
}

func TestGetFeedItem(t *testing.T) {
	const item = "Pasted Image 1.png"

	res := APITestRequest{
		method:         http.MethodGet,
		item:           item,
		body:           nil,
		cookieAuthType: AuthTypeAuth,
	}.performRequest()

	if res.StatusCode != 200 {
		t.Errorf("Expect code 200 but got %d", res.StatusCode)
	}
}

func TestGetFeedItemNonExistentFeed(t *testing.T) {
	const item = "Pasted Image 1.png"

	res := APITestRequest{
		method: http.MethodGet,
		feed:   "foo",
		item:   item,
	}.performRequest()

	if res.StatusCode != 404 {
		t.Errorf("Expect code 404 but got %d", res.StatusCode)
	}
}

func TestGetFeedItemNonExistent(t *testing.T) {
	res := APITestRequest{
		method:         http.MethodGet,
		item:           "foo",
		cookieAuthType: AuthTypeAuth,
	}.performRequest()

	if res.StatusCode != 404 {
		b, err := io.ReadAll(res.Body)
		if err != nil {
			t.Errorf("Expect code 400 but got %d (%s)", res.StatusCode, err.Error())
		}
		t.Errorf("Expect code 404 but got %d (%s)", res.StatusCode, string(b))
	}
}

func TestSetPinNoCredentials(t *testing.T) {
	const pin = "1234"

	buf := bytes.NewBuffer([]byte(pin))

	res := APITestRequest{
		method: http.MethodPatch,
		body:   buf,
	}.performRequest()

	if res.StatusCode != 401 {
		t.Errorf("Expect code 401 but got %d", res.StatusCode)
	}
}

func TestSetPin(t *testing.T) {
	pinPath := path.Join(baseDir, dataDir, testFeedName, "pin")
	t.Cleanup(func() {
		os.Remove(pinPath)
	})

	const pin = "1234"

	buf := bytes.NewBuffer([]byte(pin))

	res := APITestRequest{
		method:         http.MethodPatch,
		body:           buf,
		cookieAuthType: AuthTypeAuth,
	}.performRequest()

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

func TestSetBadPin(t *testing.T) {
	pinPath := path.Join(baseDir, dataDir, testFeedName, "pin")
	t.Cleanup(func() {
		os.Remove(pinPath)
	})

	const pin = "123"

	buf := bytes.NewBuffer([]byte(pin))

	res := APITestRequest{
		method:         http.MethodPatch,
		body:           buf,
		cookieAuthType: AuthTypeAuth,
	}.performRequest()

	if res.StatusCode != 400 {
		t.Errorf("Expect code 400 but got %d", res.StatusCode)
	}
}

func TestAddAndRemoveContent(t *testing.T) {
	filePath := path.Join(baseDir, dataDir, testFeedName, "Pasted Image 1.png")
	newFilePath := path.Join(baseDir, dataDir, testFeedName, "Pasted Image 2.png")

	t.Cleanup(func() {
		os.Remove(newFilePath)
	})

	reader, err := os.Open(filePath)

	res := APITestRequest{
		method:         http.MethodPost,
		body:           reader,
		cookieAuthType: AuthTypeAuth,
		contentType:    "image/png",
	}.performRequest()

	if res.StatusCode != 200 {
		b, err := io.ReadAll(res.Body)
		if err != nil {
			t.Error(err.Error())
		}
		t.Errorf("Expect code 200 but got %d (%s)", res.StatusCode, string(b))
	}

	_, err = os.Stat(newFilePath)
	if err != nil {
		t.Error(err.Error())
	}

	// Delete request
	res = APITestRequest{
		method:         http.MethodDelete,
		item:           "Pasted Image 2.png",
		cookieAuthType: AuthTypeAuth,
	}.performRequest()

	if res.StatusCode != 200 {
		t.Errorf("Expect code 200 but got %d", res.StatusCode)
	}

	_, err = os.Stat(newFilePath)
	if err == nil {
		t.Error("Expected file to be delete and it isn't")
	}
}

func TestAddContentTooBig(t *testing.T) {
	b := bytes.NewBuffer(make([]byte, 6*1024*1024))

	res := APITestRequest{
		method:         http.MethodPost,
		body:           b,
		cookieAuthType: AuthTypeAuth,
		contentType:    "image/png",
	}.performRequest()

	if res.StatusCode != 413 {
		b, err := io.ReadAll(res.Body)
		if err != nil {
			t.Errorf("Expect code 413 but got %d (%s)", res.StatusCode, err.Error())
		}
		t.Errorf("Expect code 413 but got %d (%s)", res.StatusCode, string(b))
	}
}

func TestAddContentWrongContentType(t *testing.T) {
	res := APITestRequest{
		method:         http.MethodPost,
		cookieAuthType: AuthTypeAuth,
		contentType:    "foo/bar",
	}.performRequest()

	if res.StatusCode != 400 {
		b, err := io.ReadAll(res.Body)
		if err != nil {
			t.Errorf("Expect code 400 but got %d (%s)", res.StatusCode, err.Error())
		}
		t.Errorf("Expect code 400 but got %d (%s)", res.StatusCode, string(b))
	}
}

func TestAddContentNonExistentFeed(t *testing.T) {
	res := APITestRequest{
		method:      http.MethodPost,
		feed:        "foo",
		contentType: "foo/bar",
	}.performRequest()

	if res.StatusCode != 404 {
		b, err := io.ReadAll(res.Body)
		if err != nil {
			t.Errorf("Expect code 400 but got %d (%s)", res.StatusCode, err.Error())
		}
		t.Errorf("Expect code 400 but got %d (%s)", res.StatusCode, string(b))
	}
}
func TestRemoveNonExistentContent(t *testing.T) {
	res := APITestRequest{
		method:         http.MethodDelete,
		cookieAuthType: AuthTypeAuth,
		item:           "foo",
	}.performRequest()

	if res.StatusCode != 404 {
		b, err := io.ReadAll(res.Body)
		if err != nil {
			t.Errorf("Expect code 404 but got %d (%s)", res.StatusCode, err.Error())
		}
		t.Errorf("Expect code 404 but got %d (%s)", res.StatusCode, string(b))
	}
}

func TestRemoveItemNonExistentFeed(t *testing.T) {
	res := APITestRequest{
		method: http.MethodDelete,
		feed:   "foo",
		item:   "foo",
	}.performRequest()

	if res.StatusCode != 404 {
		b, err := io.ReadAll(res.Body)
		if err != nil {
			t.Errorf("Expect code 404 but got %d (%s)", res.StatusCode, err.Error())
		}
		t.Errorf("Expect code 404 but got %d (%s)", res.StatusCode, string(b))
	}
}
