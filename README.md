# xai-go

Go wrapper client for X.ai gRPC APIs.

## Quick start

- Install:

  ```sh
  go get github.com/bibyzan/xai-go
  ```

- Use:

  ```go
  package main

  import (
      "context"
      "fmt"
      "log"
      "os"
      "time"

      "github.com/bibyzan/xai-go"
  )

  func main() {
      apiKey := os.Getenv("XAI_API_KEY")
      if apiKey == "" {
          log.Fatal("set XAI_API_KEY")
      }

      client := xai.NewClient(apiKey)
      defer func() { _ = client.Close() }()

      // Example 1: X source only
      search := xai.SearchParametersXOn(true, 3)
      chat := client.ChatCreateWithPrompt("grok-2", "Explain superconductors like I am five.", search)

      // Example 2: Mixed sources with passthrough arguments and date range
      from := time.Now().AddDate(0, -1, 0) // last month
      web := xai.WebSource(nil, []string{"wikipedia.org", "x.ai"}, "", true)
      rss := xai.RssSource([]string{"https://example.com/feed.xml"})
      xp := xai.XSourceWithArgs([]string{"xai"}, nil, nil, nil)
      search2 := xai.SearchParametersWithDateRange(
          xai.SearchModeOn, true, 10, &from, nil, web, rss, xp,
      )

      // Use search2 in another chat if desired
      _ = search2

      // Example 3: Regular SearchParameters (no date range)
      reg := xai.SearchParameters(xai.SearchModeOn, true, 5, web, rss, xp)
      _ = reg

      resp, err := chat.Sample(context.Background())
      if err != nil {
          log.Fatal(err)
      }
      if len(resp.GetChoices()) > 0 {
          fmt.Println(resp.GetChoices()[0].GetMessage().GetContent())
      }
  }
  ```

## Environment overrides
- XAI_API_HOST: override the endpoint (default: api.x.ai:443)
- XAI_TIMEOUT: per-RPC timeout (e.g., 30s, 2m). Default: 60s
- XAI_INSECURE: set to true/1/yes to disable TLS (local/dev only)

## Notes
- TLS is enabled by default and connections go to port 443.
- For advanced control (custom host, metadata, timeout, insecure), use NewClientWithOptions.
- Each RPC includes your API key via the "x-api-key" header; set XAI_INSECURE=true only for local/dev without TLS.
