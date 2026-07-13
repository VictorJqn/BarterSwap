package main

import (
	"net/http"
)

func (api *API) listServices(w http.ResponseWriter, r *http.Request) {
	services, err := api.services.List(r.Context(), ListServicesInput{
		Categorie: r.URL.Query().Get("categorie"),
		Ville:     r.URL.Query().Get("ville"),
		Search:    r.URL.Query().Get("search"),
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, services)
}

func (api *API) createService(w http.ResponseWriter, r *http.Request) {
	providerID, err := requireUserID(r)
	if err != nil {
		writeError(w, err)
		return
	}
	var body struct {
		Titre        string `json:"titre"`
		Description  string `json:"description"`
		Categorie    string `json:"categorie"`
		DureeMinutes int    `json:"duree_minutes"`
		Credits      int    `json:"credits"`
		Ville        string `json:"ville"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, err)
		return
	}
	svc, err := api.services.Create(r.Context(), providerID, CreateServiceInput{
		Titre:        body.Titre,
		Description:  body.Description,
		Categorie:    body.Categorie,
		DureeMinutes: body.DureeMinutes,
		Credits:      body.Credits,
		Ville:        body.Ville,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, svc)
}

func (api *API) getService(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, err)
		return
	}
	svc, err := api.services.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, svc)
}

func (api *API) updateService(w http.ResponseWriter, r *http.Request) {
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
	var body struct {
		Titre        string `json:"titre"`
		Description  string `json:"description"`
		Categorie    string `json:"categorie"`
		DureeMinutes int    `json:"duree_minutes"`
		Credits      int    `json:"credits"`
		Ville        string `json:"ville"`
		Actif        bool   `json:"actif"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, err)
		return
	}
	svc, err := api.services.Update(r.Context(), actorID, id, UpdateServiceInput{
		Titre:        body.Titre,
		Description:  body.Description,
		Categorie:    body.Categorie,
		DureeMinutes: body.DureeMinutes,
		Credits:      body.Credits,
		Ville:        body.Ville,
		Actif:        body.Actif,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, svc)
}

func (api *API) deleteService(w http.ResponseWriter, r *http.Request) {
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
	if err := api.services.Delete(r.Context(), actorID, id); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
