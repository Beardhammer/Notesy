package http

import (
	"io/fs"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/storage"
)

type modifyRequest struct {
	What  string   `json:"what"`  // Answer to: what data type?
	Which []string `json:"which"` // Answer to: which fields?
}

func NewHandler(
	imgSvc ImgService,
	fileCache FileCache,
	store *storage.Storage,
	server *settings.Server,
	assetsFs fs.FS,
) (http.Handler, error) {
	server.Clean()

	r := mux.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Security-Policy", `default-src 'self'; style-src 'unsafe-inline';`)
			next.ServeHTTP(w, r)
		})
	})
	index, static := getStaticHandlers(store, server, assetsFs)

	// NOTE: This fixes the issue where it would redirect if people did not put a
	// trailing slash in the end. I hate this decision since this allows some awful
	// URLs https://www.gorillatoolkit.org/pkg/mux#Router.SkipClean
	r = r.SkipClean(true)

	monkey := func(fn handleFunc, prefix string) http.Handler {
		return handle(fn, prefix, store, server)
	}

	r.HandleFunc("/health", healthHandler)
	r.PathPrefix("/static").Handler(static)
	r.NotFoundHandler = index

	api := r.PathPrefix("/api").Subrouter()

	tokenExpirationTime := server.GetTokenExpirationTime(DefaultTokenExpirationTime)
	api.Handle("/login", monkey(loginHandler(tokenExpirationTime), ""))
	api.Handle("/signup", monkey(signupHandler, ""))
	api.Handle("/renew", monkey(renewHandler(tokenExpirationTime), ""))

	users := api.PathPrefix("/users").Subrouter()
	users.Handle("", monkey(usersGetHandler, "")).Methods("GET")
	users.Handle("", monkey(userPostHandler, "")).Methods("POST")
	users.Handle("/{id:[0-9]+}", monkey(userPutHandler, "")).Methods("PUT")
	users.Handle("/{id:[0-9]+}", monkey(userGetHandler, "")).Methods("GET")
	users.Handle("/{id:[0-9]+}", monkey(userDeleteHandler, "")).Methods("DELETE")

	api.Path("/onlyoffice/callback").Handler(monkey(onlyofficeCallbackHandler, "/api/onlyoffice/callback")).Methods("POST")
	api.Path("/onlyoffice/token").Handler(monkey(onlyofficeTokenHandler, "/api/onlyoffice/token")).Methods("POST")
	api.PathPrefix("/onlyoffice/dl").Handler(monkey(rawAuthHandler, "/api/onlyoffice/dl")).Methods("GET")
	api.PathPrefix("/onlyoffice/save").Handler(monkey(onlyofficeCallbackAuthHandler, "/api/onlyoffice/save")).Methods("POST")
	api.PathPrefix("/drawio").Handler(monkey(drawIOCallbackHandler, "/api/drawio/callback")).Methods("POST")

	api.PathPrefix("/resources").Handler(monkey(resourceGetHandler, "/api/resources")).Methods("GET")
	api.PathPrefix("/resources").Handler(monkey(resourceDeleteHandler(fileCache), "/api/resources")).Methods("DELETE")
	api.PathPrefix("/resources").Handler(monkey(resourcePostHandler(fileCache), "/api/resources")).Methods("POST")
	api.PathPrefix("/resources").Handler(monkey(resourcePutHandler, "/api/resources")).Methods("PUT")
	api.PathPrefix("/resources").Handler(monkey(resourcePatchHandler(fileCache), "/api/resources")).Methods("PATCH")

	api.PathPrefix("/tus").Handler(monkey(tusPostHandler(), "/api/tus")).Methods("POST")
	api.PathPrefix("/tus").Handler(monkey(tusHeadHandler(), "/api/tus")).Methods("HEAD", "GET")
	api.PathPrefix("/tus").Handler(monkey(tusPatchHandler(), "/api/tus")).Methods("PATCH")

	api.PathPrefix("/usage").Handler(monkey(diskUsage, "/api/usage")).Methods("GET")

	api.Path("/shares").Handler(monkey(shareListHandler, "/api/shares")).Methods("GET")
	api.PathPrefix("/share").Handler(monkey(shareGetsHandler, "/api/share")).Methods("GET")
	api.PathPrefix("/share").Handler(monkey(sharePostHandler, "/api/share")).Methods("POST")
	api.PathPrefix("/share").Handler(monkey(shareDeleteHandler, "/api/share")).Methods("DELETE")

	// Board CRUD
	api.Handle("/boards", monkey(boardListHandler, "")).Methods("GET")
	api.Handle("/boards", monkey(boardCreateHandler, "")).Methods("POST")
	api.Handle("/boards/{boardId}", monkey(boardUpdateHandler, "")).Methods("PUT")
	api.Handle("/boards/{boardId}", monkey(boardDeleteHandler, "")).Methods("DELETE")

	// Board-scoped kanban
	api.Handle("/boards/all/kanban", monkey(boardKanbanAllHandler, "")).Methods("GET")
	api.Handle("/boards/{boardId}/kanban", monkey(boardKanbanListHandler, "")).Methods("GET")
	api.Handle("/boards/{boardId}/kanban", monkey(boardKanbanPostHandler, "")).Methods("POST")
	api.Handle("/boards/{boardId}/kanban/{id:[0-9]+}", monkey(boardKanbanGetHandler, "")).Methods("GET")
	api.Handle("/boards/{boardId}/kanban/{id:[0-9]+}", monkey(boardKanbanPutHandler, "")).Methods("PUT")
	api.Handle("/boards/{boardId}/kanban/{id:[0-9]+}", monkey(boardKanbanDeleteHandler, "")).Methods("DELETE")

	// Board-scoped events
	api.Handle("/boards/all/events", monkey(boardEventAllHandler, "")).Methods("GET")
	api.Handle("/boards/{boardId}/events", monkey(boardEventListHandler, "")).Methods("GET")
	api.Handle("/boards/{boardId}/events", monkey(boardEventPostHandler, "")).Methods("POST")
	api.Handle("/boards/{boardId}/events/{id:[0-9]+}", monkey(boardEventGetHandler, "")).Methods("GET")
	api.Handle("/boards/{boardId}/events/{id:[0-9]+}", monkey(boardEventPutHandler, "")).Methods("PUT")
	api.Handle("/boards/{boardId}/events/{id:[0-9]+}", monkey(boardEventDeleteHandler, "")).Methods("DELETE")

	// Legacy routes (keep for backward compat)
	api.Handle("/kanban", monkey(kanbanListHandler, "")).Methods("GET")
	api.Handle("/kanban/{id:[0-9]+}", monkey(kanbanGetHandler, "")).Methods("GET")
	api.Handle("/kanban", monkey(kanbanPostHandler, "")).Methods("POST")
	api.Handle("/kanban/{id:[0-9]+}", monkey(kanbanPutHandler, "")).Methods("PUT")
	api.Handle("/kanban/{id:[0-9]+}", monkey(kanbanDeleteHandler, "")).Methods("DELETE")

	api.Handle("/events", monkey(eventListHandler, "")).Methods("GET")
	api.Handle("/events/{id:[0-9]+}", monkey(eventGetHandler, "")).Methods("GET")
	api.Handle("/events", monkey(eventPostHandler, "")).Methods("POST")
	api.Handle("/events/{id:[0-9]+}", monkey(eventPutHandler, "")).Methods("PUT")
	api.Handle("/events/{id:[0-9]+}", monkey(eventDeleteHandler, "")).Methods("DELETE")

	api.Handle("/settings", monkey(settingsGetHandler, "")).Methods("GET")
	api.Handle("/settings", monkey(settingsPutHandler, "")).Methods("PUT")

	api.PathPrefix("/raw").Handler(monkey(rawHandler, "/api/raw")).Methods("GET")
	api.PathPrefix("/preview/{size}/{path:.*}").
		Handler(monkey(previewHandler(imgSvc, fileCache, server.EnableThumbnails, server.ResizePreview), "/api/preview")).Methods("GET")
	api.PathPrefix("/command").Handler(monkey(commandsHandler, "/api/command")).Methods("GET")
	api.PathPrefix("/search").Handler(monkey(searchHandler, "/api/search")).Methods("GET")

	public := api.PathPrefix("/public").Subrouter()
	public.PathPrefix("/dl").Handler(monkey(publicDlHandler, "/api/public/dl/")).Methods("GET")
	public.PathPrefix("/share").Handler(monkey(publicShareHandler, "/api/public/share/")).Methods("GET")

	return stripPrefix(server.BaseURL, r), nil
}
