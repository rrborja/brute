package brute

import (
	"net/http"
	"html/template"
	"github.com/rrborja/brute/assets"
)

var template401Page *template.Template
var template404Page *template.Template
var template700Page *template.Template

func init() {
	var data []byte
	var err error

	/* Parse 404 page template */
	data, err = assets.Asset("static/default-pages/401/401.html"); check(err)
	template401Page, err = template.New("401 Page Template").Parse(string(data)); check(err)

	/* Parse 404 page template */
	data, err = assets.Asset("static/default-pages/404/404.html"); check(err)
	template404Page, err = template.New("404 Page Template").Parse(string(data)); check(err)

	/* Parse 700 page template */
	data, err = assets.Asset("static/default-pages/700/700.html"); check(err)
	template700Page, err = template.New("700 Page Template").Parse(string(data)); check(err)


}

func defaultNotFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(404)
	template404Page.Execute(w, &struct{
		ProjectName string
		Path string
		RandomNoun string
	}{projectName, r.URL.Path, "MyEndpoint"})
}

func defaultUnauthorizedHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(401)
	template401Page.Execute(w, &struct{
		Path string
	}{Path: r.URL.Path})
}