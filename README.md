# xai-go

Go client for X.ai gRPC APIs.

Defaults
- Endpoint: api.x.ai:443 (TLS)
- Only the API key is required to construct a client
- Per-RPC timeout default: 60s

Quick start

- Install:
  go get github.com/bibyzan/xai

- Use:
  package main

  import (
      "context"
      "fmt"
      "log"
      "os"
      xai "github.com/bibyzan/xai"
  )

  func main() {
      apiKey := os.Getenv("XAI_API_KEY")
      if apiKey == "" {
          log.Fatal("set XAI_API_KEY")
      }

      client := xai.NewClient(apiKey)
      defer client.Close()

      // Python-equivalent flow:
      // search_config = SearchParameters(sources=[x_source()], mode="on", return_citations=True, max_search_results=3)
      // chat = client.chat.create(model=model, messages=[user(prompt)], search_parameters=search_config)
      // response = await asyncio.to_thread(chat.sample)

      // Example 1: X source only (like Python example)
      search := xai.SearchParametersXOn(true, 3)
      chat := client.ChatCreateWithPrompt("grok-2", "Explain superconductors like I am five.", search)

      // Example 2: Mixed sources with passthrough arguments and date range
      max := int32(10)
      from := time.Now().AddDate(0, -1, 0) // last month
      web := xai.WebSource(nil, []string{"wikipedia.org", "x.ai"}, "", true)
      rss := xai.RssSource([]string{"https://example.com/feed.xml"})
      xp := xai.XSourceWithArgs([]string{"xai"}, nil, nil, nil)
      search2 := xai.SearchParametersWithDateRange(
          v1.SearchMode_ON_SEARCH_MODE, true, &max, &from, nil, web, rss, xp,
      )
      _ = search2 // use in another chat if desired

      resp, err := chat.Sample(context.Background())
      if err != nil {
          log.Fatal(err)
      }
      if len(resp.GetChoices()) > 0 {
          fmt.Println(resp.GetChoices()[0].GetMessage().GetContent())
      }
  }

Environment overrides
- XAI_API_HOST: override the endpoint (default: api.x.ai:443)
- XAI_TIMEOUT: per-RPC timeout (e.g., 30s, 2m). Default: 60s
- XAI_INSECURE: set to true/1/yes to disable TLS (local/dev only)

Notes
- TLS is enabled by default and connections go to port 443.
- For advanced control (custom host, metadata, timeout, insecure), use NewClientWithOptions.
