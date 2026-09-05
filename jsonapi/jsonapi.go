package jsonapi

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log"
	"mailingList/mdb"
	"net/http"
	"strconv"
	"time"
)

func setJsonHeader(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
}

func fromJson[T any](body io.Reader, target T) error {
	decoder := json.NewDecoder(body)
	return decoder.Decode(&target)
}

func returnJson[T any](w http.ResponseWriter, withData func() (T, error)) {
	setJsonHeader(w)

	data, serverErr := withData()

	if serverErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		errPayload := struct {
			Err string `json:"error"`
		}{
			Err: serverErr.Error(),
		}
		errJson, err := json.Marshal(&errPayload)
		if err != nil {
			log.Print(err)
			return
		}
		w.Write(errJson)
		return
	}

	dataJson, err := json.Marshal(&data)
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(dataJson)
}

func returnErr(w http.ResponseWriter, err error, code int) {
	setJsonHeader(w)
	w.WriteHeader(code)
	errorMessage := struct {
		Err string `json:"error"`
	}{
		Err: err.Error(),
	}
	errJson, jsonErr := json.Marshal(&errorMessage)
	if jsonErr != nil {
		log.Print(jsonErr)
		return
	}
	w.Write(errJson)
}

func methodNotAllowed(w http.ResponseWriter) {
	returnErr(w, errors.New("method not allowed"), http.StatusMethodNotAllowed)
}

func CreateEmail(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			methodNotAllowed(w)
			return
		}

		entry := mdb.EmailEntry{}
		if err := fromJson(req.Body, &entry); err != nil {
			returnErr(w, errors.New("invalid JSON body"), http.StatusBadRequest)
			return
		}

		if err := mdb.CreateEmail(db, entry.Email); err != nil {
			returnErr(w, err, http.StatusBadRequest)
			return
		}

		returnJson(w, func() (interface{}, error) {
			log.Printf("JSON CreateEmail: %v\n", entry.Email)
			return mdb.GetEmail(db, entry.Email)
		})
	})
}

func GetEmail(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			methodNotAllowed(w)
			return
		}

		email := req.URL.Query().Get("email")
		if email == "" {
			returnErr(w, errors.New("email query parameter is required"), http.StatusBadRequest)
			return
		}

		returnJson(w, func() (interface{}, error) {
			log.Printf("JSON GetEmail: %v\n", email)
			return mdb.GetEmail(db, email)
		})
	})
}

func GetEmailBatch(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			methodNotAllowed(w)
			return
		}

		queryParams := req.URL.Query()
		page, err := strconv.Atoi(queryParams.Get("page"))
		if err != nil || page <= 0 {
			returnErr(w, errors.New("page query parameter is required and must be a positive integer"), http.StatusBadRequest)
			return
		}
		count, err := strconv.Atoi(queryParams.Get("count"))
		if err != nil || count <= 0 {
			returnErr(w, errors.New("count query parameter is required and must be a positive integer"), http.StatusBadRequest)
			return
		}

		queryOptions := mdb.GetEmailBatchQueryParams{
			Page:  page,
			Count: count,
		}

		returnJson(w, func() (interface{}, error) {
			log.Printf("JSON GetEmailBatch: %v\n", queryOptions)
			return mdb.GetEmailBatch(db, queryOptions)
		})
	})
}

func UpdateEmail(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPut {
			methodNotAllowed(w)
			return
		}

		entry := mdb.EmailEntry{}
		if err := fromJson(req.Body, &entry); err != nil {
			returnErr(w, errors.New("invalid JSON body"), http.StatusBadRequest)
			return
		}

		if err := mdb.UpdateEmail(db, entry); err != nil {
			returnErr(w, err, http.StatusBadRequest)
			return
		}

		returnJson(w, func() (interface{}, error) {
			log.Printf("JSON UpdateEmail: %v\n", entry.Email)
			return mdb.GetEmail(db, entry.Email)
		})
	})
}

func DeleteEmail(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodDelete {
			methodNotAllowed(w)
			return
		}

		entry := mdb.EmailEntry{}
		if err := fromJson(req.Body, &entry); err != nil {
			returnErr(w, errors.New("invalid JSON body"), http.StatusBadRequest)
			return
		}

		if err := mdb.DeleteEmail(db, entry.Email); err != nil {
			returnErr(w, err, http.StatusBadRequest)
			return
		}

		returnJson(w, func() (interface{}, error) {
			log.Printf("JSON DeleteEmail: %v\n", entry.Email)
			return mdb.GetEmail(db, entry.Email)
		})
	})
}

func Serve(ctx context.Context, db *sql.DB, bind string) {
	mux := http.NewServeMux()
	mux.Handle("/email/create", CreateEmail(db))
	mux.Handle("/email/get", GetEmail(db))
	mux.Handle("/email/get_batch", GetEmailBatch(db))
	mux.Handle("/email/update", UpdateEmail(db))
	mux.Handle("/email/delete", DeleteEmail(db))

	server := &http.Server{
		Addr:    bind,
		Handler: mux,
	}

	// Graceful shutdown when context is cancelled.
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		log.Println("JSON API server shutting down...")
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("JSON server shutdown error: %v", err)
		}
	}()

	log.Printf("JSON API server listening on %v\n", bind)
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("JSON server error: %v", err)
	}
}