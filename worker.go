package goclear

import "fmt"
import "os"
import "os/signal"
import "time"
import "sync"
import "syscall"
import "database/sql"
import _ "github.com/go-sql-driver/mysql"

// Initialize a global database object
var db *sql.DB
var SessionID int64
var records chan *VarDict
var wg sync.WaitGroup

func init() {
	InitializeConfig()

	var err error
	db, err = sql.Open("mysql", Config.DBPath)

	if err != nil {
		fmt.Println("Fail to initialize database, exit now...", err)
		os.Exit(1)
	}

	// If it is the first time, create the tables
	createSessionStmt := "create table if not exists Session (id INT AUTO_INCREMENT PRIMARY KEY , timestamp INTEGER, hostname TEXT, path TEXT)"
	createRecordsStmt := "create table if not exists Record (id INT AUTO_INCREMENT PRIMARY KEY, sessionID INTEGER REFERENCES Session (id), timestamp INTEGER, name TEXT, data TEXT)"
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
	// Currently only one worker, could support a configurable number of workers
	go Worker()
	// Finish() should be called when exiting abnormally
	go func(){
		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, syscall.SIGHUP, syscall.SIGINT,syscall.SIGTERM)
		<-signalChan
		Finish()
		fmt.Println("Clearing up before exit...")
		os.Exit(1)
	}()
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

		fmt.Println("In worker")
		varname := (*record)["name"]
		data := record.Dump()
		// save to database
		ExecuteSQLWithArguments(stmt, SessionID, timestamp, varname, data)
	}
}

func PostRecord(vardict *VarDict){
	// Just put it into work queue
	select {
	case records <- vardict:
		// Successfully sent to worker
		fmt.Println("Saving record to db")
	case <-time.After(time.Millisecond*100):
		fmt.Println("Fail to send variable value to worker")
	}
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
		fmt.Println("Fail to execute SQL:", sqlfmt, err)
	}
	return result
}
