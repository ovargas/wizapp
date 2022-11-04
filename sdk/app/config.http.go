package app

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

// fetchConfiguration makes an HTTP request to fetch configuration from the config server.
func fetchConfiguration(url string) []byte {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("unable to fetch configuration from %s\n", url)
		}
	}()

	resp, err := http.Get(url)
	if err != nil {
		panic("Couldn't load configuration, cannot start. Terminating. Error: " + err.Error())
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Couldn't load configuration. Unable to parse Body. Terminating. Error: %v, %v\n", err, resp.Status)
		return nil
	}

	return body
}
