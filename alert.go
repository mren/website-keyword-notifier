package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// AlertConfig contains the configuration for one keyword.
type AlertConfig struct {
	Email   string
	Keyword string
	URL     string
}

func main() {
	mailgunKey := os.Getenv("MAILGUN_KEY")
	mailgunDomain := os.Getenv("MAILGUN_DOMAIN")
	configURL := os.Getenv("CONFIG_URL")

	if len(mailgunKey) == 0 {
		fmt.Println("Missing MAILGUN_KEY environment variable.")
		return
	}
	if len(mailgunDomain) == 0 {
		fmt.Println("Missing MAILGUN_DOMAIN environment variable.")
		return
	}
	if len(configURL) == 0 {
		fmt.Println("Missing CONFIG_URL environment variable.")
		return
	}

	alertConfigs, err := getAlertConfigs(configURL)

	if err != nil {
		fmt.Println(err)
		return
	}

	for _, config := range alertConfigs {
		isMatch, err := urlContainsKeyword(config.URL, config.Keyword)
		if err != nil {
			fmt.Println(err)
			return
		}
		if isMatch {
			subject := fmt.Sprintf("Keyword %s found on %s.",
				config.Keyword, config.URL)
			body := fmt.Sprintf("On %s the keyword \"%s\" was found.\nConfig: %s.",
				config.URL, config.Keyword, configURL)
			err = sendEmail(mailgunKey, mailgunDomain, config.Email, subject, body)

			if err != nil {
				fmt.Println(err)
				return
			}
		}
	}
}

func getAlertConfigs(path string) ([]AlertConfig, error) {
	parsedURL, err := url.Parse(path)

	if err != nil {
		return nil, err
	}

	var jsonBlob []byte

	if parsedURL.Scheme == "file" {
		localPath := strings.Replace(path, "file://", "", 1)
		jsonBlob, err = ioutil.ReadFile(localPath)
	} else {
		response, err := http.Get(path)
		if err != nil {
			return nil, err
		}
		jsonBlob, err = ioutil.ReadAll(response.Body)
	}

	if err != nil {
		return nil, err
	}

	var alertConfigs []AlertConfig
	err = json.Unmarshal(jsonBlob, &alertConfigs)
	return alertConfigs, nil
}

func urlContainsKeyword(url string, keyword string) (bool, error) {
	response, err := http.Get(url)
	if err != nil {
		return false, err
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return false, err
	}
	bodyStr := strings.ToLower(string(body))
	return strings.Contains(bodyStr, strings.ToLower(keyword)), nil
}

func sendEmail(mailgunKey string, domain string, to string, subject string, text string) error {
	urlStr := "https://api.mailgun.net/v3/" + domain + "/messages"
	client := &http.Client{}
	data := url.Values{}
	data.Set("from", "no-reply@"+domain)
	data.Set("to", to)
	data.Set("subject", subject)
	data.Set("text", text)

	body := bytes.NewBufferString(data.Encode())
	request, err := http.NewRequest("POST", urlStr, body)
	if err != nil {
		return err
	}
	request.SetBasicAuth("api", mailgunKey)

	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	response, err := client.Do(request)
	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}
	if response.Status != "200" {
		return errors.New(string(responseBody))
	}
	fmt.Println("mail send", string(responseBody))
	return nil
}
