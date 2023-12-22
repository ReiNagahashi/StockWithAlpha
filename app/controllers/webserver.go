package controllers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"regexp"
	"stock-with-alpha/app/models"
	"stock-with-alpha/config"
	"strconv"
)

var templates = template.Must(template.ParseFiles("app/views/chart.jinja"))

func viewChartHandler(w http.ResponseWriter, r *http.Request){
	err := templates.ExecuteTemplate(w, "chart.jinja", nil)
	if err != nil{
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}


// 以下はキャンドルデータを非同期処理(AJAX)で取ってくるためにJsonにデータフォーマットに変換した上でAJAXがAPIを取りに行かせる

type JSONError struct{
	Error string `json:"error"`
	Code int `json:"code"`
}

// JSON型でエラーを返す
func APIError(w http.ResponseWriter, errMessage string, code int){
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	
	jsonError, err := json.Marshal(JSONError{Error: errMessage, Code: code})
	if err != nil{
		log.Fatal(err)
	}

	w.Write(jsonError)
}


var apiValidPath = regexp.MustCompile("^/api/candle/$")
func apiMakeHandler(fn func(http.ResponseWriter, *http.Request)) http.HandlerFunc{
	return func(w http.ResponseWriter, r *http.Request){
		m := apiValidPath.FindStringSubmatch(r.URL.Path)
		if len(m) == 0{
			APIError(w, "Not found", http.StatusNotFound)
		}
		fn(w, r)
	}
}


func apiCandleHandler(w http.ResponseWriter, r *http.Request){
	// URL.Query関数によってURL上のクエリにアクセスできる。その上Getで特定の項目を指定して取れる
	symbol := r.URL.Query().Get("symbol")
	if symbol == "" {
		// api/candle/ のみのURLにアクセスした場合はこのエラーが返ってくる
		APIError(w, "No symbol param", http.StatusBadRequest)
		return
	}

	strLimit := r.URL.Query().Get("limit")
	limit, err := strconv.Atoi(strLimit)
	if strLimit == "" || err != nil || limit < 0 || limit > 1000{
		limit = 1000
	}
	duration := r.URL.Query().Get("duration")
	if duration == ""{
		duration = "day"
	}
	durationTime := config.Config.Durations[duration]
	df, _ := models.GetAllCandle(symbol, durationTime, limit)

	sma := r.URL.Query().Get("sma")
	if sma != ""{
		strSmaPeriod1 := r.URL.Query().Get("smaPeriod1")
		strSmaPeriod2 := r.URL.Query().Get("smaPeriod2")
		strSmaPeriod3 := r.URL.Query().Get("smaPeriod3")

		period1, err := strconv.Atoi(strSmaPeriod1)
		if strSmaPeriod1 == "" || err != nil || period1 < 0{
			period1 = 7
		}

		period2, err := strconv.Atoi(strSmaPeriod2)
		if strSmaPeriod2 == "" || err != nil || period2 < 0{
			period2 = 14
		}

		period3, err := strconv.Atoi(strSmaPeriod3)
		if strSmaPeriod3 == "" || err != nil || period3 < 0{
			period3 = 50
		}

		df.AddSma(period1)
		df.AddSma(period2)
		df.AddSma(period3)

	}

	if r.URL.Query().Get("ema") != ""{
		emaValuePeriod1 := r.URL.Query().Get("emaPeriod1")
		emaValuePeriod2 := r.URL.Query().Get("emaPeriod2")
		emaValuePeriod3 := r.URL.Query().Get("emaPeriod3")


		period1, err := strconv.Atoi(emaValuePeriod1)
		if emaValuePeriod1 == "" || err != nil || period1 < 0{
			period1 = 7
		}

		period2, err := strconv.Atoi(emaValuePeriod2)
		if emaValuePeriod2 == "" || err != nil || period2 < 0{
			period2 = 14
		}

		period3, err := strconv.Atoi(emaValuePeriod3)
		if emaValuePeriod3 == "" || err != nil || period3 < 0{
			period3 = 50
		}

		df.AddEma(period1)
		df.AddEma(period2)
		df.AddEma(period3)

	}
	if r.URL.Query().Get("bbands") != ""{
		strN := r.URL.Query().Get("bbandsN")
		strK := r.URL.Query().Get("bbandsK")

		N,err := strconv.Atoi(strN)
		if strN == "" || err != nil || N < 0{
			N = 20
		}

		K, err := strconv.Atoi(strK)
		if strK == "" || err != nil || K < 0{
			K = 2
		}

		df.AddBBands(N, float64(K))
	}

	if r.URL.Query().Get("rsi") != ""{
		periodStr := r.URL.Query().Get("rsiPeriod")
		period,err := strconv.Atoi(periodStr)

		if periodStr == "" || err != nil || period < 0{
			period = 14
		}

		df.AddRsi(period)
	}

	if r.URL.Query().Get("events") != ""{
		if config.Config.BackTest{
			performance, period, buyThread, sellThread := df.OptimizeRsi()
			fmt.Println(performance)
			if performance > 0{
				fmt.Println(period, performance)
				df.Events = df.BacktestRsi(period, buyThread, sellThread)
			}
		}else{
			firstTime := df.Candles[0].DateTime
			df.AddEvents(firstTime)
		}
	}

	js, err := json.Marshal(df)
	if err != nil{
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func StartWebServer() error {
	http.HandleFunc("/api/candle/", apiMakeHandler(apiCandleHandler))
	http.HandleFunc("/chart/", viewChartHandler)
	return http.ListenAndServe(fmt.Sprintf(":%d", config.Config.Port), nil)
}

