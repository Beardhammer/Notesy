package http

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
)


type DrawIOCallback struct {
	Diagram   string   `json:"diagram"`
}

var drawIOCallbackHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	// Make sure we can read the stream correctly
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	defer r.Body.Close()


	var data DrawIOCallback
	err = json.Unmarshal(body, &data)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	// Convert XML to a reader
	bodyReader := strings.NewReader(data.Diagram)

	// get the file to save from the ?save param
	docPath := r.URL.Query().Get("save")
	if docPath == "" {
		return http.StatusInternalServerError, errors.New("unable to get file save path")
	}

	//  Ensure we can write/edit file
	if !d.user.Perm.Modify || !d.Check(docPath) {
		return http.StatusForbidden, nil
	}

	// Verify hooks and write the file
	err = d.RunHook(func() error {
		_, writeErr := writeFile(d.user.Fs, docPath, bodyReader)
		if writeErr != nil {
			return writeErr
		}
		return nil
	}, "save", docPath, "", d.user)



	if err != nil {
		return http.StatusInternalServerError, err
	}

	// Generate a response back
	resp := map[string]int{
		"error": 0,
	}

	return renderJSON(w, r, resp)
})
