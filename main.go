package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/marni/goigc"
	"google.golang.org/appengine"
	"log"
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
var igcArray []IGCArray
var igcs []IGCType
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
	w.Header().Set("Content-Type", "application/json")

	t, _ := time.Parse(time.RFC3339, startTime.UTC().Format(time.RFC3339))
	apiInfo.Uptime = t.String()

	json.NewEncoder(w).Encode(apiInfo)
}

/*	GET /api/igc
	Response type:	application/json
	Response code:	200
	Response:		JSON array
 */
func GetIGC(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var arr []int
	
	for _, item := range igcArray {
		arr = append(arr, item.Id)
	}
	
	json.NewEncoder(w).Encode(arr)
}

/*	POST /api/igc
	Response type:	application/json
	Response code:	200 || 404
	Response:		JSON object
 */
func PostIGC(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

	url := r.FormValue("url")

	if url != "" {
		track, err := igc.ParseLocation(url)

		if err != nil {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		} else {
			for _, item := range igcArray {
				if item.Data.HDate == track.Date.String() {
					fmt.Fprintln(w, "Error: Data already exists")
					http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
					return
				}
			}

			points := track.Points

			lastLat := points[0].Lat
			lastLng := points[0].Lng

			var trackLength float64

			for i := 1; i < len(points); i++ {
				trackLength += Distance(float64(lastLat), float64(lastLng), float64(points[i].Lat), float64(points[i].Lng))
				lastLat = points[i].Lat
				lastLng = points[i].Lng
			}

			lastId++
			igcArray = append(igcArray, IGCArray{lastId, url, IGCType{track.Date.String(), track.Pilot, track.GliderType, track.GliderID, trackLength}})
			json.NewEncoder(w).Encode(IGCPostDisplay{igcArray[lastId - 1].Id})
		}
	} else {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}
}

/*	GET /api/igc/<id>
	Response type:	application/json
	Response code: 	200 || 404
	Response: 		JSON object
 */
func GetID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	params := mux.Vars(r)
	id, err := strconv.ParseInt(params["id"], 10, 64)

	if err != nil {
		log.Fatal(err.Error())
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	var found = false
	var res IGCType

	for _, item := range igcArray {
		if id == int64(item.Id) {
			found = true
			res = item.Data
		}
	}

	if found {
		json.NewEncoder(w).Encode(res)
	} else {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}
}

/* 	GET /api/igc/<id>/<field>
	Response type:	text/plain
	Response code:	200 || 404
	Response: 		IGCType field
 */
func GetIDField(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	params := mux.Vars(r)
	var res string

	for _, item := range igcs {
		if item.GliderId == params["id"] {

			switch params["field"] {
			case "h_date":
				res = item.HDate
				break
			case "pilot":
				res = item.Pilot
				break
			case "glider":
				res = item.Glider
				break
			case "glider_id":
				res = item.GliderId
				break
			case "track_length":
				res = fmt.Sprintf("%f", item.TrackLength)
				break
			default:
				res = ""
				break
			}
		}
	}

	if res != "" {
		fmt.Fprint(w, res)
	} else {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}
}

func main() {
	lastId = 0

	apiInfo = APIInfo{Uptime: "", Info:"Service for IGCType tracks.", Version:"v1"}

	router := mux.NewRouter()
	router.HandleFunc("/igcinfo/api", GetAPI).Methods("GET")
	router.HandleFunc("/igcinfo/api/igc", GetIGC).Methods("GET")
	router.HandleFunc("/igcinfo/api/igc", PostIGC).Methods("POST")
	router.HandleFunc("/igcinfo/api/igc/{id}", GetID).Methods("GET")
	router.HandleFunc("/igcinfo/api/igc/{id}/{field}", GetIDField).Methods("GET")

	http.Handle("/", router)
	appengine.Main()
}