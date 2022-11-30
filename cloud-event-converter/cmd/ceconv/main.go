package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/vorteil/direktiv-listeners/cloud-event-converter/pkg/ceconv"
)

func main() {

	r := mux.NewRouter()
	r.HandleFunc("/{path:.*}", eventHandler)

	srv := &http.Server{
		Addr:    cfg.Server.Bind,
		Handler: r,
	}

	if cfg.Server.TLS.Enabled {
		log.Printf("(https) listen and serve @ '%s'\n", cfg.Server.Bind)
		err := srv.ListenAndServeTLS(cfg.Server.TLS.Cert, cfg.Server.TLS.Key)
		if err != nil {
			log.Fatal(err.Error())
		}
	} else {
		log.Printf("(http) listen and serve @ '%s'\n", cfg.Server.Bind)
		err := srv.ListenAndServe()
		if err != nil {
			log.Fatal(err.Error())
		}
	}

}

func eventHandler(w http.ResponseWriter, r *http.Request) {

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		respondErr(w, err)
		return
	}

	if len(b) == 0 {
		badRequest(w, err)
		return
	}

	m, err := ceconv.MapFromByteSlice(b)
	if err != nil {
		respondErr(w, err)
		return
	}

	req, err := transformPayload(r, m)
	if err != nil {
		respondErr(w, err)
		return
	}
	defer req.Body.Close()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		respondErr(w, err)
		return
	}
	defer resp.Body.Close()

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func respondErr(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	io.Copy(w, strings.NewReader(err.Error()))
}

func badRequest(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	io.Copy(w, strings.NewReader(err.Error()))
}

func respond(w http.ResponseWriter, b []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(b)))

	_, err := io.Copy(w, bytes.NewReader(b))
	if err != nil {
		log.Printf("respond() failed: %s", err.Error())
	}
}

func transformPayload(r *http.Request, m map[string]interface{}) (*http.Request, error) {

	for _, rule := range cfg.Rules {

		c, err := ceconv.LoadCondition(rule.Condition)
		if err != nil {
			return nil, err
		}

		ok, err := c.Evaluate(m)
		if err != nil {
			return nil, err
		}

		if !ok {
			continue
		}

		mods := strings.Join(rule.Modifiers, " | ")
		mod, err := ceconv.LoadModifier(mods)
		if err != nil {
			return nil, err
		}

		out, err := mod.Modify(m)
		if err != nil {
			return nil, err
		}

		if len(out) == 0 {
			return nil, fmt.Errorf("no results")
		}

		req, err := http.NewRequest(http.MethodPost, rule.Endpoint, bytes.NewReader([]byte(out[0])))
		if err != nil {
			return nil, err
		}

		for k, v := range r.Header {
			if strings.ToLower(k) != "content-length" {
				for _, x := range v {
					req.Header.Add(k, x)
				}
			}
		}

		return req, nil
	}

	return nil, fmt.Errorf("no rules matched the incoming request payload")
}
