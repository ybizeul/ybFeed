package feed

import (
	"bytes"
	"os"
	"testing"
)

func TestGetFeedItemData(t *testing.T) {
	t.Cleanup(func() {
		os.RemoveAll("tests/feed1")
	})
	f, err := NewFeed("tests/feed1")
	if err != nil {
		t.Fatal(err)
	}

	reader := bytes.NewReader([]byte("test"))

	err = f.AddItem("text/plain", reader)
	if err != nil {
		t.Fatal(err)
	}

	pf, err := f.Public()
	if err != nil {
		t.Fatal(err)
	}
	i := pf.Items[0]
	b, err := f.GetItemData(i.Name)
	if len(b) == 0 || err != nil {
		t.Fatal(err)
	}
}

func TestPathTraversalGet(t *testing.T) {
	t.Cleanup(func() {
		os.RemoveAll("tests/feed1")
		os.RemoveAll("tests/feed2")
	})
	_, err := NewFeed("tests/feed1")
	if err != nil {
		t.Fatal(err)
	}

	f, err := NewFeed("tests/feed2")
	if err != nil {
		t.Fatal(err)
	}

	b, err := f.GetItemData("../feed1/config.json")

	if len(b) != 0 || err == nil {
		t.Fatal("Path traversal not blocked")
	}
}

func TestPathTraversalDelete(t *testing.T) {
	t.Cleanup(func() {
		os.RemoveAll("tests/feed1")
		os.RemoveAll("tests/feed2")
	})
	_, err := NewFeed("tests/feed1")
	if err != nil {
		t.Fatal(err)
	}

	f, err := NewFeed("tests/feed2")
	if err != nil {
		t.Fatal(err)
	}

	err = f.RemoveItem("../feed1/config.json")

	if err == nil {
		t.Fatal("Path traversal not blocked")
	}
}

func TestPathTraversalPublicItem(t *testing.T) {
	t.Cleanup(func() {
		os.RemoveAll("tests/feed1")
		os.RemoveAll("tests/feed2")
	})
	_, err := NewFeed("tests/feed1")
	if err != nil {
		t.Fatal(err)
	}

	f, err := NewFeed("tests/feed2")
	if err != nil {
		t.Fatal(err)
	}

	p, err := f.GetPublicItem("../feed1/config.json")

	if p != nil || err == nil {
		t.Fatal("Path traversal not blocked")
	}
}

func TestPublicItem(t *testing.T) {
	t.Cleanup(func() {
		os.RemoveAll("tests/feed1")
	})
	f, err := NewFeed("tests/feed1")
	if err != nil {
		t.Fatal(err)
	}

	reader := bytes.NewReader([]byte("test"))

	err = f.AddItem("text/plain", reader)
	if err != nil {
		t.Fatal(err)
	}

	pf, err := f.Public()
	if err != nil {
		t.Fatal(err)
	}
	i := pf.Items[0]

	p, err := f.GetPublicItem(i.Name)

	if p == nil || err != nil {
		t.Fatal(err)
	}
}
