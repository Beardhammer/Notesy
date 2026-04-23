package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/filebrowser/filebrowser/v2/shout"
)

const maxShoutBodyLen = 2000

type shoutPostRequest struct {
	Body string `json:"body"`
}

var shoutPostHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	var req shoutPostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return http.StatusBadRequest, err
	}
	defer r.Body.Close()

	body := strings.TrimSpace(req.Body)
	if body == "" {
		return http.StatusBadRequest, nil
	}
	if len(body) > maxShoutBodyLen {
		return http.StatusBadRequest, nil
	}

	m := &shout.Message{
		Author:    d.user.Username,
		Body:      body,
		CreatedAt: time.Now().Unix(),
	}
	if err := d.store.Shouts.Save(m); err != nil {
		return http.StatusInternalServerError, err
	}
	d.store.Shouts.Hub().Broadcast(m)
	return renderJSON(w, r, m)
})

var shoutListHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	msgs, err := d.store.Shouts.All()
	if err != nil {
		return http.StatusInternalServerError, err
	}
	if msgs == nil {
		msgs = []*shout.Message{}
	}
	return renderJSON(w, r, msgs)
})

var shoutStreamHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return http.StatusInternalServerError, nil
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	flusher.Flush()

	// Backfill from Last-Event-ID header.
	var lastID uint
	if h := r.Header.Get("Last-Event-ID"); h != "" {
		if id, err := strconv.ParseUint(h, 10, 0); err == nil {
			lastID = uint(id)
		}
	}
	if lastID > 0 {
		backfill, err := d.store.Shouts.Since(lastID)
		if err == nil {
			for _, m := range backfill {
				if err := writeSSE(w, m); err != nil {
					return 0, nil
				}
				flusher.Flush()
			}
		}
	}

	client := d.store.Shouts.Hub().Subscribe()
	defer d.store.Shouts.Hub().Unsubscribe(client)

	ping := time.NewTicker(30 * time.Second)
	defer ping.Stop()

	for {
		select {
		case m, ok := <-client.Ch:
			if !ok {
				return 0, nil
			}
			if err := writeSSE(w, m); err != nil {
				return 0, nil
			}
			flusher.Flush()
		case <-ping.C:
			_, _ = w.Write([]byte(": ping\n\n"))
			flusher.Flush()
		case <-r.Context().Done():
			return 0, nil
		}
	}
})

func writeSSE(w http.ResponseWriter, m *shout.Message) error {
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	_, err = w.Write([]byte("id: " + strconv.FormatUint(uint64(m.ID), 10) + "\ndata: " + string(data) + "\n\n"))
	return err
}
