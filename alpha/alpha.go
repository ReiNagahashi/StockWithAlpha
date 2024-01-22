package alpha

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"stock-with-alpha/app/models"
	"stock-with-alpha/config"
	"strconv"
	"strings"
	"time"
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

	urlValues := url.Values{}
	for key, val := range query{
		urlValues.Add(key, val)
	}
	
	if len(query) > 0{
		urlPath = urlPath + "query?" + urlValues.Encode()
	}

	apiURL, err := url.Parse(urlPath)
	if err != nil{
		return
	}

	endpoint := baseURL.ResolveReference(apiURL).String()
	fmt.Printf("action=doRequest, endpoint=%s", endpoint)

	req, err := http.NewRequest(method, endpoint, bytes.NewBuffer(data))
	if err != nil{
		return
	}
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

// Meta data for Weekly
type MetaDataWeekly struct {
	Information   string `json:"1. Information"`
	Symbol        string `json:"2. Symbol"`
	LastRefreshed string `json:"3. Last Refreshed"`
	TimeZone      string `json:"4. Time Zone"`
	
}

// Meta data for Daily
type MetaDataDaily struct {
	Information   string `json:"1. Information"`
	Symbol        string `json:"2. Symbol"`
	LastRefreshed string `json:"3. Last Refreshed"`
	OutputSize	string 	 `json:"4. Output Size"`
	TimeZone      string `json:"5. Time Zone"`
}


// TimeSeriesEntry はAPIから取得したデータの1エントリを表す構造体です。
type TimeSeriesEntry struct {
	Open   string `json:"1. open"`
	High   string `json:"2. high"`
	Low    string `json:"3. low"`
	Close  string `json:"4. close"`
	Volume string `json:"5. volume"`
}

type Ticker struct {
	Symbol string
	Date  string          // 日付を格納するフィールド
	Data  TimeSeriesEntry // 日付に対応するデータを格納するフィールド
}

// Response はAPIのレスポンス全体を表す構造体です。
type ResponseWeekly struct {
	MetaDataWeekly         MetaDataWeekly         			`json:"Meta Data"`
	Ticker 			 map[string]TimeSeriesEntry `json:"Weekly Time Series"`
}

// Response Daily
type ResponseDaily struct {
	MetaDataDaily         MetaDataDaily         			`json:"Meta Data"`
	Ticker 			 map[string]TimeSeriesEntry `json:"Time Series (Daily)"`
}

func (api *APIClient) GetDailyTicker(symbol, name, f string, duration time.Duration, ch chan <- error) {
	resp, err := api.doRequest("GET", "", map[string]string{"symbol": symbol, "function": f, "apikey": config.Config.ApiKey}, nil)
	if err != nil{
		ch <- err
	}
	
	var response = ResponseDaily{}
	
	if err := json.Unmarshal([]byte(resp), &response); err != nil {
		ch <- err
	}

	for date, data := range response.Ticker {
		symbol := response.MetaDataDaily.Symbol
		name := name
		open,_ := strconv.ParseFloat(data.Open, 64)
		close,_ := strconv.ParseFloat(data.Close, 64)
		high, _ := strconv.ParseFloat(data.High, 64)
		low, _ := strconv.ParseFloat(data.Low, 64)
		volume, _ := strconv.ParseFloat(data.Volume, 64)
		datePart := strings.Split(date, " ")[0]

		parsedDate, err := time.Parse("2006-01-02", datePart)
		if err != nil {
			ch <- err
		}
		
		candle := models.Candle{
			Symbol: symbol,
			Name: name,
			Duration: duration,
			Open: open,
			Close: close,
			High: high,
			Low: low,
			Volume: volume,
			DateTime: parsedDate,
		}

		models.CreateCandleWithDuration(candle, symbol, name, parsedDate, duration)

	}
}



// func (api *APIClient) GetWeeklyTicker(symbol, f string, duration time.Duration) error{
// 	resp, err := api.doRequest("GET", "", map[string]string{"symbol": symbol, "function": f, "apikey": config.Config.ApiKey}, nil)
// 	if err != nil{
// 		return err
// 	}
	
// 	var response = ResponseWeekly{}
	
// 	if err := json.Unmarshal([]byte(resp), &response); err != nil {
// 		return err
// 	}

// 	for date, data := range response.Ticker {
// 		symbol := response.MetaDataWeekly.Symbol
// 		open,_ := strconv.ParseFloat(data.Open, 64)
// 		close,_ := strconv.ParseFloat(data.Close, 64)
// 		high, _ := strconv.ParseFloat(data.High, 64)
// 		low, _ := strconv.ParseFloat(data.Low, 64)
// 		volume, _ := strconv.ParseFloat(data.Volume, 64)
// 		datePart := strings.Split(date, " ")[0]

// 		parsedDate, err := time.Parse("2006-01-02", datePart)
// 		if err != nil {
// 			log.Println("Error parsing date:", err)
// 		}

// 		candle := models.Candle{
// 			Symbol: symbol,
// 			Duration: duration,
// 			Open: open,
// 			Close: close,
// 			High: high,
// 			Low: low,
// 			Volume: volume,
// 			DateTime: parsedDate,
// 		}


// 		models.CreateCandleWithDuration(candle, symbol, name, parsedDate, duration)

// 	}
	
// 	return nil
// }


type TickerByKeyword struct {
	BestMatches []struct {
		Symbol      string `json:"1. symbol"`
		Name        string `json:"2. name"`
		Type      string `json:"3. type"`
		Region     string `json:"4. region"`
		MarketOpen string `json:"5. marketOpen"`
		MarketClose string `json:"6. marketClose"`
		Timezone  string `json:"7. timezone"`
		Currency  string `json:"8. currency"`
		MatchScore string `json:"9. matchScore"`
	} `json:"bestMatches"`
}


func (api *APIClient) GetTickerInfo(f, keyword string) (*TickerByKeyword, error){
	resp, err := api.doRequest("GET", "", map[string]string{"function": f, "keywords": keyword, "apikey": config.Config.ApiKey}, nil)
	if err != nil{
		return nil, err
	}
	
	var response = TickerByKeyword{}

	if err := json.Unmarshal([]byte(resp), &response); err != nil {
		return nil, err
	}

	return &response, nil

}
