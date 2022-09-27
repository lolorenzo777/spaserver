// Copyright 2022 by lolorenzo77. All rights reserved.
// Use of this source code is governed by MIT licence that can be found in the LICENSE file.

// this main package contains the web assembly source code.
// It's compiled into a '.wasm' file with "GOOS=js GOARCH=wasm go build -o ../webapp/main.wasm"
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/gowebapi/webapi"
	"github.com/gowebapi/webapi/dom"
	"github.com/gowebapi/webapi/html"
	"github.com/gowebapi/webapi/html/htmlevent"
)

const (
	html_smiley = `<i class="bi bi-emoji-smile"></i>`
)

// the main func is required by the wasm GO builder
// prints will appears in the console of the browser
func main() {
	c := make(chan struct{})
	fmt.Println("Go/WASM loaded")

	// here start the code to customize
	// Welcome code
	welcomeE := GetElementById("welcome")
	if welcomeE != nil {
		welcomeMsg := fmt.Sprintf("Welcome from Web Assembly code written in go %v", html_smiley)
		welcomeE.SetInnerHTML(welcomeMsg)
	}

	// handle button OnClick
	buttonCallApi := GetButtonById("btncallapi")
	buttonCallApi.SetOnClick(func(event *htmlevent.MouseEvent, currentTarget *html.HTMLElement) {
		log.Println("buttonCallApi clicked")
		ApiGetHealth()
	})

	// here end the code to customize
	fmt.Println("Go/WASM runing")
	<-c
}

// ApiGetHealth call the /api/health on the server, returning a texte with a server counter.
// This is necessarilly done async in a seperate go routine, see https://golang.org/pkg/syscall/js/#FuncOf
func ApiGetHealth() {

	go func() {
		msgE := GetElementById("servermessage")
		status := "dead"

		body, err := ApiGet("health", "msgcallapi")
		if err != nil {
			msgE.SetTextContent(&status)
			return
		}

		// parse received data
		type dataHealth struct {
			Health  string `json:'health'`
			Counter string `json:'counter'`
		}
		data := dataHealth{}
		jsonErr := json.Unmarshal(body, &data)
		if jsonErr != nil {
			status = jsonErr.Error()
			msgE.SetTextContent(&status)
			return
		}

		// display result
		status = fmt.Sprintf("%v, call counter: %s", data.Health, data.Counter)
		msgE.SetTextContent(&status)
	}()
}

/*
  helpers
*/

func ApiGet(apiname string, alertid string) ([]byte, error) {
	// call the server
	resp, errGet := http.Get("/api/" + apiname)
	if errGet != nil {
		log.Println(errGet)
		if alertid != "" {
			ShowAlert(alertid, errGet.Error(), "alert-danger")
		}
		return []byte{}, errGet
	}
	if resp.StatusCode != http.StatusOK {
		log.Println(errGet)
		err := fmt.Errorf("%v (%s)", resp.StatusCode, http.StatusText(resp.StatusCode))
		if alertid != "" {
			ShowAlert(alertid, err.Error(), "alert-warning")
		}
		return []byte{}, err
	}

	// extract the response
	body, errRead := io.ReadAll(resp.Body)
	if errRead != nil {
		log.Println(errRead)
		if alertid != "" {
			ShowAlert(alertid, errRead.Error(), "alert-danger")
		}
		return []byte{}, errRead
	}

	if alertid != "" {
		ShowAlert(alertid, "200 (OK)", "alert-success")
	}
	return body, nil
}

func GetElementById(elementId string) (htmlE *dom.Element) {
	doc := webapi.GetWindow().Document()
	if doc != nil {
		htmlE = doc.GetElementById(elementId)
	}
	if doc == nil || htmlE == nil {
		log.Printf("unable to find html element id=%q\n", elementId)
	}
	return htmlE
}

func GetButtonById(elementId string) (button *html.HTMLButtonElement) {
	htmlE := GetElementById(elementId)
	if htmlE != nil {
		button = html.HTMLButtonElementFromWrapper(htmlE)
		if button == nil {
			log.Printf("element id=%q is not a button\n", elementId)
		}
	}
	return button
}

var alertTypes = []string{"alert-primary", "alert-secondary", "alert-success", "alert-danger", "alert-warning", "alert-info", "alert-light", "alert-dark"}

func ShowAlert(elementid string, msg string, alerttype string) {
	msgE := GetElementById(elementid)
	msgE.SetTextContent(&msg)

	goodType := false
	classdata := append(alertTypes, "d-none", "d-block")
	class := msgE.ClassName()
	for _, data := range classdata {
		class = strings.Replace(class, data, "", -1)
		if data[:1] != "d" && data == alerttype {
			goodType = true
		}
	}
	if goodType {
		class += " d-block " + alerttype
		msgE.SetClassName(class)
	} else {
		log.Printf("unmanaged alert-type %q\n", alerttype)
	}
}
