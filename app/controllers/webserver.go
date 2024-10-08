package controllers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"regexp"
	"stock-with-alpha/alpha"
	"stock-with-alpha/app/models"
	"stock-with-alpha/config"
	"stock-with-alpha/utils"
	"strconv"
	"sync"
)

var templates = template.Must(template.ParseFiles("app/views/home.html", "app/views/company_detail.html", "app/views/chart.html"))

func makeViewHandler(templateName string) http.HandlerFunc{
	return func(w http.ResponseWriter, r *http.Request){
		err := templates.ExecuteTemplate(w, templateName, nil)
		if err != nil{
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}


// 以下はキャンドルデータを非同期処理(AJAX)で取ってくるためにJsonにデータフォーマットに変換した上でAJAXがAPIを取りに行かせる
type JSONError struct {
	Error string `json:"error"`
	Code  int    `json:"code"`
}

// JSON型でエラーを返す
func APIError(w http.ResponseWriter, errMessage string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	jsonError, err := json.Marshal(JSONError{Error: errMessage, Code: code})
	if err != nil {
		log.Fatal(err)
	}

	w.Write(jsonError)
}

//apiMakeHandlerを汎用関数として作ることで、全てのAPIハンドラーにおいてこの関数を使って登録することができる
func apiMakeHandler(validPath *regexp.Regexp, fn func(http.ResponseWriter, *http.Request)) http.HandlerFunc{
	return func(w http.ResponseWriter, r *http.Request){
		m := validPath.FindStringSubmatch(r.URL.Path)
		if len(m) == 0{
			APIError(w, "Not fount", http.StatusNotFound)
			return
		}
		fn(w, r)
	}
}

// キャンドル送信用のAPIの正規表現
var candleApiValidPath = regexp.MustCompile("^/api/candle/$")


func apiCandleHandler(w http.ResponseWriter, r *http.Request) {
	// URL.Query関数によってURL上のクエリにアクセスできる。その上Getで特定の項目を指定して取れる
	symbol := r.URL.Query().Get("symbol")
	name := r.URL.Query().Get("name")
	if symbol == "" || name == "" {
		// api/candle/ のみのURLにアクセスした場合はこのエラーが返ってくる
		APIError(w, "No symbol or name param", http.StatusBadRequest)
		return
	}

	strLimit := r.URL.Query().Get("limit")
	limit, err := strconv.Atoi(strLimit)
	if strLimit == "" || err != nil || limit < 0 || limit > 1000 {
		limit = 1000
	}
	duration := r.URL.Query().Get("duration")
	if duration == "" {
		duration = "day"
	}
	durationTime := config.Config.Durations[duration]
	df, _ := models.GetAllCandle(symbol, name, durationTime, limit)

	sma := r.URL.Query().Get("sma")
	if sma != "" {
		strSmaPeriod1 := r.URL.Query().Get("smaPeriod1")
		strSmaPeriod2 := r.URL.Query().Get("smaPeriod2")
		strSmaPeriod3 := r.URL.Query().Get("smaPeriod3")

		period1, err := strconv.Atoi(strSmaPeriod1)
		if strSmaPeriod1 == "" || err != nil || period1 < 0 {
			period1 = 7
		}

		period2, err := strconv.Atoi(strSmaPeriod2)
		if strSmaPeriod2 == "" || err != nil || period2 < 0 {
			period2 = 14
		}

		period3, err := strconv.Atoi(strSmaPeriod3)
		if strSmaPeriod3 == "" || err != nil || period3 < 0 {
			period3 = 50
		}

		df.AddSma(period1)
		df.AddSma(period2)
		df.AddSma(period3)

	}

	if r.URL.Query().Get("ema") != "" {
		emaValuePeriod1 := r.URL.Query().Get("emaPeriod1")
		emaValuePeriod2 := r.URL.Query().Get("emaPeriod2")
		emaValuePeriod3 := r.URL.Query().Get("emaPeriod3")

		period1, err := strconv.Atoi(emaValuePeriod1)
		if emaValuePeriod1 == "" || err != nil || period1 < 0 {
			period1 = 7
		}

		period2, err := strconv.Atoi(emaValuePeriod2)
		if emaValuePeriod2 == "" || err != nil || period2 < 0 {
			period2 = 14
		}

		period3, err := strconv.Atoi(emaValuePeriod3)
		if emaValuePeriod3 == "" || err != nil || period3 < 0 {
			period3 = 50
		}

		df.AddEma(period1)
		df.AddEma(period2)
		df.AddEma(period3)

	}
	if r.URL.Query().Get("bbands") != "" {
		strN := r.URL.Query().Get("bbandsN")
		strK := r.URL.Query().Get("bbandsK")

		N, err := strconv.Atoi(strN)
		if strN == "" || err != nil || N < 0 {
			N = 20
		}

		K, err := strconv.Atoi(strK)
		if strK == "" || err != nil || K < 0 {
			K = 2
		}

		df.AddBBands(N, float64(K))
	}

	if r.URL.Query().Get("rsi") != "" {
		periodStr := r.URL.Query().Get("rsiPeriod")
		period, err := strconv.Atoi(periodStr)

		if periodStr == "" || err != nil || period < 0 {
			period = 14
		}

		df.AddRsi(period)
	}

	if r.URL.Query().Get("events") != "" {
		if config.Config.BackTest {
			df.Events = Ai.SignalEvents[df.Symbol]
		} else {
			firstTime := df.Candles[0].DateTime
			df.AddEvents(firstTime)
		}
	}

	js, err := json.Marshal(df)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

// ティッカー検索
var tickerSearchApiValidPath = regexp.MustCompile("^/api/ticker_search/?$")

func tickerSearchHandler(w http.ResponseWriter, r *http.Request) {
	keyword := r.URL.Query().Get("keyword")
	if keyword == "" {
		APIError(w, "No keyword!", http.StatusBadRequest)
	}

	apiClient := alpha.New(config.Config.ApiKey)

	tickerByKeyword, err := apiClient.GetTickerInfo("SYMBOL_SEARCH", keyword)
	if err != nil {
		log.Fatalln(err)
	}

	js, err := json.Marshal(tickerByKeyword)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

// ingestion company info and candle: 引数としてのティッカー情報をもとにCompanyFandamental と candleテーブルの両方を作る関数
var ingestionCandleApiValidPath = regexp.MustCompile("^/api/ingestion_candle/$")

func ingestionCandleHandler(w http.ResponseWriter, r *http.Request) {
	symbol := r.URL.Query().Get("symbol")
	name := r.URL.Query().Get("name")
	if symbol == "" || name == "" {
		APIError(w, "No Symbol or Name!", http.StatusBadRequest)
	}

	apiClient := alpha.New(config.Config.ApiKey)

	// symbolの名前でテーブルを作成
	models.CreateTableBySymbol(symbol, name)

	var wg sync.WaitGroup
	wg.Add(2)
	errChan := make(chan error)
	// ティッカーをAPIで持ってきて構造体にする。その後にテーブルにティッカーに基づいて作られたキャンドルデータを挿入
	go func() {
		defer wg.Done()
		apiClient.GetDailyTicker(symbol, name, "TIME_SERIES_DAILY", config.Config.Durations["day"], errChan)
		apiClient.GetCompanyOverview(symbol, errChan)
	}()

	go func() {
		defer wg.Done()
		utils.ErrorHandler(errChan)
	}()

	wg.Wait()

	Ai.Trade(symbol, name)

}

// dropCandleTable: 引数としてのシンボルをもとにテーブルを削除する関数
var dropCandleTableApiValidPath = regexp.MustCompile("^/api/drop_candle_table/$")

func dropCandleTable(w http.ResponseWriter, r *http.Request) {
	symbol := r.URL.Query().Get("symbol")

	if symbol == "" {
		APIError(w, "No Symbol or Name!", http.StatusBadRequest)
	}

	// symbolを基にテーブルを削除
	models.DropTableBySymbol(symbol)
}

// displayCandleTables: 保存しているキャンドルテーブルの名前を返す
var displayCandleTablesApiValidPath = regexp.MustCompile("^/api/display_tables/$")

func displayCandleTablesHandler(w http.ResponseWriter, r *http.Request) {
	// 日足の既存のテーブル名を全て取得
	symbolWithName := models.GetCandleTableNames()
	js, err := json.Marshal(symbolWithName)
	if err != nil{
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

// displaySavedCompanies: 保存しているキャンドルテーブルの名前を返す
var displaySavedCompaniesApiValidPath = regexp.MustCompile("^/api/display_saved_companies/$")

func displaySavedCompaniesHandler(w http.ResponseWriter, r *http.Request) {
	var requestData struct{
		Symbols []string `json:"portfolio_tickers"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil{
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	// キーを業界・値をシンボルにしたマップを作る
	companiesByIndustry := make(map[string][]string)
	companies := models.CompaniesFundamental{}

	for _, symbol := range requestData.Symbols {
		company := models.GetCompanyBySymbol(symbol)
		companiesByIndustry[company.Industry] = append(companiesByIndustry[company.Industry], symbol)
		companies.Companies = append(companies.Companies, *company)
	}

	js, err := json.Marshal(companies)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}


// fetchBS: 引数としてのシンボルをもとに貸借対照表を取ってくる
var fetchBSApiValidPath = regexp.MustCompile("^/api/fetch_bs/$")

func fetchBSHandler(w http.ResponseWriter, r *http.Request) {
	var requestData struct{
		Symbol string `json:"symbol"`
	}

	// リクエストボディの解析
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		// エラーハンドリング: 不正なJSON、またはその他のエラー
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	symbol := requestData.Symbol

	if symbol == "" {
		APIError(w, "No Symbol or Name!", http.StatusBadRequest)
	}

	apiClient := alpha.New(config.Config.ApiKey)

	// 構造体AnnualReports[]を返す
	annualReports := apiClient.GetBS(symbol)
	
	js, err := json.Marshal(annualReports)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}


// fetchPL: 引数としてのシンボルをもとに損益計算書を取ってくる
var fetchPLApiValidPath = regexp.MustCompile("^/api/fetch_pl/$")

func fetchPLHandler(w http.ResponseWriter, r *http.Request) {
	var requestData struct{
		Symbol string `json:"symbol"`
	}

	// リクエストボディの解析
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		// エラーハンドリング: 不正なJSON、またはその他のエラー
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	symbol := requestData.Symbol

	if symbol == "" {
		APIError(w, "No Symbol or Name!", http.StatusBadRequest)
	}

	apiClient := alpha.New(config.Config.ApiKey)

	// 構造体AnnualReports[]を返す
	annualReports := apiClient.GetPL(symbol)
	
	js, err := json.Marshal(annualReports)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}



// fetchCF: 引数としてのシンボルをもとに損益計算書を取ってくる
var fetchCFApiValidPath = regexp.MustCompile("^/api/fetch_cf/$")

func fetchCFHandler(w http.ResponseWriter, r *http.Request) {
	var requestData struct{
		Symbol string `json:"symbol"`
	}

	// リクエストボディの解析
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		// エラーハンドリング: 不正なJSON、またはその他のエラー
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	symbol := requestData.Symbol

	if symbol == "" {
		APIError(w, "No Symbol or Name!", http.StatusBadRequest)
	}

	apiClient := alpha.New(config.Config.ApiKey)

	// 構造体AnnualReports[]を返す
	annualReports := apiClient.GetCF(symbol)
	
	js, err := json.Marshal(annualReports)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}



func StartWebServer() error {
	http.HandleFunc("/api/candle/", apiMakeHandler(candleApiValidPath, apiCandleHandler))
	http.HandleFunc("/api/ticker_search/", apiMakeHandler(tickerSearchApiValidPath, tickerSearchHandler))
	http.HandleFunc("/api/ingestion_candle/", apiMakeHandler(ingestionCandleApiValidPath, ingestionCandleHandler))
	http.HandleFunc("/api/drop_candle_table/", apiMakeHandler(dropCandleTableApiValidPath, dropCandleTable))
	http.HandleFunc("/api/fetch_bs/", apiMakeHandler(fetchBSApiValidPath, fetchBSHandler))
	http.HandleFunc("/api/fetch_pl/", apiMakeHandler(fetchPLApiValidPath, fetchPLHandler))
	http.HandleFunc("/api/fetch_cf/", apiMakeHandler(fetchCFApiValidPath, fetchCFHandler))
	http.HandleFunc("/api/display_tables/", apiMakeHandler(displayCandleTablesApiValidPath, displayCandleTablesHandler))	
	http.HandleFunc("/api/display_saved_companies/", apiMakeHandler(displaySavedCompaniesApiValidPath, displaySavedCompaniesHandler))
	http.HandleFunc("/company_detail/", makeViewHandler("company_detail.html"))
	http.HandleFunc("/chart/", makeViewHandler("chart.html"))
	http.HandleFunc("/", makeViewHandler("home.html"))

	return http.ListenAndServe(fmt.Sprintf(":%d", config.Config.Port), nil)
}
