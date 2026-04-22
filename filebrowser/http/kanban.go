package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	"github.com/filebrowser/filebrowser/v2/errors"
	"github.com/filebrowser/filebrowser/v2/kanban"
)

var kanbanListHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	tasks, err := d.store.Kanban.All()
	if err == errors.ErrNotExist {
		return renderJSON(w, r, []*kanban.Task{})
	}
	if err != nil {
		return http.StatusInternalServerError, err
	}

	return renderJSON(w, r, tasks)
})

var kanbanGetHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 0)
	if err != nil {
		return http.StatusBadRequest, err
	}

	task, err := d.store.Kanban.GetByID(uint(id))
	if err != nil {
		return errToStatus(err), err
	}

	return renderJSON(w, r, task)
})

var kanbanPostHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	var task kanban.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		return http.StatusBadRequest, err
	}
	defer r.Body.Close()

	task.ID = 0
	task.CreatedBy = d.user.Username
	task.CreatedAt = time.Now().Unix()
	task.UpdatedAt = task.CreatedAt

	if task.Column == "" {
		task.Column = "todo"
	}

	if err := d.store.Kanban.Save(&task); err != nil {
		return http.StatusInternalServerError, err
	}

	return renderJSON(w, r, task)
})

var kanbanPutHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 0)
	if err != nil {
		return http.StatusBadRequest, err
	}

	_, err = d.store.Kanban.GetByID(uint(id))
	if err != nil {
		return errToStatus(err), err
	}

	var task kanban.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		return http.StatusBadRequest, err
	}
	defer r.Body.Close()

	task.ID = uint(id)
	task.UpdatedAt = time.Now().Unix()

	if err := d.store.Kanban.Save(&task); err != nil {
		return http.StatusInternalServerError, err
	}

	return renderJSON(w, r, task)
})

var kanbanDeleteHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 0)
	if err != nil {
		return http.StatusBadRequest, err
	}

	err = d.store.Kanban.Delete(uint(id))
	if err != nil {
		return errToStatus(err), err
	}

	return http.StatusOK, nil
})
