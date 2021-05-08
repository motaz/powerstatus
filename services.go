package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

const OFFTIME = 25
const DELAYTIME = 10

func setStatus(w http.ResponseWriter, r *http.Request) {

	key := r.FormValue("key")
	if key != "" {

		db, err := SQLConnection()
		if err == nil {
			defer db.Close()
			success, result := GetOneStatus(db, key)

			if success {
				println(result.LastTime.String())
				if hasLongTime(result.LastTime, OFFTIME) {
					UpdatePowerOn(db, key)
				}
				updated := UpdateLastTime(db, key)
				if updated {
					fmt.Fprintln(w, key+", has updated")
				}

			}
		}
	}

}

func displayStatus(w http.ResponseWriter, r *http.Request) {

	r.Header.Set("content-type", "text/html")
	fmt.Fprintln(w, "<html><head><title>Power Status</title>")
	fmt.Fprintln(w, `<link href="/powerstatus/static/style.css" rel="stylesheet" type="text/css">`)
	fmt.Fprintln(w, "<meta http-equiv='refresh' content='30'>")
	fmt.Fprintln(w, "</head><body>")
	fmt.Fprintln(w, "<h2>Power Status</h2>")
	db, err := SQLConnection()
	if err == nil {
		defer db.Close()
		success, list := GetStatuses(db)
		format := "2006-01-02 15:04:05 MST"
		fmt.Fprintln(w, "<b>Status at: </b>"+time.Now().Format(format))

		if success {
			fmt.Fprintln(w, "<table class=dtable><tr><th>ID</th><th>Office Name</th>")
			fmt.Fprintln(w, "<th>Last Status</th><th>Status</th><th>Since</th>")
			fmt.Fprintln(w, "<th>Last On Time</th></tr>")
			for _, item := range list {
				fmt.Fprintln(w, "<tr>")
				fmt.Fprintf(w, "<td>%d</td>", item.ID)
				fmt.Fprintf(w, "<td>%s</td>", item.OfficeName)
				fmt.Fprintf(w, "<td>%s</td>", item.LastTime.Format(format))

				var statusColor, statusName string
				var calcTime time.Time

				if hasLongTime(item.LastTime, OFFTIME) {
					statusColor = "red"
					statusName = "Off"
					calcTime = item.LastTime
				} else if hasLongTime(item.LastTime, DELAYTIME) {
					statusColor = "yellow"
					statusName = "Delayed"
					calcTime = item.OnTime
				} else {
					statusColor = "lime"
					statusName = "On"
					calcTime = item.OnTime
				}
				fmt.Fprintf(w, "<td bgcolor='%s'>%s</td>", statusColor, statusName)

				since := fmt.Sprintf("%s", time.Now().Sub(calcTime))
				if strings.Contains(since, "m") {
					since = since[:strings.Index(since, "m")+1]
				}
				since = strings.ReplaceAll(since, "h", "h ")
				since = strings.ReplaceAll(since, "d", "d ")
				fmt.Fprintf(w, "<td>%s</td>", since)

				fmt.Fprintf(w, "<td>%s</td>", item.OnTime.Format(format))

				fmt.Fprintln(w, "</tr>")
			}
			fmt.Fprintln(w, "</table>")
		}
	}
	fmt.Fprintln(w, "</body></html>")

}

func hasLongTime(atime time.Time, minutes int) (longtime bool) {

	dur := time.Duration(minutes)

	advTime := atime.Add(time.Minute * dur)

	longtime = time.Now().After(advTime)

	return
}
