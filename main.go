package main

import (
	htemplate "html/template"
	"fmt"
	"io"
//	ttemplate "text/template"
//	"reflect"
	"bytes"
	"strings"
	"strconv"
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

type Singledoc struct {
	ID       string `json:"_id"`
	Login    string `json:"login"`
	Lat      string `json:"lat"`
	Long     string `json:"long"`
	Business struct {
		Category string   `json:"category"`
		Name     string   `json:"name"`
		Address  string   `json:"address"`
		Locality string   `json:"locality"`
		City     string   `json:"city"`
		Pincode  string   `json:"pincode"`
		State    string   `json:"state"`
		Tag    []string   `json:"tag"`
		Phone    string   `json:"phone"`
		AlternatePhone    string   `json:"alternatephone"`
		Likes    string   `json:"likes"`
		Dislikes string   `json:"dislikes"`
		Image    string   `json:"image"`
		Product  *[]struct { //defined as pointer to omit null values
			ID      string `json:"id"`
			Image   string `json:"Image"`
			Name    string `json:"name"`
			Desc    string `json:"desc"`
			Pricers string `json:"priceRs"`
		} `json:",omitempty"` 
	} `json:"business"`
}
var singledoc Singledoc

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
	Mobile:=r.FormValue("mobilenumber")
	Session_token:=uuid.NewV4().String()
	
	//insert token into sqlite
	sqlconn,err:=sql.Open("sqlite3","./auth")
	if err != nil { log.Print("error opening sql db-INSERT: ", err)	}
	defer sqlconn.Close() 
	
	insertsql,err=sqlconn.Prepare("INSERT INTO session_auth (mobile, session_token, up_ts) VALUES(?,?,strftime('%s','now'))")
	if err != nil { log.Print("Prepare of INSERT failed: ", err)	}
	
	_,err=insertsql.Exec(Mobile,Session_token)
	if err != nil { log.Println("Error executing Insert Stmt",err)	}
	
	//set cookie
	http.SetCookie(w, &http.Cookie{
		Name:    "session",
		Value:   Session_token,
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
	
	cookie1, err := r.Cookie("session")
	if err != nil { log.Print("retrieve Cookie failed: ", err)	} //retrieve cookie
	
	log.Print("cookie: ", cookie1.Value)
	
	sqlconn,err:=sql.Open("sqlite3","./auth")
	if err != nil { log.Print("error opening sql db-SELECT: ", err)	}
	defer sqlconn.Close() 
	
	selectsql:=sqlconn.QueryRow("SELECT mobile,session_token, up_ts FROM session_auth where session_token=?",cookie1.Value)
	
	var db_mobile, db_session_token string
	var db_up_ts int
	
	err=selectsql.Scan(&db_mobile, &db_session_token, &db_up_ts)
	if err != nil { log.Println("Error Fetching Select Stmt",err)	}
			
	r.ParseMultipartForm(32 << 20)

	singledoc.Business.Name = r.FormValue("Business Name")
	singledoc.Business.Address = r.FormValue("Address")
	singledoc.Business.Locality = r.FormValue("Locality")
	singledoc.Business.City = r.FormValue("city")
	singledoc.Business.Pincode = r.FormValue("Pincode")
	singledoc.Business.State = r.FormValue("State")
	singledoc.Business.Phone = r.FormValue("Phone")
	singledoc.Business.Category = r.FormValue("Business Category")
	singledoc.Business.Likes = "0"
	singledoc.Business.Dislikes = "0"
	if (r.FormValue("Alternative Phone") !=""){singledoc.Business.AlternatePhone = r.FormValue("Alternative Phone")}

	singledoc.Business.Tag = r.Form["tag"] // slice	
		if s, err := strconv.ParseFloat(r.FormValue("latitude"), 64); err == nil {
		singledoc.Lat = fmt.Sprintf("%.7f", s)	//round off to 7 digits
	} else {log.Print("string to float failed for Lat: ", err)} //convert string to float
	
	if s, err := strconv.ParseFloat(r.FormValue("longitude"), 64); err == nil {
		singledoc.Long = fmt.Sprintf("%.7f", s)	//round off to 7 digits
	} else {log.Print("string to float failed for Long: ", err)} //convert string to float
	
	singledoc.Login=db_mobile
	singledoc.ID=singledoc.Lat[0:strings.Index(singledoc.Lat, ".") +3 ] + ":" + singledoc.Long + ":" + singledoc.Login

	filepath1:="assets"+"/b"+ singledoc.Lat[0:strings.Index(singledoc.Lat, ".") ] + "_" +
		singledoc.Lat[(strings.Index(singledoc.Lat,".")+1) : (strings.Index(singledoc.Lat, ".")+3)] 
		
	businessfilepath:="/b" + singledoc.Login + "_business.jpg"
	
	filepath:= filepath1 + businessfilepath
	
	in, _, err := r.FormFile("businesspicture") // image
	if err == nil { 
		log.Print("Uploading Business picture file")
		defer in.Close()
		if _, err := os.Stat(filepath1); os.IsNotExist(err) {
			log.Print("inside 1")
			if err=os.MkdirAll(filepath1, 0770);err!=nil {log.Print("mkdirall failed: ", err)	} 			 			
		} 
		out, err := os.Create(filepath) 
		if err != nil { log.Print("create file failed: ", err)	
		} else {
			defer out.Close()		
			io.Copy(out, in)
		}	
	} else { log.Print("Uploading Business picture file failed: ", err)	}
	
	
//post in db	
	client := &http.Client{Timeout: time.Second * 2,} // Timeout after 2 seconds
	
	buf := new(bytes.Buffer)
	json.NewEncoder(buf).Encode(singledoc)
	log.Println(buf)
	req, err := http.NewRequest("POST", "http://admin:admin@127.0.0.1:5984/business", buf)
	req.Header.Add("Content-Type", "application/json")
	if err != nil {	log.Print("error in calling all_docs - 1:",err)	}

//  execute req
	resp, err := client.Do(req)
	if err != nil {	log.Print("error in calling all_docs - 2:",err)	}
	defer resp.Body.Close()
	
	log.Println(resp.Status)
	
	b, _ := json.Marshal(singledoc)
	log.Println(string(b))
//	log.Print(reflect.TypeOf(r.Form["tag"]))
		
	http.ServeFile(w, r, "home.html")
}
func main() {	

	err := godotenv.Load()
	if err != nil { log.Println("Error loading .env file", err)	}
	
	port := os.Getenv("PORT")
	if port == "" {	port = "3000" }
	
	fs := http.FileServer(http.Dir("assets"))
	
	mux := http.NewServeMux()
	mux.Handle("/assets/", http.StripPrefix("/assets/", fs))
	
	mux.HandleFunc("/index", indexHandler)
	mux.HandleFunc("/signin", signinHandler)
	mux.HandleFunc("/addbusiness", addbusinessHandler)
	mux.HandleFunc("/home", homeHandler)
	
	http.ListenAndServe(":"+port, mux)	
}