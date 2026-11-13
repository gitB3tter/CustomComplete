package main

import (
	"context"
	"customcomplete/internal/catalog"
	"customcomplete/internal/constraints"
	"customcomplete/internal/mediatypes"
	"customcomplete/internal/progress"
	"customcomplete/internal/store"
	"customcomplete/internal/suggestions"
	"customcomplete/internal/textprocessing"
	"customcomplete/internal/ui"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"slices"
	"sync"
	"testing"
	"time"
)

func TestCreateToSuggestionsFlow(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	readWrite, err := store.GetReadWrite(dbPath)
	if err != nil {
		t.Fatalf("create readWrite for test: %v", err)
	}
	t.Cleanup(func() { readWrite.Close() })

	read, err := store.GetRead(dbPath)
	if err != nil {
		t.Fatalf("create read for test: %v", err)
	}
	t.Cleanup(func() { read.Close() })

	broker := progress.NewBroker()
	jobs := make(chan int64, 100)
	var wg sync.WaitGroup
	ctxWorker, cancelWorkers := context.WithCancel(context.Background())
	t.Cleanup(func() {
		cancelWorkers()
		wg.Wait()
	})
	wg.Add(1)
	go textprocessing.Worker(ctxWorker, 1, readWrite, read, broker, jobs, &wg)

	service := catalog.NewService(readWrite, read, broker, jobs)

	testLabelRaw := "testlabel"
	testLabel, codedErr := constraints.ParseLabel(testLabelRaw)
	if codedErr != nil {
		t.Fatalf(`parse "%s" as label: %v`, testLabelRaw, err)
	}
	testContentsRaw := []byte("CustomComplete turns text into your own autocomplete engine. Upload, paste, or type text; it learns from it and suggests the next word or completes the one you're typing. Go + SQLite, with a server-rendered Web-UI and a REST API.")
	testContents, codedErr := catalog.ParseSourceContents(testContentsRaw)
	if codedErr != nil {
		t.Fatalf(`parse "%s" as contents: %v`, testContentsRaw, err)
	}

	_, docID, err := service.CreateDocument(testLabel, constraints.DeriveSafeFilename(testLabelRaw), testContents)
	if err != nil {
		t.Fatalf(`error creating Document: %v`, err)
	}

	deadline := time.Now().Add(2 * time.Second)
	for {
		status, err := service.GetDocumentStatus(docID)
		if err != nil {
			t.Fatalf("get progress: %v", err)
		}
		if status == catalog.StatusComplete {
			break
		}
		if status == catalog.StatusFailed {
			t.Fatalf("document processing failed")
		}
		if time.Now().After(deadline) {
			t.Fatalf("test-text processing took too long (last status: %s)", status)
		}
		time.Sleep(10 * time.Millisecond)
	}

	baseTemplates, err := ui.NewBaseTemplates()
	if err != nil {
		t.Errorf("get basetemplates: %v", err)
	}

	handler, err := suggestions.NewHandler(read, baseTemplates)
	if err != nil {
		t.Fatalf("suggestions handler: %v", err)
	}
	mux := http.NewServeMux()
	handler.Register(mux)

	tests := map[string][]string{
		"CustomComplete ":                 {"turns"},
		"CustomComplete turns text into ": {"your"},
		"CustomComplete turns text into your own autocomplete engine ":                {"Upload"},
		"CustomComplete turns text into your own autocomplete engine. Upload ":        {"paste"},
		"CustomComplete turns text into your own autocomplete engine. Upload, p":      {"paste"},
		"CustomComplete turns text into your own autocomplete engine. Upload, paste,": {"or"},
		"a":       {"and", "autocomplete"},
		"t":       {"turns", "type", "the", "typing"},
		"YOU'RE ": {"typing"},
		"zzzzz":   {},
	}

	for context, expect := range tests {
		t.Run(context, func(t *testing.T) {
			rec := httptest.NewRecorder()
			query := url.Values{}
			query.Add("text", context)
			target := suggestions.Path.URL()
			target.RawQuery = query.Encode()
			r := httptest.NewRequest("GET", target.String(), nil)
			r.Header.Add("Accept", mediatypes.LinksetJSON.String())
			mux.ServeHTTP(rec, r)
			var body struct {
				Linkset []struct {
					Item []struct {
						Href  string `json:"href"`
						Title string `json:"title"`
					} `json:"item"`
				} `json:"linkset"`
			}
			if rec.Code != http.StatusOK {
				t.Fatalf(`received status %d for suggestions for "%s"`, rec.Code, context)
			} else if rec.Header().Get("Content-Type") != mediatypes.LinksetJSON.String() {
				t.Fatalf(`received content type %s for suggestions for "%s"`, rec.Header().Get("Content-Type"), context)
			} else if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("unmarshal body %q: %v", rec.Body.String(), err)
			} else if len(body.Linkset) == 0 {
				t.Fatalf("expected non-empty linkset, got: %s", rec.Body.String())
			}
			var suggestions []string
			for _, item := range body.Linkset[0].Item {
				suggestions = append(suggestions, item.Title)
			}
			if len(expect) == 0 && len(suggestions) != 0 {
				t.Fatalf("expected no suggestions, got %v", suggestions)
			}
			for _, expected := range expect {
				if !slices.Contains(suggestions, expected) {
					t.Fatalf(`suggestions dont contain "%s" for context "%s"`, expected, context)
				}
			}
		})

	}

}
