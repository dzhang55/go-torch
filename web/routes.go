package web

import (
	"encoding/gob"
	"encoding/json"
	"html/template"
	"io"
	"net/http"
	"math/rand"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/dzhang55/go-torch/config"
	"github.com/dzhang55/go-torch/tasks"
	"github.com/dzhang55/go-torch/transcription"
)

type route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type transcriptionJobData struct {
	AudioURL       string   `json:"audioURL"`
	EmailAddresses []string `json:"emailAddresses"`
	SearchWords    []string `json:"searchWords"`
}

type flash struct {
	Title string
	Body  string
}

var routes = []route{
	route{
		"add_job",
		"POST",
		"/add_job",
		initiateImageJobHandler,
	},
	route{
		"add_job_json",
		"POST",
		"/add_job_json",
		initiateImageJobHandlerJSON,
	},
	route{
		"health",
		"GET",
		"/health",
		healthHandler,
	},
	route{
		"job_status",
		"GET",
		"/job_status/{id}",
		jobStatusHandler,
	},
	route{
		"form",
		"GET",
		"/",
		formHandler,
	},
}

var (
	store        = sessions.NewCookieStore([]byte(config.Config.SecretKey))
	flashSession = "flash"
)

func init() {
	// register the flash struct with gob so that it can be stored in sessions
	gob.Register(&flash{})
}

// initiateImageJobHandlerJSON takes a POST request containing a json object,
// decodes it into a transcriptionJobData struct, and starts a transcription task.
func initiateImageJobHandlerJSON(w http.ResponseWriter, r *http.Request) {
	jsonData := new(transcriptionJobData)

	// unmarshal from the response body directly into our struct
	if err := json.NewDecoder(r.Body).Decode(jsonData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	executer := tasks.DefaultTaskExecuter
	executer.QueueTask(transcription.MakeIBMTaskFunction(jsonData.AudioURL, jsonData.EmailAddresses, jsonData.SearchWords))
}

// initiateImageJobHandler takes a POST request from a form,
// decodes it into a transcriptionJobData struct, and starts a transcription task.
func initiateImageJobHandler(w http.ResponseWriter, r *http.Request) {
	// executer := tasks.DefaultTaskExecuter
	// TODO: Use emails to email results
	// emails := strings.Split(r.FormValue("emails"), ",")

	// TODO: Classify the image as a torch or not.
	// id := executer.QueueTask(transcription.MakeIBMTaskFunction(r.FormValue("url"), emails, words))
	random := rand.Float32()
	output := ""
	if random < 0.33 {
		output = "This is definitely a very cool torch!"
	} else if random < 0.66 {
		output = "This torch makes me want to wet my bed!"
	} else {
		output = "This is not a torch."
	}

	session, err := store.Get(r, flashSession)
	if err != nil {
		log.Fatal(err)
	}

	session.AddFlash(flash{
		Title: "Profiling Successful!",
		Body:  output,
	})
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusFound)
}

// healthHandler returns a 200 response to the client if the server is healthy.
func healthHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "OK :)")
}

// jobStatusHandler returns the status of a task with given id.
func jobStatusHandler(w http.ResponseWriter, r *http.Request) {
	args := mux.Vars(r)
	id := args["id"]

	executer := tasks.DefaultTaskExecuter
	status := executer.GetTaskStatus(id)
	io.WriteString(w, status.String())
}

func formHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/form.html")
	if err != nil {
		log.Fatal(err)
	}

	session, err := store.Get(r, flashSession)
	if err != nil {
		log.Fatal(err)
	}

	flashes := session.Flashes()
	session.Save(r, w)

	err = t.Execute(w, flashes)
	if err != nil {
		log.Fatal(err)
	}
}
