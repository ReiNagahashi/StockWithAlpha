package controllers

import (
	"fmt"
	"html/template"
	"net/http"
	"stock-with-alpha/config"
)

var templates = template.Must(template.ParseFiles("app/views/google.html"))

func viewCharHandler(w http.ResponseWriter, r *http.Request){
	err := templates.ExecuteTemplate(w, "google.html", nil)
	if err != nil{
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func StartWebServer() error {
	http.HandleFunc("/chart/", viewCharHandler)
	return http.ListenAndServe(fmt.Sprintf(":%d", config.Config.Port), nil)
}