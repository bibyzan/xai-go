package xai

import (
	"context"
	"time"

	v1 "github.com/bibyzan/xai-proto/gen/go/github.com/bibyzan/xai-proto/gen/go/xai/api/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Wrapper enum values to avoid importing proto enums in user code.
// These mirror v1.SearchMode values.
// Use: xai.SearchModeOn, xai.SearchModeOff, xai.SearchModeAuto.
const (
	SearchModeOff  = v1.SearchMode_OFF_SEARCH_MODE
	SearchModeOn   = v1.SearchMode_ON_SEARCH_MODE
	SearchModeAuto = v1.SearchMode_AUTO_SEARCH_MODE
)

// Chat is a small convenience wrapper mirroring the Python-style API where you
// build a request (chat = client.chat.create(...)) and then call chat.sample().
// In Go, Sample simply invokes the synchronous GetCompletion RPC.
type Chat struct {
	client v1.ChatClient
	req    *v1.GetCompletionsRequest
}

// ChatCreate constructs a Chat handle from model, messages, and optional search parameters.
// Use Chat.Sample(ctx) to synchronously fetch the completion.
func (c *Client) ChatCreate(model string, messages []*v1.Message, search *v1.SearchParameters) *Chat {
	return &Chat{
		client: c.chat,
		req: &v1.GetCompletionsRequest{
			Model:            model,
			Messages:         messages,
			SearchParameters: search,
		},
	}
}

// ChatCreateWithPrompt is a convenience helper that builds a single user message from prompt.
func (c *Client) ChatCreateWithPrompt(model, prompt string, search *v1.SearchParameters) *Chat {
	return c.ChatCreate(model, []*v1.Message{User(prompt)}, search)
}

// Sample executes the chat completion request and returns the response.
func (ch *Chat) Sample(ctx context.Context) (*v1.GetChatCompletionResponse, error) {
	return ch.client.GetCompletion(ctx, ch.req)
}

// User creates a user-role message with a single text content.
func User(text string) *v1.Message {
	return &v1.Message{
		Role: v1.MessageRole_ROLE_USER,
		Content: []*v1.Content{
			{Content: &v1.Content_Text{Text: text}},
		},
	}
}

// ContentText creates a content wrapper for plain text (useful for advanced message assembly).
func ContentText(text string) *v1.Content { return &v1.Content{Content: &v1.Content_Text{Text: text}} }

// XSource returns a Source configured for X (formerly Twitter) with no filters.
func XSource() *v1.Source { return &v1.Source{Source: &v1.Source_X{X: &v1.XSource{}}} }

// XSourceWithArgs configures an X source with included/excluded handles and optional thresholds.
// Pass nil for postFavoriteCount or postViewCount to leave them unset.
func XSourceWithArgs(includedHandles, excludedHandles []string, postFavoriteCount, postViewCount *int32) *v1.Source {
	sx := &v1.XSource{
		IncludedXHandles: includedHandles,
		ExcludedXHandles: excludedHandles,
	}
	if postFavoriteCount != nil {
		sx.PostFavoriteCount = postFavoriteCount
	}
	if postViewCount != nil {
		sx.PostViewCount = postViewCount
	}
	return &v1.Source{Source: &v1.Source_X{X: sx}}
}

// WebSource creates a web search source with allowed/excluded sites, optional country, and safe search.
// Pass country as empty string to leave it unset.
func WebSource(excludedWebsites, allowedWebsites []string, country string, safeSearch bool) *v1.Source {
	ws := &v1.WebSource{
		ExcludedWebsites: excludedWebsites,
		AllowedWebsites:  allowedWebsites,
		SafeSearch:       safeSearch,
	}
	if country != "" {
		ws.Country = &country
	}
	return &v1.Source{Source: &v1.Source_Web{Web: ws}}
}

// NewsSource creates a news search source with excluded websites, optional country, and safe search.
func NewsSource(excludedWebsites []string, country string, safeSearch bool) *v1.Source {
	ns := &v1.NewsSource{
		ExcludedWebsites: excludedWebsites,
		SafeSearch:       safeSearch,
	}
	if country != "" {
		ns.Country = &country
	}
	return &v1.Source{Source: &v1.Source_News{News: ns}}
}

// RssSource creates an RSS source with the provided feed links.
func RssSource(links []string) *v1.Source { return &v1.Source{Source: &v1.Source_Rss{Rss: &v1.RssSource{Links: links}}} }

// SearchParameters builds SearchParameters with explicit mode, sources, and flags.
// If maxSearchResults <= 0, it will be left unset. Provide any number of sources.
func SearchParameters(mode v1.SearchMode, returnCitations bool, maxSearchResults int32, sources ...*v1.Source) *v1.SearchParameters {
	sp := &v1.SearchParameters{
		Mode:            mode,
		Sources:         sources,
		ReturnCitations: returnCitations,
	}
	if maxSearchResults > 0 {
		// take address of local copy; it will escape to heap as needed
		sp.MaxSearchResults = &maxSearchResults
	}
	return sp
}

// SearchParametersWithDateRange builds SearchParameters and passes through optional from/to date range.
// Provide nil for from or to to leave them unset.
func SearchParametersWithDateRange(mode v1.SearchMode, returnCitations bool, maxSearchResults int32, from, to *time.Time, sources ...*v1.Source) *v1.SearchParameters {
	sp := SearchParameters(mode, returnCitations, maxSearchResults, sources...)
	if from != nil {
		sp.FromDate = timestamppb.New(*from)
	}
	if to != nil {
		sp.ToDate = timestamppb.New(*to)
	}
	return sp
}

// SearchParametersXOn builds SearchParameters similar to the Python snippet:
//   sources=[x_source()], mode="on", return_citations=true, max_search_results=3
func SearchParametersXOn(returnCitations bool, maxSearchResults int32) *v1.SearchParameters {
	return &v1.SearchParameters{
		Mode:             v1.SearchMode_ON_SEARCH_MODE,
		Sources:          []*v1.Source{XSource()},
		ReturnCitations:  returnCitations,
		MaxSearchResults: &maxSearchResults,
	}
}
