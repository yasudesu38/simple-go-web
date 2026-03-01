package main

import (
    "fmt"
    "net/http"
    "net/http/httptest"
    "os"
    "strings"
    "testing"
)

func TestGetVideoID(t *testing.T) {
    cases := map[string]string{
        "https://www.youtube.com/watch?v=abcd1234": "abcd1234",
        "https://youtu.be/xyz987":                "xyz987",
        "https://www.youtube.com/embed/qwerty":   "qwerty",
        "https://youtube.com/watch?v=foo&other=bar": "foo",
    }
    for in, want := range cases {
        got, err := getVideoID(in)
        if err != nil {
            t.Errorf("getVideoID(%s) error: %v", in, err)
            continue
        }
        if got != want {
            t.Errorf("getVideoID(%s) = %s; want %s", in, got, want)
        }
    }
}

func TestFetchTitleAndTranscript(t *testing.T) {
    mux := http.NewServeMux()
    mux.HandleFunc("/oembed", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintln(w, `{"title":"Test Video"}`)
    })
    mux.HandleFunc("/timedtext", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintln(w, `<transcript><text>line1</text><text>line2</text></transcript>`)
    })
    srv := httptest.NewServer(mux)
    defer srv.Close()

    // override global variables
    httpClient = srv.Client()
    oembedBase = srv.URL + "/oembed"
    transcriptBase = srv.URL + "/timedtext"

    // call functions with arbitrary inputs
    title, err := fetchTitle("anything")
    if err != nil || title != "Test Video" {
        t.Fatalf("fetchTitle returned %q, %v", title, err)
    }
    lines, err := fetchTranscript("id123")
    if err != nil || len(lines) != 2 || lines[0] != "line1" {
        t.Fatalf("fetchTranscript returned %v, %v", lines, err)
    }
}

func TestHandlerIntegration(t *testing.T) {
    mux := http.NewServeMux()
    mux.HandleFunc("/oembed", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintln(w, `{"title":"StubTitle"}`)
    })
    mux.HandleFunc("/timedtext", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintln(w, `<transcript><text>foo</text></transcript>`)
    })
    srv := httptest.NewServer(mux)
    defer srv.Close()

    // override http behavior
    httpClient = srv.Client()
    oembedBase = srv.URL + "/oembed"
    transcriptBase = srv.URL + "/timedtext"
    videoURL = "https://youtu.be/anything"

    req := httptest.NewRequest("GET", "/", nil)
    w := httptest.NewRecorder()
    handler(w, req)
    resp := w.Result()
    if resp.StatusCode != http.StatusOK {
        t.Fatalf("handler status = %d", resp.StatusCode)
    }
    body := w.Body.String()
    if !strings.Contains(body, "StubTitle") || !strings.Contains(body, "foo") {
        t.Errorf("handler body did not include expected content: %s", body)
    }
}

func TestMainEnv(t *testing.T) {
    // ensure VIDEO_URL default when not set
    os.Unsetenv("VIDEO_URL")
    loadConfig()
    if videoURL == "" {
        t.Fatal("expected default videoURL to be set")
    }
}
