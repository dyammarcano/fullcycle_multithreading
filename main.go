package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	viaCEPUrl    = "https://viacep.com.br/ws/%s/json"
	brasilAPIUrl = "https://brasilapi.com.br/api/cep/v1/%s"
)

func requestViaCEP(cep string) (string, error) {
	resp, err := requestHelper(fmt.Sprintf(viaCEPUrl, cep))
	if err != nil {
		return "", err
	}

	data := bytes.NewBuffer(nil)
	if _, err = data.ReadFrom(resp.Body); err != nil {
		return "", err
	}

	return data.String(), nil
}

func requestBrasilAPI(cep string) (string, error) {
	resp, err := requestHelper(fmt.Sprintf(brasilAPIUrl, cep))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	data := bytes.NewBuffer(nil)
	if _, err = data.ReadFrom(resp.Body); err != nil {
		return "", err
	}

	return data.String(), nil
}

func requestHelper(url string) (*http.Response, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		var netErr net.Error
		ok := errors.As(err, &netErr)
		if ok && netErr.Timeout() {
			return nil, fmt.Errorf("request to %s timed out", url)
		}
		return nil, err
	}

	return resp, nil
}

func main() {
	cep := "01153000"

	viaCEP := make(chan string)
	brasilAPI := make(chan string)

	go func() {
		result, err := requestViaCEP(cep)
		if err != nil {
			fmt.Println("ViaCEP API error")
			return
		}
		viaCEP <- result
	}()

	go func() {
		result, err := requestBrasilAPI(cep)
		if err != nil {
			fmt.Println("BrasilAPI error")
			return
		}
		brasilAPI <- result
	}()

	select {
	case v := <-viaCEP:
		fmt.Printf("viacep.com.br: %s\n", strings.ReplaceAll(v, "\n", "")) // Remove new line
		os.Exit(0)
	case b := <-brasilAPI:
		fmt.Printf("brasilapi.com.br: %s\n", strings.ReplaceAll(b, "\n", "")) // Remove new line
		os.Exit(0)
	}
}
