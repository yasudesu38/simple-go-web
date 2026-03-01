package main

import (
    "encoding/json"
    "encoding/xml"
    "errors"
    "fmt"
    "html/template"
    "log"
    "net/http"
    "net/url"
    "os"
    "strings"
    "time"
)

// configuration, can be replaced with file-based source later
var videoURL string

func loadConfig() {
    videoURL = os.Getenv("VIDEO_URL")
    if videoURL == "" {
        // default to the example video from the requirement
        videoURL = "https://www.youtube.com/watch?v=IYfvmAbwRvs&t=1898s"
    }
}

func init() {
    loadConfig()
}

// helper to extract YouTube video ID from various URL forms
func getVideoID(raw string) (string, error) {
    u, err := url.Parse(raw)
    if err != nil {
        return "", err
    }
    switch u.Host {
    case "youtu.be":
        return strings.Trim(u.Path, "/"), nil
    case "www.youtube.com", "youtube.com":
        qs := u.Query()
        if v := qs.Get("v"); v != "" {
            return v, nil
        }
        // sometimes embedded in path
        if strings.HasPrefix(u.Path, "/embed/") {
            return strings.TrimPrefix(u.Path, "/embed/"), nil
        }
    }
    return "", errors.New("could not determine video ID")
}

// injectable variables for HTTP interactions (override in tests)
var httpClient = &http.Client{Timeout: 10 * time.Second}
var oembedBase = "https://www.youtube.com/oembed"
var transcriptBase = "https://video.google.com/timedtext"

// fetch title using oEmbed endpoint
func fetchTitle(rawURL string) (string, error) {
    resp, err := httpClient.Get(oembedBase + "?url=" + url.QueryEscape(rawURL) + "&format=json")
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
        return "", fmt.Errorf("oembed returned status %d", resp.StatusCode)
    }
    var o struct {
        Title string `json:"title"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&o); err != nil {
        return "", err
    }
    return o.Title, nil
}

// fetch english transcript lines via timedtext API
func fetchTranscript(videoID string) ([]string, error) {
    urlStr := transcriptBase + "?lang=en&v=" + url.QueryEscape(videoID)
    resp, err := httpClient.Get(urlStr)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("transcript endpoint returned %d", resp.StatusCode)
    }
    var envelope struct {
        Text []struct {
            XMLName xml.Name `xml:"text"`
            Body    string   `xml:",chardata"`
        } `xml:"text"`
    }
    if err := xml.NewDecoder(resp.Body).Decode(&envelope); err != nil {
        return nil, err
    }
    lines := make([]string, len(envelope.Text))
    for i, t := range envelope.Text {
        lines[i] = t.Body
    }
    return lines, nil
}

// page data for template
type pageData struct {
    Title      string
    VideoURL   string
    Transcript []string
}

var pageTmpl = template.Must(template.New("page").Parse(`<!doctype html>
<html lang="en">
<head><meta charset="utf-8"><title>{{.Title}}</title></head>
<body>
<h1>{{.Title}}</h1>
<p>Source: <a href="{{.VideoURL}}">{{.VideoURL}}</a></p>
<pre>{{range .Transcript}}{{.}}
{{end}}</pre>
</body>
</html>`))

func handler(w http.ResponseWriter, r *http.Request) {
    vid, err := getVideoID(videoURL)
    if err != nil {
        http.Error(w, "invalid video URL", http.StatusInternalServerError)
        return
    }
    title, err := fetchTitle(videoURL)
    if err != nil {
        http.Error(w, "failed to fetch title: "+err.Error(), http.StatusInternalServerError)
        return
    }
    transcript, err := fetchTranscript(vid)
    if err != nil {
        http.Error(w, "failed to fetch transcript: "+err.Error(), http.StatusInternalServerError)
        return
    }
    data := pageData{Title: title, VideoURL: videoURL, Transcript: transcript}
    if err := pageTmpl.Execute(w, data); err != nil {
        log.Println("template execute error:", err)
    }
}

func main() {
    http.HandleFunc("/", handler)
    log.Println("Starting server on :8080, video URL:", videoURL)
    if err := http.ListenAndServe(":8080", nil); err != nil {
        log.Fatalf("could not start server: %v", err)
    }
}
