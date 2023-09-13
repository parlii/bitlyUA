package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/bitly/go-simplejson"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/idna"
)

const (
	bitlyAPIBase   = "https://api-ssl.bitly.com/v4"
	bitlyShortenEP = "/shorten"
	bitlyToken     = "9d2b7f4ae41b97b16443090d7997cb6e84a667f8" // Update with your Bitly token
)

// URLType is a custom type to describe the type of the URL.
type URLType string

const (
	Encoded  URLType = "Encoded"
	Punycode URLType = "Punycode"
	IDN      URLType = "IDN"
	ASCII    URLType = "ASCII"
)

type URLUATestResult struct {
	OriginalURL     string
	OriginalURLType URLType

	FormattedURL string
	BitlyURL     string

	StoredURL     string
	StoredURLType URLType

	RedirectedDestinationURL     string
	RedirectedDestinationURLType URLType

	Error error
}

func TestBitlySupportForUA(t *testing.T) {
	rawURLs, err := loadURLsFromFile("urls.txt")
	require.NoError(t, err)

	var results []URLUATestResult

	for _, rawURL := range rawURLs {
		result := processURL(rawURL)
		results = append(results, result)
	}

	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile("results.json", data, 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func loadURLsFromFile(filename string) ([]string, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var URLs []string

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "#") {
			continue
		}
		if strings.Contains(line, "@") {
			continue
		}
		URLs = append(URLs, line)
	}

	return URLs, nil
}

func processURL(rawURL string) URLUATestResult {
	result := URLUATestResult{
		OriginalURL:  rawURL,
		FormattedURL: fmt.Sprintf("http://%s/", rawURL),
	}

	var err error

	result.OriginalURLType, err = getURLType(result.OriginalURL)
	if err != nil {
		result.Error = err
		return result
	}

	shortenResp, err := shortenURL(result.FormattedURL)
	if err != nil {
		result.Error = err
		return result
	}

	bitlinkResp, err := getBitlink(shortenResp.Get("link").MustString())
	if err != nil {
		result.Error = err
		return result
	}

	result.StoredURL = bitlinkResp.Get("long_url").MustString()
	result.StoredURLType, err = getURLType(result.StoredURL)
	if err != nil {
		result.Error = err
		return result
	}

	result.BitlyURL = shortenResp.Get("link").MustString()

	redirectedURL, err := fetchRedirectDestination(result.BitlyURL)
	if err != nil {
		result.Error = err
		return result
	}

	result.RedirectedDestinationURL = redirectedURL

	urlType, err := getURLType(redirectedURL)
	if err != nil {
		result.Error = err
		return result
	}
	result.RedirectedDestinationURLType = urlType

	return result
}

// getURLType determines the type of the given URL.
func getURLType(rawURL string) (URLType, error) {
	if !strings.HasPrefix(rawURL, "http") {
		rawURL = "http://" + rawURL
	}

	if !strings.HasSuffix(rawURL, "/") {
		rawURL = rawURL + "/"
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	// Convert domain to punycode and check if it matches the original.
	// If it doesn't match, then it was an IDN domain.
	puny, err := idna.ToASCII(parsedURL.Host)
	if err != nil {
		return "", err
	}

	if puny != parsedURL.Host {
		return IDN, nil
	}

	// Check for punycode (xn--).
	if strings.Contains(parsedURL.Host, "xn--") {
		return Punycode, nil
	}

	// Check for percentage-encoded characters in the URL.
	if strings.Contains(rawURL, "%") {
		return Encoded, nil
	}

	return ASCII, nil
}

func fetchRedirectDestination(bitlyURL string) (string, error) {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Get(bitlyURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	return resp.Header.Get("Location"), nil
}

func sendRequest(method, endpoint string, body io.Reader) (*simplejson.Json, error) {
	req, err := http.NewRequest(method, bitlyAPIBase+endpoint, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+bitlyToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return simplejson.NewJson(b)
}

func shortenURL(longURL string) (*simplejson.Json, error) {
	js := simplejson.New()
	js.Set("long_url", longURL)

	data, err := js.MarshalJSON()
	if err != nil {
		return nil, err
	}

	return sendRequest("POST", bitlyShortenEP, bytes.NewReader(data))
}

func getBitlink(shortURL string) (*simplejson.Json, error) {
	bitlink_id := strings.TrimPrefix(shortURL, "https://")

	return sendRequest("GET", "/bitlinks/"+bitlink_id, nil)
}
