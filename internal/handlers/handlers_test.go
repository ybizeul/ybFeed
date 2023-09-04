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

	"github.com/Appboy/webpush-go"
	"github.com/davecgh/go-spew/spew"
	"github.com/ybizeul/ybfeed/internal/feed"
)

const baseDir = "../../test/"
const dataDir = "./data"
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

func (t APITestRequest) performRequest() (*http.Response, error) {
	const goodSecret = "b90e516e-b256-41ff-a84e-a9e8d5b6fe30"
	const badSecret = "foo"

	api, err := NewApiHandler(path.Join(baseDir, dataDir))
	if err != nil {
		return nil, err
	}
	api.MaxBodySize = 5 * 1024 * 1024
	r := api.GetServer()

	authQuery := ""
	switch t.queryAuthType {
	case AuthTypeAuth:
		authQuery = "?secret=" + goodSecret
	case AuthTypeFail:
		authQuery = "?secret=" + badSecret
	}

	path := "/api/feed/"
	if t.feed == "" {
		path = path + url.QueryEscape(testFeedName)
	} else {
		path = path + url.QueryEscape(t.feed)
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

	r.ServeHTTP(w, req)
	//api.ApiHandleFunc(w, req)

	return w.Result(), nil
}

func TestFirstStart(t *testing.T) {
	newDir := "FOO"
	t.Cleanup(func() {
		os.RemoveAll(path.Join(baseDir, newDir))
	})

	_, err := NewApiHandler(path.Join(baseDir, newDir))

	if err != nil {
		t.Error(err)
	}

	c, err := APIConfigFromFile(path.Join(baseDir, dataDir, "config.json"))
	if err != nil {
		t.Error(err)
	}

	if c.NotificationSettings == nil ||
		len(c.NotificationSettings.VAPIDPrivateKey) == 0 ||
		len(c.NotificationSettings.VAPIDPrivateKey) == 0 {
		t.Error("Invalid config file")
	}

}
func TestIndexHtml(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	RootHandlerFunc(w, req)
	res := w.Result()
	if res.StatusCode != 200 {
		t.Errorf("Expect code 200 but got %d", res.StatusCode)
	}
}

func TestServiceWorker(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/service-worker.js", nil)
	w := httptest.NewRecorder()
	RootHandlerFunc(w, req)
	res := w.Result()
	if res.StatusCode != 200 {
		t.Errorf("Expect code 200 but got %d", res.StatusCode)
	}
}

func TestCreateFeed(t *testing.T) {
	const feedName = "6ff1146b6830 #86b8f59f550189a8f91f"

	t.Cleanup(func() {
		os.RemoveAll(path.Join(baseDir, dataDir, feedName))
	})

	res, _ := APITestRequest{
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
	res, _ := APITestRequest{
		method: http.MethodGet,
		body:   nil,
	}.performRequest()

	if res.StatusCode != 401 {
		spew.Dump(res)
		t.Errorf("Expect code 401 but got %d", res.StatusCode)
	}
	// Read cookie
	cookies := res.Cookies()
	found := false
	fmt.Println(cookies)
	for _, c := range cookies {
		found = (c.Name == "Secret")
	}

	if found == true {
		t.Errorf("Cookie has been set in reply")
	}
}

func TestGetFeedBadCookie(t *testing.T) {
	res, _ := APITestRequest{
		method:         http.MethodGet,
		body:           nil,
		cookieAuthType: AuthTypeFail,
	}.performRequest()

	if res.StatusCode != 401 {
		t.Errorf("Expect code 401 but got %d", res.StatusCode)
	}
}

func TestGetFeedCookieAuth(t *testing.T) {

	res, _ := APITestRequest{
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
	res, _ := APITestRequest{
		method:        http.MethodGet,
		body:          nil,
		queryAuthType: AuthTypeFail,
	}.performRequest()

	if res.StatusCode != 401 {
		t.Errorf("Expect code 401 but got %d", res.StatusCode)
	}
}

func TestGetFeedURLAuth(t *testing.T) {
	res, _ := APITestRequest{
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

	res, _ := APITestRequest{
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

	res, _ := APITestRequest{
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

	res, _ := APITestRequest{
		method: http.MethodGet,
		feed:   "foo",
		item:   item,
	}.performRequest()

	if res.StatusCode != 404 {
		t.Errorf("Expect code 404 but got %d", res.StatusCode)
	}
}

func TestGetFeedItemNonExistent(t *testing.T) {
	res, _ := APITestRequest{
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

	res, _ := APITestRequest{
		method: http.MethodPatch,
		body:   buf,
	}.performRequest()

	if res.StatusCode != 401 {
		t.Errorf("Expect code 401 but got %d", res.StatusCode)
	}
}

func TestSetPin(t *testing.T) {
	t.Cleanup(func() {
		c, _ := feed.FeedConfigForFeed(
			&feed.Feed{
				Path: path.Join(baseDir, dataDir, testFeedName),
			},
		)
		c.PIN = nil
		_ = c.Write()
	})

	const pin = "1234"

	buf := bytes.NewBuffer([]byte(pin))

	res, _ := APITestRequest{
		method:         http.MethodPatch,
		body:           buf,
		cookieAuthType: AuthTypeAuth,
	}.performRequest()

	if res.StatusCode != 200 {
		t.Errorf("Expect code 200 but got %d", res.StatusCode)
	}

	c, err := feed.FeedConfigForFeed(
		&feed.Feed{
			Path: path.Join(baseDir, dataDir, testFeedName),
		},
	)

	if err != nil {
		t.Errorf(err.Error())
	}
	pin_written := c.PIN.PIN

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

	res, _ := APITestRequest{
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

	reader, _ := os.Open(filePath)

	res, _ := APITestRequest{
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

	_, err := os.Stat(newFilePath)
	if err != nil {
		t.Error(err.Error())
	}

	// Delete request
	res, _ = APITestRequest{
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

	res, _ := APITestRequest{
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
	res, _ := APITestRequest{
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
	res, _ := APITestRequest{
		method:      http.MethodPost,
		feed:        "foo",
		contentType: "foo/bar",
	}.performRequest()

	if res.StatusCode != 404 {
		b, err := io.ReadAll(res.Body)
		if err != nil {
			t.Errorf("Expect code 404 but got %d (%s)", res.StatusCode, err.Error())
		}
		t.Errorf("Expect code 400 but got %d (%s)", res.StatusCode, string(b))
	}
}
func TestRemoveNonExistentContent(t *testing.T) {
	res, _ := APITestRequest{
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
	res, _ := APITestRequest{
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

func TestSubscribeFeedNotifications(t *testing.T) {
	body := `{"endpoint":"http://test.com","keys":{"auth":"AUTH","p256dh":"P256DH"}}`
	b := bytes.NewBuffer([]byte(body))

	t.Cleanup(func() {
		f, _ := feed.GetFeed(path.Join(baseDir, dataDir, testFeedName))
		f.Config.Subscriptions = []webpush.Subscription{}
		_ = f.Config.Write()
	})

	res, _ := APITestRequest{
		method:         http.MethodPost,
		feed:           testFeedName,
		item:           "subscription",
		cookieAuthType: AuthTypeAuth,
		body:           b,
	}.performRequest()

	if res.StatusCode != 200 {
		t.Errorf("Expected 200 but got %d", res.StatusCode)
	}
	f, err := feed.GetFeed(path.Join(baseDir, dataDir, testFeedName))

	if err != nil {
		t.Error(err)
	}

	if len(f.Config.Subscriptions) == 0 {
		t.Error("No subscriptions found")
		return
	}

	if f.Config.Subscriptions[0].Endpoint != "http://test.com" {
		t.Errorf("Bad endpoint, got %s", f.Config.Subscriptions[0].Endpoint)
	}

	if f.Config.Subscriptions[0].Keys.Auth != "AUTH" {
		t.Errorf("Bad auth, got %s", f.Config.Subscriptions[0].Keys.Auth)
	}

	if f.Config.Subscriptions[0].Keys.P256dh != "P256DH" {
		t.Errorf("Bad p256dh, got %s", f.Config.Subscriptions[0].Keys.P256dh)
	}
}
