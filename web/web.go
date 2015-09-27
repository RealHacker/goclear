package main

import "net/http"
import "fmt"
import "os"
import "encoding/json"
import "github.com/gorilla/mux"
import "database/sql"
import "html/template"
import "strconv"
import _ "github.com/go-sql-driver/mysql"

var db *sql.DB

func init() {
	// TODO: This will come from common config file
	DBPath := "root@/goclear"

	var err error
	db, err = sql.Open("mysql", DBPath)

	if err != nil {
		fmt.Println("Fail to initialize database, exit now...", err)
		os.Exit(1)
	}
}

func main(){
	router := mux.NewRouter()
	router.HandleFunc("/", listHandler)
	router.HandleFunc("/{sessionid:[0-9]+}/", recordsHandler)
	router.HandleFunc("/{sessionid:[0-9]+}/{recordid}/", recordsHandler)
	http.ListenAndServe(":8000", router)
}

func listHandler(w http.ResponseWriter, r *http.Request){
	sql := "select id, timestamp, hostname, path from Session order by id desc"
	rows, err := db.Query(sql)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	type Session struct {
		Id int64 `json:"id"`
		Timestamp int64 `json:"timestamp"`
		Hostname string `json:"hostname"`	
		Path string `json:"path"`
	}
	sessions := make([]Session, 0)
	for rows.Next() {
		var s Session
		err = rows.Scan(&s.Id, &s.Timestamp, &s.Hostname, &s.Path)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		sessions = append(sessions, s)
	}
	tpl, err := template.ParseFiles("index.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = tpl.Execute(w, sessions)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func recordsHandler(w http.ResponseWriter, r *http.Request){
	vars := mux.Vars(r)
	sessionid, err := strconv.Atoi(vars["sessionid"])
	if err!= nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	var recordStart int
	recordidStr, ok := vars["recordid"]
	if !ok {
		recordStart = 0
	} else {
		recordStart, _ = strconv.Atoi(recordidStr)
	}

	q := "select id, timestamp, name, data from Record where sessionID=? and id>? order by id limit 10"
	rows, err := db.Query(q, sessionid, recordStart)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	type Record struct {
		Id int64 `json:"id"`
		Timestamp int64 `json:"timestamp"`
		Name string `json:"name"`	
		Data string `json:"data"`
	}
	records := make([]Record, 0)
	for rows.Next() {
		var r Record
		err = rows.Scan(&r.Id, &r.Timestamp, &r.Name, &r.Data)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		records = append(records, r)
	}
	// marshal to json
	b, err := json.Marshal(records)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}