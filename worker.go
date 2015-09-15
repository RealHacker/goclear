package goclear

import "fmt"
import "os"
import "time"
import "sync"
import "database/sql"
import _ "github.com/mattn/go-sqlite3"

// Initialize a global database object
var db *sql.DB
var SessionID int64
var records chan *VarDict
var wg sync.WaitGroup

func init() {
	InitializeConfig()
	db, err := sql.Open("sqlite", Config.DBPath)
	if err != nil {
		fmt.Println("Fail to initialize database, exit now...")
		return 
	}

	// If it is the first time, create the tables
	createSessionStmt := "create table if not exists Session (id INTEGER PRIMARY KEY, timestamp INTEGER, hostname TEXT, path TEXT)"
	createRecordsStmt := "create table if not exists Record (id INTEGER PRIMARY KEY, sessionID INTEGER REFERENCES Session (id), timestamp INTEGER, name TEXT, data TEXT"
	ExecuteSQL(createSessionStmt)
	ExecuteSQL(createRecordsStmt)

	// Get the execution environment, including hostname, path and timestamp
	hostname, err1 := os.Hostname()
	cwd, err2 := os.Getwd()
	if err1 != nil || err2 != nil {
		fmt.Println("Fail to get environment information:", err1, err2)
	}
	timestamp := time.Now().Unix()
	newSessionStmt := "insert into Session (timestamp, hostname, path) values (?, ?, ?)"
	result := ExecuteSQLWithArguments(newSessionStmt, timestamp, hostname, cwd)
	
	// keep record of the session id
	SessionID, err = result.LastInsertId()
	if err != nil {
		fmt.Println("Fail to get session ID, exit")
		return
	}

	// Initialize the channel for VarDicts to be saved 
	records = make(chan *VarDict, 100)
	// Start the worker to listen on the channel
	wg.Add(1)
	go Worker()

}

// This function should be called before exiting the application, both normal exit and killing
func Finish() {
	// close the channel
	close(records)
	// Wait for the goroutine to finish
	wg.Wait()
	// Close the database
	db.Close()
}

func Worker() {
	if db == nil || SessionID == 0 {
		return
	}
	defer wg.Done()
	// The task of this worker is to take VarDicts from the queue and put them into database
	stmt := "insert into Record (sessionID, timestamp, name, data) values (?, ?, ?, ?)"
	for record := range records {
		timestamp := time.Now().Unix()
		varname := *record["name"]
		data := record.Dump()
		// save to database
		ExecuteSQLWithArguments(stmt, SessionID, timestamp, varname, data)
	}
}

func PostRecord(vardict *VarDict){
	// Just put it into work queue
	records <- vardict
}

func ExecuteSQL(sql string) sql.Result{
	if db == nil {
		return nil
	}
	result, err := db.Exec(sql)
	if err != nil {
		fmt.Println("Fail to execute SQL:", sql, err)
	}
	return result
}

func ExecuteSQLWithArguments(sqlfmt string, args ...interface{}) sql.Result{
	if db == nil {
		return nil
	}
	result, err := db.Exec(sqlfmt, args...)
	if err != nil {
		fmt.Println("Fail to execute SQL:", sqlfmt)
	}
	return result
}