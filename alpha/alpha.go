package alpha

import (
	"bytes"
	"encoding/json"
	// "fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"stock-with-alpha/config"
)

const baseURL = "https://www.alphavantage.co/"

type APIClient struct{
	key 		string
	httpClient 	*http.Client
}


func New(key string) *APIClient{
	apiClient := &APIClient{key, &http.Client{}}
	return apiClient
}


// リクエストをwebサーバに投げる
func (api *APIClient) doRequest(method, urlPath string, query map[string]string, data []byte) (body []byte, err error){
	baseURL, err := url.Parse(baseURL)
	if err != nil{
		return 
	}

	if len(query) > 0{
		urlPath += "query?"
		for key, val := range query{
			urlPath += key + "=" + val + "&"
		}
		urlPath = urlPath[:len(urlPath) - 1]
	}

	apiURL, err := url.Parse(urlPath)
	if err != nil{
		return
	}

	endpoint := baseURL.ResolveReference(apiURL).String()
	log.Printf("action=doRequest, endpoint=%s", endpoint)

	req, err := http.NewRequest(method, endpoint, bytes.NewBuffer(data))
	if err != nil{
		return
	}

	// q := req.URL.Query()
	// for key, val := range query{
	// 	q.Add(key, val)
	// }
	req.URL.RawQuery = req.URL.Query().Encode()

	resp, err := api.httpClient.Do(req)
	if err != nil{
		return nil, err
	}
	defer resp.Body.Close()
	body, err = io.ReadAll(resp.Body)
	if err != nil{
		return nil, err
	}

	return body, nil
}


	
type Ticker struct {
	GlobalQuote struct {
		Symbol           string `json:"01. symbol"`
		Open             string `json:"02. open"`
		High             string `json:"03. high"`
		Low              string `json:"04. low"`
		Price            string `json:"05. price"`
		Volume           string `json:"06. volume"`
		LatestTradingDay string `json:"07. latest trading day"`
		PreviousClose    string `json:"08. previous close"`
		Change           string `json:"09. change"`
		ChangePercent    string `json:"10. change percent"`
	} `json:"Global Quote"`
}


func (api *APIClient) GetTicker(symbol, f string) (*Ticker, error){
	resp, err := api.doRequest("GET", "", map[string]string{"symbol": symbol, "function": f, "apikey": config.Config.ApiKey}, nil)
	if err != nil{
		return nil, err
	}

	var ticker Ticker
	err = json.Unmarshal(resp, &ticker)
	if err != nil{
		return nil, err
	}

	return &ticker, nil
}