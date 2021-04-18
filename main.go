package main

import (
	htemplate "html/template"
//	ttemplate "text/template"
	"log"
	"net/http"
	"os"
	"io/ioutil"
	"time"
	"encoding/json"

	"github.com/joho/godotenv"
//	"github.com/PuerkitoBio/goquery"
//	"github.com/tidwall/gjson"
)

type Businesstable struct {
	TotalRows int `json:"total_rows"`
	Offset    int `json:"offset"`
	Rows      []struct {
		ID    string `json:"id"`
		Key   string `json:"key"`
		Value struct {
			Rev string `json:"rev"`
		} `json:"value"`
		Doc struct {
			ID       string `json:"_id"`
			Rev      string `json:"_rev"`
			Login    string `json:"login"`
			Lat      string `json:"lat"`
			Long     string `json:"long"`
			Business struct {
				Name     string   `json:"name"`
				Address  string   `json:"address"`
				Locality string   `json:"locality"`
				City     string   `json:"city"`
				Pincode  string   `json:"pincode"`
				Category []string `json:"category"`
				Phone    string   `json:"phone"`
				Likes    string   `json:"likes"`
				Dislikes string   `json:"dislikes"`
				Image    string   `json:"image"`
				Product  []struct {
					ID      string `json:"id"`
					Image   string `json:"Image"`
					Name    string `json:"name"`
					Desc    string `json:"desc"`
					Pricers string `json:"priceRs"`
				} `json:"product"`
			} `json:"business"`
		} `json:"doc"`
	} `json:"rows"`
}

var businessobject Businesstable

//Index page
func indexHandler(w http.ResponseWriter, r *http.Request) {
	client := &http.Client{Timeout: time.Second * 2,} // Timeout after 2 seconds

	req, err := http.NewRequest("GET", "http://admin:admin@127.0.0.1:5984/business/_all_docs?startkey=%22%22&limit=10&include_docs=true", nil)
	if err != nil {	log.Print("error in calling all_docs - 1:",err)	}

//  call db
	resp, err := client.Do(req)
	if err != nil {	log.Print("error in calling all_docs - 2:",err)	}

	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil { log.Print("error reading response all_docs: ",err)}
	
//move db data to struct    
	err = json.Unmarshal(bodyBytes, &businessobject)      
    if err != nil { log.Print("error unmarshalling string to json: ",err)}

	t, err := htemplate.ParseFiles("index.html")
    if err != nil { log.Print("template parsing error: ", err)	}
	
    err = t.Execute(w, businessobject.Rows) 
    if err != nil { log.Print("template executing error: ", err) }	
}

//sendOTP page
func sendOTPHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "sendOTP.html")
}

//verify page
func verifyOTPHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	log.Printf("%s",r.FormValue("mobilenumber"))
	http.ServeFile(w, r, "verifyOTP.html")
}

func main() {	
	err := godotenv.Load()
	if err != nil { log.Println("Error loading .env file")	}
	
	port := os.Getenv("PORT")
	if port == "" {	port = "3000" }

	fs := http.FileServer(http.Dir("assets"))
	
	mux := http.NewServeMux()
	mux.Handle("/assets/", http.StripPrefix("/assets/", fs))
	
	mux.HandleFunc("/index", indexHandler)
	mux.HandleFunc("/sendotp", sendOTPHandler)
	mux.HandleFunc("/verifyotp", verifyOTPHandler)
	
	http.ListenAndServe(":"+port, mux)	
}
