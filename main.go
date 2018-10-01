package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/marni/goigc"
	"google.golang.org/appengine"
	"math"
	"net/http"
	"strconv"
	"time"
)

type IGCType struct {
	HDate       string 	`json:"H_date,omitempty"`
	Pilot       string 	`json:"pilot,omitempty"`
	Glider      string 	`json:"glider,omitempty"`
	GliderId    string 	`json:"glider_id,omitempty"`
	TrackLength float64	`json:"track_length,omitempty"`
}

type IGCArray struct {
	Id		int 		`json:"id,omitempty"`
	Url		string		`json:"url,omitempty"`
	Data	IGCType		`json:"data,omitempty"`
}

type IGCPostDisplay struct {
	Id		int			`json:"id,omitempty"`
}

type APIInfo struct {
	Uptime  string `json:"uptime,omitempty"`
	Info    string `json:"info,omitempty"`
	Version string `json:"version,ommitempty"`
}

var startTime = time.Now()
var igcData []IGCArray
var lastId int
var apiInfo APIInfo

/**
	Code found at: https://gist.github.com/cdipaolo/d3f8db3848278b49db68
 */
func hsin(theta float64) float64 {
	return math.Pow(math.Sin(theta/2), 2)
}

/**
	Code found at : https://gist.github.com/cdipaolo/d3f8db3848278b49db68
 */
func Distance(lat1, lon1, lat2, lon2 float64) float64 {
	// convert to radians
	// must cast radius as float to multiply later
	var la1, lo1, la2, lo2, r float64
	la1 = lat1 * math.Pi / 180
	lo1 = lon1 * math.Pi / 180
	la2 = lat2 * math.Pi / 180
	lo2 = lon2 * math.Pi / 180

	r = 6378100 // Earth radius in METERS

	// calculate
	h := hsin(la2-la1) + math.Cos(la1)*math.Cos(la2)*hsin(lo2-lo1)

	return 2 * r * math.Asin(math.Sqrt(h))
}

/*	GET /api
	Response type: 	application/json
	Response code: 	200
	Response: 		JSON object
 */
func GetAPI(w http.ResponseWriter, r *http.Request) {
	// Get uptime
	t, err := time.Parse(time.RFC3339, startTime.UTC().Format(time.RFC3339))

	if err == nil {
		// Set header content-type
		w.Header().Set("Content-Type", "application/json")

		// Set uptime
		apiInfo.Uptime = t.String()

		// Display JSON-data
		json.NewEncoder(w).Encode(apiInfo)
	} else {
		// Display error message
		http.Error(w, err.Error(), http.StatusNotFound)
	}
}

/*	GET /api/igc
	Response type:	application/json
	Response code:	200
	Response:		JSON array
 */
func GetIGC(w http.ResponseWriter, r *http.Request) {
	// Check if we have data
	if len(igcData) > 0 {
		// Set header content-type
		w.Header().Set("Content-Type", "application/json")

		// Declare int array
		var result []int

		// Loop through data and append it to array
		for _, item := range igcData {
			result = append(result, item.Id)
		}

		// Display JSON-data
		json.NewEncoder(w).Encode(result)
	} else {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}
}

/*	POST /api/igc
	Response type:	application/json
	Response code:	200 || 404
	Response:		JSON object
 */
func PostIGC(w http.ResponseWriter, r *http.Request) {
	url := r.FormValue("url")

	// Check if field is not emplty
	if url != "" {
		// Get track-data from IGC-library
		track, err := igc.ParseLocation(url)

		// There was an error on getting track-data
		if err != nil {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		} else {
			// Loop through current data
			for _, item := range igcData {
				// Check if data already has been saved
				if item.Data.HDate == track.Date.String() {
					w.Header().Set("Content-Type", "text/plain")

					fmt.Fprintln(w, "Error: Data already exists")
					http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
					return
				}
			}

			// Get all points from the track
			points := track.Points

			// Save first points
			lastLat := points[0].Lat
			lastLng := points[0].Lng

			var trackLength float64

			// Loop through all points
			for i := 1; i < len(points); i++ {
				// Add opp distance
				trackLength += Distance(float64(lastLat), float64(lastLng), float64(points[i].Lat), float64(points[i].Lng))
				// Save current point over previous points
				lastLat = points[i].Lat
				lastLng = points[i].Lng
			}

			// Set header content-type
			w.Header().Set("Content-Type", "application/json")

			// Increment lsat ID
			lastId++
			// Create data
			igcData = append(igcData,
				IGCArray{
					lastId,
					url,
					IGCType{
						track.Date.String(),
						track.Pilot,
						track.GliderType,
						track.GliderID,
						trackLength}})

			// Display last ID inserted
			json.NewEncoder(w).Encode(IGCPostDisplay{
				igcData[lastId - 1].Id})
		}
	} else {
		// Show 404 error
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}
}

/*	GET /api/igc/<id>
	Response type:	application/json
	Response code: 	200 || 404
	Response: 		JSON object
 */
func GetID(w http.ResponseWriter, r *http.Request) {
	// Get parameters from the URL
	params := mux.Vars(r)
	// Parsing int from string
	id, err := strconv.ParseInt(params["id"], 10, 64)

	// No error in parsing int
	if err == nil {
		var found = false
		var result IGCType

		// Loop through data
		for _, item := range igcData {
			// Check if id exists and set data if so
			if id == int64(item.Id) {
				found = true
				result = item.Data
			}
		}

		if found {
			// Set header content-type
			w.Header().Set("Content-Type", "application/json")

			// Display JSON-data
			json.NewEncoder(w).Encode(result)
		} else {
			// Item is not found
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		}
	} else {
		// Display error in int parsing
		http.Error(w, err.Error(), http.StatusNotFound)
	}
}

/* 	GET /api/igc/<id>/<field>
	Response type:	text/plain
	Response code:	200 || 404
	Response: 		IGCType field
 */
func GetIDField(w http.ResponseWriter, r *http.Request) {
	// Get parameters
	params := mux.Vars(r)
	var result string

	// Loop through data
	for _, item := range igcData {
		// Check if id exists
		if item.Data.GliderId == params["id"] {
			// Check if field exists
			switch params["field"] {
			case "h_date":
				result = item.Data.HDate
				break
			case "pilot":
				result = item.Data.Pilot
				break
			case "glider":
				result = item.Data.Glider
				break
			case "glider_id":
				result = item.Data.GliderId
				break
			case "track_length":
				result = fmt.Sprintf("%f", item.Data.TrackLength)
				break
			default:
				result = ""
				break
			}
		}
	}

	// Check results is not blank
	if result != "" {
		// Set header content-type
		w.Header().Set("Content-Type", "text/plain")

		// Print result
		fmt.Fprint(w, result)
	} else {
		// Display 404 not found
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}
}

func main() {
	version := "1.0"
	info := "Service for IGCType tracks"
	lastId = 0

	apiInfo = APIInfo{Uptime: "", Info: info, Version: version}

	router := mux.NewRouter()
	subrouter := router.PathPrefix("/igcinfo").Subrouter()
	subrouter.HandleFunc("/api", GetAPI).Methods("GET")
	subrouter.HandleFunc("/api/igc", GetIGC).Methods("GET")
	subrouter.HandleFunc("/api/igc", PostIGC).Methods("POST")
	subrouter.HandleFunc("/api/igc/{id}", GetID).Methods("GET")
	subrouter.HandleFunc("/api/igc/{id}/{field}", GetIDField).Methods("GET")

	http.Handle("/", subrouter)
	appengine.Main()
}