package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/bitly/go-simplejson"
	"github.com/stretchr/testify/require"
)

const (
	bitlyAPIBase   = "https://api-ssl.bitly.com/v4"
	bitlyShortenEP = "/shorten"
	bitlyToken     = "9d2b7f4ae41b97b16443090d7997cb6e84a667f8"
)

func TestBitlySupportForScripts(t *testing.T) {
	data, err := ioutil.ReadFile("urls.txt")
	require.NoError(t, err)

	lines := strings.Split(string(data), "\n")

	o, err := os.Create("urls_output.txt")
	require.NoError(t, err)
	defer o.Close()

	for _, url := range lines {
		o.WriteString(url)

		// skip non urls
		if strings.HasPrefix(url, "#") || url == "" {
			o.WriteString("\n")
			continue
		}

		// mark emails
		if strings.Contains(url, "@") {
			o.WriteString(" 📫\n")
			continue
		}

		formattedURL := fmt.Sprintf("http://%s/", url)

		shortenURLResp, err := shortenURL(formattedURL)
		if err != nil {
			o.WriteString(" ERROR: " + err.Error() + " ❌\n")
			continue
		}

		longURL1 := shortenURLResp.Get("long_url").MustString()

		getBitlinkResp, err := getBitlink(shortenURLResp.Get("link").MustString())
		if err != nil {
			o.WriteString(" ERROR: " + err.Error() + " ❌\n")
			continue
		}

		longURL2 := getBitlinkResp.Get("long_url").MustString()
		if longURL1 == formattedURL && longURL2 == formattedURL {
			o.WriteString(" ✅")
		} else {
			o.WriteString(" ❌" + longURL1 + " " + longURL2)
		}

		// get the redirect destination
		destination, err := FetchRedirectDestination(shortenURLResp.Get("link").MustString())
		if err != nil {
			o.WriteString(" ERROR: " + err.Error() + " ❌\n")
			continue
		}

		if destination == formattedURL {
			o.WriteString(" ✅")
		} else {
			o.WriteString(" ❌ " + destination)
		}

		o.WriteString("\n")
	}
}

// FetchRedirectDestination takes a shortened Bitly URL, follows its redirect,
// and returns the destination URL.
func FetchRedirectDestination(bitlyURL string) (string, error) {
	// Create an HTTP client with redirect policy not to follow redirects.
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

	// Check for the "Location" header in the response
	location := resp.Header.Get("Location")
	if location == "" {
		return "", fmt.Errorf("no redirect location found")
	}
	return location, nil
}

func sendRequest(method, url string, body io.Reader, headers map[string]string) (*simplejson.Json, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return simplejson.NewJson(respBody)
}

func shortenURL(longURL string) (*simplejson.Json, error) {
	if longURL == "" {
		return nil, nil
	}

	js := simplejson.New()
	js.Set("long_url", longURL)

	data, err := js.MarshalJSON()
	if err != nil {
		return nil, err
	}

	headers := map[string]string{
		"Authorization": "Bearer " + bitlyToken,
		"Content-Type":  "application/json",
	}
	return sendRequest("POST", bitlyAPIBase+bitlyShortenEP, bytes.NewReader(data), headers)
}

func getBitlink(shortURL string) (*simplejson.Json, error) {
	shortURL = strings.TrimPrefix(shortURL, "https://")

	headers := map[string]string{
		"Authorization": "Bearer " + bitlyToken,
	}
	return sendRequest("GET", bitlyAPIBase+"/bitlinks/"+shortURL, nil, headers)
}
