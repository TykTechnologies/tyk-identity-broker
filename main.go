package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/pat"
	"github.com/lonelycode/tyk-auth-proxy/providers"
	"github.com/lonelycode/tyk-auth-proxy/tap"
)

func init() {}

func main() {

	var config string = `
	{
		"UseProviders": [
			{
				"Name": "gplus",
				"Key": "504206531762-e3nk43d2svtut98odmknrclf6aa1hd4n.apps.googleusercontent.com",
				"Secret": "kRqL0F0ysPiM2sv-oyEwkw2F"
			}
		]
	}
	`

	thisProvider := providers.Social{}
	thisProvider.Init(tap.DummyIdentityHandler{}, []byte(config))

	p := pat.New()
	p.Get("/auth/{provider}/callback", thisProvider.HandleCallback)
	p.Get("/auth/{provider}", thisProvider.Handle)

	// p.Get("/", func(res http.ResponseWriter, req *http.Request) {
	// 	t, _ := template.New("foo").Parse(indexTemplate)
	// 	t.Execute(res, nil)
	// })
	fmt.Println("Listening...")
	http.ListenAndServe(":3000", p)

}
