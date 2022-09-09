package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"
)

const OFFTIME = 15
const DELAYTIME = 8

func setStatus(w http.ResponseWriter, r *http.Request) {

	key := r.FormValue("key")
	if key != "" {

		db, err := SQLConnection()
		if err == nil {
			defer db.Close()
			success, result := GetOneStatus(db, key)

			if success {

				longtime, dur := hasLongTime(result.LastTime, OFFTIME)
				if longtime {
					UpdatePowerOn(db, key, dur)

				}
				atime := r.FormValue("atime")
				boottime := time.Now()
				if atime != "" {
					atime = strings.Replace(atime, "_", " ", 1)
					boottime, _ = time.Parse("2006-01-02 15:04:05", atime)

				}
				updated := UpdateLastTime(db, key, boottime)
				if updated {

					fmt.Fprintln(w, fmt.Sprintf("%s, has updated: %v ", key, dur))

				}

			}
		}
	}

}

type officestatus struct {
	ID          int
	OfficeName  string
	LastTime    string
	StatusColor string
	StatusName  string
	Since       string
	OnTime      string
	LastOffTime string
	BootTime    string
	BootSince   string
}

type displayStatusStruct struct {
	LastStatus   string
	OfficeStatus []officestatus
}

func displayStatus(w http.ResponseWriter, r *http.Request) {

	var status displayStatusStruct
	r.Header.Set("content-type", "text/html")
	db, err := SQLConnection()
	if err == nil {
		defer db.Close()
		success, list := GetStatuses(db)
		format := "2006-01-02 15:04:05"
		status.LastStatus = time.Now().Format(format)

		if success {
			fmt.Fprintf(w, "<html>")
			for _, item := range list {
				var record officestatus
				record.ID = item.ID
				record.OfficeName = item.OfficeName
				record.LastTime = item.LastTime.Format(format)
				record.LastOffTime = item.LastOffTime
				if strings.Contains(record.LastOffTime, "h") {
					record.LastOffTime = strings.ReplaceAll(record.LastOffTime, "h", "h ")

				}

				var statusColor, statusName string
				var calcTime time.Time

				longtime, _ := hasLongTime(item.LastTime, OFFTIME)
				delayLongtime, _ := hasLongTime(item.LastTime, DELAYTIME)
				if longtime {
					statusColor = "red"
					statusName = "Off"
					calcTime = item.LastTime
				} else if delayLongtime {
					statusColor = "yellow"
					statusName = "Delayed"
					calcTime = item.OnTime
				} else {
					statusColor = "lime"
					statusName = "On"
					calcTime = item.OnTime
				}
				record.StatusColor = statusColor
				record.StatusName = statusName

				record.Since = getTimeDiff(calcTime)
				record.BootSince = getTimeDiff(item.BootTime)

				record.OnTime = item.OnTime.Format(format)
				record.BootTime = item.BootTime.Format(format)
				status.OfficeStatus = append(status.OfficeStatus, record)

			}
		}
	}
	Template, err := template.ParseFiles("templates/index.html")
	if err != nil {
		fmt.Println("Error loading template : ", err)
	}
	Template.Execute(w, status)
}

func getTimeDiff(atime time.Time) (since string) {

	since = fmt.Sprintf("%s", time.Now().Sub(atime))
	if strings.Contains(since, "m") {
		since = since[:strings.Index(since, "m")+1]
	}
	since = strings.ReplaceAll(since, "h", "h ")
	since = strings.ReplaceAll(since, "d", "d ")
	return
}

func hasLongTime(atime time.Time, minutes int) (longtime bool, diff time.Duration) {

	dur := time.Duration(minutes)
	diff = time.Now().Sub(atime)

	advTime := atime.Add(time.Minute * dur)

	longtime = time.Now().After(advTime)

	return
}
