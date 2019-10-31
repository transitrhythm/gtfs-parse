package main

import (
	"encoding/json"
	"fmt"
	"image/color"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg/draw"
)

type server struct {
	sync.RWMutex
	tripStart         time.Time
	tripID            string
	routeID           string
	vehicleID         string
	passengerCapacity int
	passengerLoad     int
	data              []time.Duration
	authorities       map[string]authority
	agencies          map[string]agency
	stops             map[string]stop
}

type authority struct {
	AuthorityID int
	Descriptor  string
}

type agency struct {
	AuthorityID int
	AgencyID    int
	Descriptor  string
}

type stop struct {
	StopID           string
	StopCode         string
	StopName         string
	StopDescription  string
	StopSequence     int
	IsTimingPoint    bool
	Lat              float32
	Lng              float32
	DistanceTraveled float32
}

func setupAuthorities() map[string]authority {
	authorities := map[string]authority{
		"bcTransit":          {1, "BC Transit"},
		"vancouverTranslink": {2, "Vancouver Translink"},
	}
	return authorities
}

func setupAgencies() (agencies map[string]agency) {
	agencies = map[string]agency{
		"comox":    {1, 12, "Comox Valley Transit System"},
		"kamloops": {1, 8, "Kamloops Transit System"},
		"kelowna":  {1, 7, "Kelowna Regional Transit System"},
		"nanaimo":  {1, 5, "RDN Transit System"},
		"squamish": {1, 4, "Squamish Transit System"},
		"victoria": {1, 1, "Victoria Regional Transit System"},
		"whistler": {1, 3, "Whistler Transit System"},
	}
	return agencies
}

func errorHandler(h func(http.ResponseWriter, *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := h(w, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (s *server) getAuthorities(w http.ResponseWriter, r *http.Request) error {
	s.RLock()
	defer s.RUnlock()
	data, err := json.Marshal(s.authorities)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	jsonData := string(data)
	fmt.Fprintf(w, "%s", jsonData)
	return nil
}

func (s *server) setAuthority(w http.ResponseWriter, r *http.Request) error {
	s.RLock()
	defer s.RUnlock()

	return nil
}

func (s *server) getAgencies(w http.ResponseWriter, r *http.Request) error {
	s.RLock()
	defer s.RUnlock()
	data, err := json.Marshal(s.agencies)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	jsonData := string(data)
	fmt.Fprintf(w, "%s", jsonData)
	return nil
}

func (s *server) setAgency(w http.ResponseWriter, r *http.Request) error {
	s.RLock()
	defer s.RUnlock()

	// setupFeed(agency)

	return nil
}

func httpServer() {
	var s server
	s.authorities = setupAuthorities()
	s.agencies = setupAgencies()

	http.HandleFunc("/", errorHandler(s.root))
	http.HandleFunc("/statz", errorHandler(s.statz))
	http.HandleFunc("/statz/trip.png", errorHandler(s.trip))
	http.HandleFunc("/statz/getAuthorities", errorHandler(s.getAuthorities))
	http.HandleFunc("/statz/getAgencies", errorHandler(s.getAgencies))
	http.HandleFunc("/statz/setAuthority", errorHandler(s.setAuthority))
	http.HandleFunc("/statz/setAgency", errorHandler(s.setAgency))

	log.Fatal(http.ListenAndServe("localhost:8081", nil))
}

func (s *server) statz(w http.ResponseWriter, r *http.Request) (err error) {
	var imageSplice []string
	htmlInterval := 1000
	htmlColumns := 3
	htmlZoom := 99 / htmlColumns
	htmlImage := fmt.Sprintf(`<img src="/statz/trip.png?rand=0" style="width:%d%%">`, htmlZoom)
	trips := 1
	for i := 0; i < trips; i++ {
		imageSplice = append(imageSplice, htmlImage)
	}
	htmlImages := strings.Join(imageSplice, " ")
	htmlText := fmt.Sprintf("%s", htmlImages)
	htmlPrelude := fmt.Sprintf(Prelude, htmlColumns, htmlColumns, htmlColumns)
	htmlEpilog := fmt.Sprintf(Epilog, htmlInterval)
	htmlText = htmlPrelude + HTMLDropdown + htmlImages + htmlEpilog
	fmt.Fprintf(w, "%s", htmlText)
	return err
}

func (s *server) trip(w http.ResponseWriter, r *http.Request) error {
	s.RLock()
	defer s.RUnlock()

	xys := make(plotter.XYs, len(s.data))
	for i, d := range s.data {
		xys[i].X = float64(i)
		xys[i].Y = float64(d) / float64(time.Millisecond)
	}
	sc, err := plotter.NewScatter(xys)
	if err != nil {
		return errors.Wrap(err, "could not create trip")
	}
	sc.GlyphStyle.Shape = draw.CrossGlyph{}

	avgs := make(plotter.XYs, len(s.data))
	sum := 0.0
	for i, d := range s.data {
		avgs[i].X = float64(i)
		sum += float64(d)
		avgs[i].Y = sum / (float64(i+1) * float64(time.Millisecond))
	}
	l, err := plotter.NewLine(avgs)
	if err != nil {
		return errors.Wrap(err, "could not create line")
	}
	l.Color = color.RGBA{G: 255, A: 255}

	g := plotter.NewGrid()
	g.Horizontal.Color = color.RGBA{R: 255, A: 255}
	g.Vertical.Width = 0

	p, err := plot.New()
	if err != nil {
		return errors.Wrap(err, "could not create plot")
	}
	p.Add(sc, l, g)
	heading := fmt.Sprintf("Trip %s profile", "123")
	p.Title.Text = heading
	p.Y.Label.Text = "mins"
	p.X.Label.Text = "Trip stage (mins)"
	p.X.Min = 0
	p.X.Max = 60
	p.Y.Min = 0
	p.Y.Max = 30

	wt, err := p.WriterTo(512, 100, "png")
	if err != nil {
		return errors.Wrap(err, "could not create writer to")
	}

	_, err = wt.WriteTo(w)
	return errors.Wrap(err, "could not write to output")
}

func (s *server) root(w http.ResponseWriter, r *http.Request) (err error) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return err
	}

	query := r.URL.RawQuery
	var data []byte
	// request := r.RequestURI
	switch words := strings.Split(query, "="); strings.ToLower(words[0]) {
	case "agency":
		data, err = SetAgency(s, words[1])
	case "authority":
		data, err = SetAuthority(s, words[1])
	case "trip":
		data, err = GetTripJSON(s, words[1])
	case "stop":
		data, err = GetStopJSON(s, words[1])
	}
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	jsonData := string(data)
	fmt.Fprintf(w, "%s", jsonData)

	// x := 500 + 200*rand.NormFloat64()
	// d := time.Duration(x) * time.Millisecond
	// // time.Sleep(d)
	// fmt.Fprintln(w, "slept for", d)

	// s.Lock()
	// s.data = append(s.data, d)
	// if len(s.data) > 1000 {
	// 	s.data = s.data[len(s.data)-1000:]
	// }
	// s.Unlock()
	return err
}


// SetAuthority -
func SetAuthority(s *server, authority string) (data []byte, err error) {
	if len(authority) == 0 {
		data, err = json.Marshal(s.authorities)
	} else {
		data, err = 
	}
	return data, err
}

// SetAgency -
func SetAgency(s *server, agency string) (data []byte, err error) {
	data, err = json.Marshal(s.agencies)
	return data, err
}

// GetTripJSON -
func GetTripJSON(s *server, tripID string) (data []byte, err error) {
	return data, err
}

// GetStopJSON -
func GetStopJSON(s *server, stopCode string) (data []byte, err error) {
	return data, err
}