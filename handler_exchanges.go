package main

import (
	"net/http"
)

func (api *API) createExchange(w http.ResponseWriter, r *http.Request) {
	requesterID, err := requireUserID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	var body struct {
		ServiceID int `json:"service_id"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, err)
		return
	}
	ex, err := api.exchanges.Create(r.Context(), requesterID, CreateExchangeInput{
		ServiceID: body.ServiceID,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, ex)
}

func (api *API) listExchanges(w http.ResponseWriter, r *http.Request) {
	actorID, err := requireUserID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	exchanges, err := api.exchanges.List(r.Context(), actorID, r.URL.Query().Get("status"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, exchanges)
}

func (api *API) getExchange(w http.ResponseWriter, r *http.Request) {
	actorID, err := requireUserID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, err)
		return
	}
	ex, err := api.exchanges.GetByID(r.Context(), actorID, id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, ex)
}

func (api *API) acceptExchange(w http.ResponseWriter, r *http.Request) {
	actorID, err := requireUserID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, err)
		return
	}
	ex, err := api.exchanges.Accept(r.Context(), actorID, id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, ex)
}

func (api *API) rejectExchange(w http.ResponseWriter, r *http.Request) {
	actorID, err := requireUserID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, err)
		return
	}
	ex, err := api.exchanges.Reject(r.Context(), actorID, id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, ex)
}

func (api *API) completeExchange(w http.ResponseWriter, r *http.Request) {
	actorID, err := requireUserID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, err)
		return
	}
	ex, err := api.exchanges.Complete(r.Context(), actorID, id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, ex)
}

func (api *API) cancelExchange(w http.ResponseWriter, r *http.Request) {
	actorID, err := requireUserID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, err)
		return
	}
	ex, err := api.exchanges.Cancel(r.Context(), actorID, id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, ex)
}
