package line

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

// ReceivedMessage is a body of callback sent by LINE
type ReceivedMessage struct {
	Results []Result `json:"result"`
}

// Result is a message or a operation.
type Result struct {
	ID          string   `json:"id"`
	From        string   `json:"from"`
	FromChannel int      `json:"fromChannel"`
	To          []string `json:"to"`
	ToChannel   int      `json:"toChannel"`
	EventType   string   `json:"eventType"`
	Content     Content  `json:"content"`
}

// SendingMessage is a body for Posting API
type SendingMessage struct {
	To        []string `json:"to"`
	ToChannel int      `json:"toChannel"`
	EventType string   `json:"eventType"`
	Content   Content  `json:"content"`
}

// Content is a content of a message
type Content struct {
	ID             string          `json:"id"`
	ContentType    int             `json:"contentType"`
	From           string          `json:"from"`
	CreatedTime    int             `json:"createdTime"`
	To             []string        `json:"to"`
	ToType         int             `json:"toType"`
	ConentMetadata ContentMetadata `json:"contentMetadata"`
	Text           string          `json:"text"`
	Location       Location        `json:"location"`
}

// ContentType constants
const (
	ContentTypeText int = 1 + iota
	ContentTypeImage
	ContentTypeVideo
	ContentTypeAudio
	ContentTypeUndefined5
	ContentTypeUndefined6
	ContentTypeLocation
	ContentTypeSticker
	ContentTypeUndefined9
	ContentTypeContact
)

// Location is a location related data
type Location struct {
	Title     string `json:"title"`
	Address   string `json:"address"`
	Latitude  int    `json:"latitude"`
	Longitude int    `json:"Longitude"`
}

// ContentMetadata is a metadata of content
type ContentMetadata struct {
	STKPKGID    string `json:"STKPKGID"`
	STKID       string `json:"STKID"`
	STKVER      string `json:"STKVER"`
	STKTXT      string `json:"STKTXT"`
	Mid         string `json:"mid"`
	DisplayName string `json:"displayName"`
}

// Response is a body returned by LINE
type Response struct {
	Failed    []interface{} `json:"failed"`
	MessageID string        `json:"messageId"`
	Timestamp float64       `json:"timestamp"`
	Version   int           `json:"version"`
}

const (
	baseEndPoint     = "https://trialbot-api.line.me"
	defaultToChannel = 1383378250
	defaultEventType = "138311608800106203"
)

// API represents LINE BOT API
type API struct {
	channelID     string
	channelSecret string
	mID           string
	proxyURL      *url.URL
	Logger        *log.Logger
	Debug         bool
}

// New creates an API instance
func New(channelID string, channelSecret string, mid string) *API {
	return &API{
		channelID:     channelID,
		channelSecret: channelSecret,
		mID:           mid,
		Logger:        log.New(os.Stderr, "line-bot-api", 0),
		Debug:         false,
	}
}

// DecodeMessage decodes body to ReceivedMessage struct.
func (api *API) DecodeMessage(body io.Reader) (*ReceivedMessage, error) {
	var m ReceivedMessage
	b, err := ioutil.ReadAll(body)
	if err != nil {
		return nil, err
	}
	if api.Debug {
		api.Logger.Println("RecievedMessage Body:", string(b))
	}
	err = json.Unmarshal(b, &m)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// SetProxy set Proxy from URL string.
func (api *API) SetProxy(proxyURL string) error {
	var err error
	api.proxyURL, err = url.Parse(proxyURL)
	return err
}

// SendMessage sends a text to LINE
func (api *API) SendMessage(to []string, text string) error {
	m := &SendingMessage{
		To:        to,
		ToChannel: defaultToChannel,
		EventType: defaultEventType,
		Content: Content{
			ContentType: 1,
			ToType:      1,
			Text:        text,
		},
	}
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}
	if len(b) > 8192 {
		return errors.New("The size of API request is too long (over 8192 bytes).")
	}
	req, err := http.NewRequest("POST", baseEndPoint+"/v1/events", bytes.NewBuffer(b))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json; charset=UTF-8")
	req.Header.Add("X-Line-ChannelID", api.channelID)
	req.Header.Add("X-Line-ChannelSecret", api.channelSecret)
	req.Header.Add("X-Line-Trusted-User-With-ACL", api.mID)

	client := &http.Client{
		Timeout: time.Duration(30 * time.Second),
	}
	if api.proxyURL != nil {
		client.Transport = &http.Transport{Proxy: http.ProxyURL(api.proxyURL)}
	}

	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return err
	}
	api.Logger.Print("Response: ", result)
	return nil
}
