package weread

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const mockSearchResponse = `{
  "books": [
    {
      "bookInfo": {
        "title": "三体",
        "author": "刘慈欣",
        "bookId": "CB_3p7052d00f3f8f66"
      }
    },
    {
      "bookInfo": {
        "title": "球状闪电",
        "author": "刘慈欣",
        "bookId": "CB_abc123def456"
      }
    }
  ]
}`

const mockEmptyResponse = `{"books": []}`

func newTestClient(ts *httptest.Server) *Client {
	cfg := DefaultConfig()
	cfg.BaseURL = ts.URL
	cfg.Rate = 0
	return NewClient(cfg)
}

func TestSearchSendsHeaders(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") == "" {
			t.Error("request carried no User-Agent")
		}
		if r.Header.Get("Referer") == "" {
			t.Error("request carried no Referer")
		}
		_, _ = w.Write([]byte(mockSearchResponse))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	_, err := c.Search(context.Background(), "三体", 10)
	if err != nil {
		t.Fatal(err)
	}
}

func TestSearchParsesBooks(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(mockSearchResponse))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	books, err := c.Search(context.Background(), "三体", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(books) != 2 {
		t.Fatalf("got %d books, want 2", len(books))
	}

	b := books[0]
	if b.Rank != 1 {
		t.Errorf("rank = %d, want 1", b.Rank)
	}
	if b.Title != "三体" {
		t.Errorf("title = %q, want 三体", b.Title)
	}
	if b.Author != "刘慈欣" {
		t.Errorf("author = %q, want 刘慈欣", b.Author)
	}
	if b.BookID != "CB_3p7052d00f3f8f66" {
		t.Errorf("bookId = %q", b.BookID)
	}
	wantURL := srv.URL + "/web/bookDetail/CB_3p7052d00f3f8f66"
	if b.URL != wantURL {
		t.Errorf("url = %q, want %q", b.URL, wantURL)
	}

	b2 := books[1]
	if b2.Rank != 2 {
		t.Errorf("rank = %d, want 2", b2.Rank)
	}
}

func TestSearchLimit(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(mockSearchResponse))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	books, err := c.Search(context.Background(), "刘慈欣", 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(books) != 1 {
		t.Errorf("got %d books with limit=1, want 1", len(books))
	}
}

func TestSearchEmptyResult(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(mockEmptyResponse))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	books, err := c.Search(context.Background(), "zzznoresults", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(books) != 0 {
		t.Errorf("got %d books, want 0", len(books))
	}
}

func TestSearchRetriesOn503(t *testing.T) {
	var hits int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		if hits < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_, _ = w.Write([]byte(mockSearchResponse))
	}))
	defer srv.Close()

	cfg := DefaultConfig()
	cfg.BaseURL = srv.URL
	cfg.Rate = 0
	cfg.Retries = 5
	c := NewClient(cfg)

	start := time.Now()
	books, err := c.Search(context.Background(), "三体", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(books) == 0 {
		t.Error("got 0 books after retries")
	}
	if hits != 3 {
		t.Errorf("server saw %d hits, want 3", hits)
	}
	if time.Since(start) < 500*time.Millisecond {
		t.Error("retries did not back off")
	}
}
