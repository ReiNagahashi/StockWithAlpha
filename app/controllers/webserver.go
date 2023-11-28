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

var templates = template.Must(template.ParseFiles("app/views/google.jinja"))

func viewCharHandler(w http.ResponseWriter, r *http.Request){
	limit := 100
	df, _ := models.GetAllCandle(config.Config.Symbol, limit)

	fmt.Printf("Dataframe is %v\n", df)

	err := templates.ExecuteTemplate(w, "google.jinja", df.Candles)
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

	df, _ := models.GetAllCandle(symbol, limit)

	js, err := json.Marshal(df)
	if err != nil{
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}


func StartWebServer() error {
	http.HandleFunc("/api/candle/", apiMakeHandler(apiCandleHandler))
	http.HandleFunc("/chart/", viewCharHandler)
	return http.ListenAndServe(fmt.Sprintf(":%d", config.Config.Port), nil)
}

