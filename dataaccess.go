package main

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/motaz/codeutils"

	_ "github.com/go-sql-driver/mysql"
)

func GetConfigurationParameter(param, defaultValue string) string {

	value := codeutils.GetConfigValue("powerstatus.ini", param)
	if value == "" {
		value = defaultValue
	}
	return value
}

var databaseServer, databaseUser, database, password string

func init() {

	databaseServer = GetConfigurationParameter("server", "localhost")
	databaseUser = GetConfigurationParameter("dbuser", "")

	database = GetConfigurationParameter("database", "Power")

	password = GetConfigurationParameter("dbpassword", "")

}

func SQLConnection() (*sql.DB, error) {

	connectionString := fmt.Sprintf("%v:%v@tcp(%s:3306)/%v?parseTime=true",
		databaseUser, password, databaseServer, database)

	var err error
	var db *sql.DB
	db, err = sql.Open("mysql", connectionString)
	if err != nil {

		println("Error in SQLConnection: " + err.Error())

	}

	return db, err
}

func UpdateLastTime(connection *sql.DB, key string) bool {

	sqlStatement := "update status set LastTime = now() where officekey = ?"

	_, err := connection.Exec(sqlStatement, key)
	if err == nil {

		return true
	} else {
		lastError := fmt.Sprintf("Error in UpdateLastTime: %v ", err.Error())
		println(lastError)

		return false
	}

}

func UpdatePowerOn(connection *sql.DB, key string, duration time.Duration) bool {

	sqlStatement := "update status set OnTime = now(), LastOffTime=? where officekey = ?"

	LastOffTime := fmt.Sprintf("%v", duration)
	if strings.Contains(LastOffTime, "m") {
		LastOffTime = LastOffTime[:strings.Index(LastOffTime, "m")+1]

	}
	_, err := connection.Exec(sqlStatement, LastOffTime, key)
	if err == nil {

		return true
	} else {
		lastError := fmt.Sprintf("Error in UpdatePowerOn: %v ", err.Error())
		println(lastError)

		return false
	}

}

type StatusType struct {
	ID          int
	Key         string
	OfficeName  string
	LastTime    time.Time
	OnTime      time.Time
	LastOffTime string
}

func readTime(aTime sql.NullString) (timeResult time.Time) {

	format := "2006-01-02T15:04:05Z"
	myloc := time.Local

	if aTime.Valid {
		timeResult, _ = time.ParseInLocation(format, aTime.String, myloc)

	} else {
		timeResult, _ = time.Parse("2006-01-02", "2000-01-01")
	}
	return
}

func GetStatuses(connection *sql.DB) (success bool, result []StatusType) {

	sqlStatement := `SELECT id, officekey, OfficeName, LastTime, OnTime, LastOffTime from status`
	rows, err := connection.Query(sqlStatement)

	if err != nil {

		println("Error in GetStatuses: " + err.Error())
		success = false

	} else {
		var aresult StatusType
		for rows.Next() {
			var lastTime sql.NullString
			var ontime sql.NullString
			err = rows.Scan(&aresult.ID, &aresult.Key, &aresult.OfficeName, &lastTime, &ontime, &aresult.LastOffTime)
			if err == nil {
				aresult.LastTime = readTime(lastTime)
				aresult.OnTime = readTime(ontime)

				result = append(result, aresult)
			} else {
				println(err.Error())
			}
		}

		success = true
	}
	return

}

func GetOneStatus(connection *sql.DB, key string) (success bool, result StatusType) {

	sqlStatement := `SELECT id, officekey, OfficeName, LastTime, OnTime from status ` +
		` where officekey = ?`
	row := connection.QueryRow(sqlStatement, key)

	var lastTime sql.NullString
	var ontime sql.NullString
	err := row.Scan(&result.ID, &result.Key, &result.OfficeName, &lastTime, &ontime)
	if err == nil {
		success = true
		result.LastTime = readTime(lastTime)
		result.OnTime = readTime(ontime)

	} else {
		success = false
		println(err.Error())
	}

	return

}
