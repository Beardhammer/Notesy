package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"github.com/filebrowser/filebrowser/v2/board"
	"github.com/filebrowser/filebrowser/v2/errors"
	"github.com/filebrowser/filebrowser/v2/event"
	"github.com/filebrowser/filebrowser/v2/kanban"
)

var boardListHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	boards, err := d.store.Boards.All()
	if err == errors.ErrNotExist {
		return renderJSON(w, r, []*board.Board{})
	}
	if err != nil {
		return http.StatusInternalServerError, err
	}
	return renderJSON(w, r, boards)
})

var boardCreateHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	var b board.Board
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		return http.StatusBadRequest, err
	}
	defer r.Body.Close()

	b.ID = uuid.New().String()
	b.CreatedBy = d.user.Username
	b.CreatedAt = time.Now().Unix()
	b.UpdatedAt = b.CreatedAt

	if err := d.store.Boards.Save(&b); err != nil {
		return http.StatusInternalServerError, err
	}

	return renderJSON(w, r, b)
})

var boardUpdateHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	id := mux.Vars(r)["boardId"]

	existing, err := d.store.Boards.GetByID(id)
	if err != nil {
		return errToStatus(err), err
	}

	var b board.Board
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		return http.StatusBadRequest, err
	}
	defer r.Body.Close()

	existing.Name = b.Name
	existing.UpdatedAt = time.Now().Unix()

	if err := d.store.Boards.Save(existing); err != nil {
		return http.StatusInternalServerError, err
	}

	return renderJSON(w, r, existing)
})

var boardDeleteHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	id := mux.Vars(r)["boardId"]

	// Delete all tasks and events for this board
	if err := d.store.Kanban.DeleteByBoard(id); err != nil {
		return http.StatusInternalServerError, err
	}
	if err := d.store.Events.DeleteByBoard(id); err != nil {
		return http.StatusInternalServerError, err
	}

	if err := d.store.Boards.Delete(id); err != nil {
		return errToStatus(err), err
	}

	return http.StatusOK, nil
})

// Board-scoped kanban handlers

var boardKanbanListHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	boardID := mux.Vars(r)["boardId"]

	tasks, err := d.store.Kanban.AllByBoard(boardID)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	return renderJSON(w, r, tasks)
})

var boardKanbanAllHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	tasks, err := d.store.Kanban.All()
	if err == errors.ErrNotExist {
		return renderJSON(w, r, []*kanban.Task{})
	}
	if err != nil {
		return http.StatusInternalServerError, err
	}

	// Enrich tasks with board name
	boards, _ := d.store.Boards.All()
	boardMap := make(map[string]string)
	for _, b := range boards {
		boardMap[b.ID] = b.Name
	}

	type taskWithBoard struct {
		*kanban.Task
		BoardName string `json:"boardName"`
	}

	result := make([]taskWithBoard, len(tasks))
	for i, t := range tasks {
		result[i] = taskWithBoard{Task: t, BoardName: boardMap[t.BoardID]}
	}

	return renderJSON(w, r, result)
})

var boardKanbanGetHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	vars := mux.Vars(r)
	id, err := parseUint(vars["id"])
	if err != nil {
		return http.StatusBadRequest, err
	}

	task, err := d.store.Kanban.GetByID(id)
	if err != nil {
		return errToStatus(err), err
	}

	return renderJSON(w, r, task)
})

var boardKanbanPostHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	boardID := mux.Vars(r)["boardId"]

	var task kanban.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		return http.StatusBadRequest, err
	}
	defer r.Body.Close()

	task.ID = 0
	task.BoardID = boardID
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

var boardKanbanPutHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	vars := mux.Vars(r)
	id, err := parseUint(vars["id"])
	if err != nil {
		return http.StatusBadRequest, err
	}

	existing, err := d.store.Kanban.GetByID(id)
	if err != nil {
		return errToStatus(err), err
	}

	var task kanban.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		return http.StatusBadRequest, err
	}
	defer r.Body.Close()

	task.ID = id
	task.BoardID = existing.BoardID
	task.UpdatedAt = time.Now().Unix()

	if err := d.store.Kanban.Save(&task); err != nil {
		return http.StatusInternalServerError, err
	}

	return renderJSON(w, r, task)
})

var boardKanbanDeleteHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	vars := mux.Vars(r)
	id, err := parseUint(vars["id"])
	if err != nil {
		return http.StatusBadRequest, err
	}

	if err := d.store.Kanban.Delete(id); err != nil {
		return errToStatus(err), err
	}

	return http.StatusOK, nil
})

// Board-scoped event handlers

var boardEventListHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	boardID := mux.Vars(r)["boardId"]

	events, err := d.store.Events.AllByBoard(boardID)
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

var boardEventAllHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	events, err := d.store.Events.All()
	if err == errors.ErrNotExist {
		return renderJSON(w, r, []*event.Event{})
	}
	if err != nil {
		return http.StatusInternalServerError, err
	}

	boards, _ := d.store.Boards.All()
	boardMap := make(map[string]string)
	for _, b := range boards {
		boardMap[b.ID] = b.Name
	}

	type eventWithBoard struct {
		*event.Event
		BoardName string `json:"boardName"`
	}

	result := make([]eventWithBoard, len(events))
	for i, e := range events {
		result[i] = eventWithBoard{Event: e, BoardName: boardMap[e.BoardID]}
	}

	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")

	if from != "" && to != "" {
		var filtered []eventWithBoard
		for _, e := range result {
			if e.Date >= from && e.Date <= to {
				filtered = append(filtered, e)
			}
		}
		return renderJSON(w, r, filtered)
	}

	return renderJSON(w, r, result)
})

var boardEventGetHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	vars := mux.Vars(r)
	id, err := parseUint(vars["id"])
	if err != nil {
		return http.StatusBadRequest, err
	}

	ev, err := d.store.Events.GetByID(id)
	if err != nil {
		return errToStatus(err), err
	}

	return renderJSON(w, r, ev)
})

var boardEventPostHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	boardID := mux.Vars(r)["boardId"]

	var ev event.Event
	if err := json.NewDecoder(r.Body).Decode(&ev); err != nil {
		return http.StatusBadRequest, err
	}
	defer r.Body.Close()

	ev.ID = 0
	ev.BoardID = boardID
	ev.CreatedBy = d.user.Username
	ev.CreatedAt = time.Now().Unix()
	ev.UpdatedAt = ev.CreatedAt

	if err := d.store.Events.Save(&ev); err != nil {
		return http.StatusInternalServerError, err
	}

	return renderJSON(w, r, ev)
})

var boardEventPutHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	vars := mux.Vars(r)
	id, err := parseUint(vars["id"])
	if err != nil {
		return http.StatusBadRequest, err
	}

	existing, err := d.store.Events.GetByID(id)
	if err != nil {
		return errToStatus(err), err
	}

	var ev event.Event
	if err := json.NewDecoder(r.Body).Decode(&ev); err != nil {
		return http.StatusBadRequest, err
	}
	defer r.Body.Close()

	ev.ID = id
	ev.BoardID = existing.BoardID
	ev.UpdatedAt = time.Now().Unix()

	if err := d.store.Events.Save(&ev); err != nil {
		return http.StatusInternalServerError, err
	}

	return renderJSON(w, r, ev)
})

var boardEventDeleteHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	vars := mux.Vars(r)
	id, err := parseUint(vars["id"])
	if err != nil {
		return http.StatusBadRequest, err
	}

	if err := d.store.Events.Delete(id); err != nil {
		return errToStatus(err), err
	}

	return http.StatusOK, nil
})

// parseUint is a helper to parse route vars as uint.
func parseUint(s string) (uint, error) {
	v, err := strconv.ParseUint(s, 10, 0)
	return uint(v), err
}
