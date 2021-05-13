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
	"database/sql"

	"github.com/joho/godotenv"
	"github.com/satori/go.uuid"
	_ "github.com/mattn/go-sqlite3"
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
var sqlcon *sql.DB
var insertsql *sql.Stmt

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

//login page
func signinHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	mobile:=r.FormValue("mobilenumber")
	session_token:=uuid.NewV4().String()
	
	//insert token into sql
	_,err:=insertsql.Exec(mobile,session_token)
	if err != nil { log.Println("Error executing Insert Stmt",err)	}
	
	//set cookie
	http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   session_token,
		Expires: time.Now().Add(24 * time.Hour),
	})
	http.ServeFile(w, r, "home.html")
}

//addbusiness
func addbusinessHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "addbusiness.html")
}

//post signin
func homeHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "home.html")
}

//open db connection and prepare sql
func opensqlDB() {
	sqlconn,err:=sql.Open("sqlite3","./auth")
	if err != nil { log.Print("error opening sql db: ", err)	}
	
	insertsql,err=sqlconn.Prepare("INSERT INTO session_auth (mobile, session_token, up_ts) VALUES(?,?,strftime('%s','now'))")
	if err != nil { log.Print("Prepare of INSERT failed: ", err)	}
	
	//defer insertsql.Close() // Prepared statements take up server resources and should be closed after use
 
}

func main() {	
	err := godotenv.Load()
	if err != nil { log.Println("Error loading .env file", err)	}
	
	port := os.Getenv("PORT")
	if port == "" {	port = "3000" }
	
	opensqlDB();
	
	fs := http.FileServer(http.Dir("assets"))
	
	mux := http.NewServeMux()
	mux.Handle("/assets/", http.StripPrefix("/assets/", fs))
	
	mux.HandleFunc("/index", indexHandler)
	mux.HandleFunc("/signin", signinHandler)
	mux.HandleFunc("/addbusiness", addbusinessHandler)
	mux.HandleFunc("/home", homeHandler)
	
	http.ListenAndServe(":"+port, mux)	
}