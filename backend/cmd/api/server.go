package main

import (
	"net/http"
)

func server () error {
	srv := http.Server{
		Addr:	":8080",
		Handler: routes(),
	}

	err := srv.ListenAndServe()
	if err != nil {
		return err
	}

	return nil
}

func simpleHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("simple response"))
}