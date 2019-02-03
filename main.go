package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"sync"
	"time"

	"github.com/gobuffalo/packr"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/zserge/webview"
)

func main() {
	home, _ := homedir.Dir()
	datapath := path.Join(home, ".pomodoro")
	err := os.MkdirAll(datapath, os.FileMode(0722))
	if err != nil {
		log.Fatal(err)
	}
	box := packr.NewBox("./public")
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		defer ln.Close()
		http.Handle("/", http.FileServer(box))
		http.Handle("/events", eventHandler(datapath))
		log.Fatal(http.Serve(ln, nil))
	}()
	url := "http://" + ln.Addr().String()
	w := webview.New(webview.Settings{
		Width:  800,
		Height: 520,
		Title:  "Pomodoro",
		URL:    url,
	})
	log.Printf("DEBUG: starting at %s", url)
	defer w.Exit()
	w.Run()
}

type Event struct {
	Date time.Time `json:"ts"`
	Desc string    `json:"desc"`
}

type eventStorage struct {
	sync.Mutex
	file io.ReadWriteCloser
}

func NewEventStorage(rwc io.ReadWriteCloser) eventStorage {
	return eventStorage{file: rwc}
}

func NewEvent(desc string) Event {
	return Event{
		Date: time.Now(),
		Desc: desc,
	}
}

func (e *eventStorage) WriteEvent(event Event) error {
	e.Lock()
	defer e.Unlock()
	w := csv.NewWriter(e.file)
	w.Write([]string{event.Date.UTC().Format(time.RFC3339), event.Desc})
	if err := w.Error(); err != nil {
		return err
	}
	w.Flush()
	return nil
}

func (e *eventStorage) ReadEvents() ([]Event, error) {
	es := []Event{}
	e.Lock()
	defer e.Unlock()
	r := csv.NewReader(e.file)
	records, err := r.ReadAll()
	if err != nil {
		return es, err
	}
	for _, record := range records {
		t, err := time.ParseInLocation(time.RFC3339, record[0], time.UTC)
		if err != nil {
			return es, err
		}
		es = append(es, Event{Date: t, Desc: record[1]})
	}
	return es, nil
}

func eventHandler(directory string) http.Handler {
	today := time.Now().UTC().Format("2006-01-02")
	filename := fmt.Sprintf("events-%s.csv", today)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f, err := os.OpenFile(path.Join(directory, filename), os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}
		defer f.Close()
		es := NewEventStorage(f)
		switch r.Method {
		case http.MethodGet:
			events, err := es.ReadEvents()
			if err != nil {
				w.WriteHeader(500)
				w.Write([]byte(err.Error()))
				return
			}
			e, err := json.Marshal(events)
			if err != nil {
				w.WriteHeader(500)
				w.Write([]byte(err.Error()))
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write(e)
		case http.MethodPost:
			desc := r.URL.Query().Get("desc")
			err := es.WriteEvent(NewEvent(desc))
			if err != nil {
				w.WriteHeader(500)
				w.Write([]byte(err.Error()))
				return
			}
			w.Write([]byte("Thanks Poster: " + desc))
		default:
			http.Error(w, "Unexpected Method "+r.Method, http.StatusBadRequest)
		}
	})
}
