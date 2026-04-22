package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	"github.com/filebrowser/filebrowser/v2/errors"
	"github.com/filebrowser/filebrowser/v2/event"
)

var eventListHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	events, err := d.store.Events.All()
	if err == errors.ErrNotExist {
		return renderJSON(w, r, []*event.Event{})
	}
	if err != nil {
		return http.StatusInternalServerError, err
	}

	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")

	if from != "" && to != "" {
		var filtered []*event.Event
		for _, e := range events {
			if e.Date >= from && e.Date <= to {
				filtered = append(filtered, e)
			}
		}
		return renderJSON(w, r, filtered)
	}

	return renderJSON(w, r, events)
})

var eventGetHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 0)
	if err != nil {
		return http.StatusBadRequest, err
	}

	ev, err := d.store.Events.GetByID(uint(id))
	if err != nil {
		return errToStatus(err), err
	}

	return renderJSON(w, r, ev)
})

var eventPostHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	var ev event.Event
	if err := json.NewDecoder(r.Body).Decode(&ev); err != nil {
		return http.StatusBadRequest, err
	}
	defer r.Body.Close()

	ev.ID = 0
	ev.CreatedBy = d.user.Username
	ev.CreatedAt = time.Now().Unix()
	ev.UpdatedAt = ev.CreatedAt

	if err := d.store.Events.Save(&ev); err != nil {
		return http.StatusInternalServerError, err
	}

	return renderJSON(w, r, ev)
})

var eventPutHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 0)
	if err != nil {
		return http.StatusBadRequest, err
	}

	_, err = d.store.Events.GetByID(uint(id))
	if err != nil {
		return errToStatus(err), err
	}

	var ev event.Event
	if err := json.NewDecoder(r.Body).Decode(&ev); err != nil {
		return http.StatusBadRequest, err
	}
	defer r.Body.Close()

	ev.ID = uint(id)
	ev.UpdatedAt = time.Now().Unix()

	if err := d.store.Events.Save(&ev); err != nil {
		return http.StatusInternalServerError, err
	}

	return renderJSON(w, r, ev)
})

var eventDeleteHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 0)
	if err != nil {
		return http.StatusBadRequest, err
	}

	err = d.store.Events.Delete(uint(id))
	if err != nil {
		return errToStatus(err), err
	}

	return http.StatusOK, nil
})
