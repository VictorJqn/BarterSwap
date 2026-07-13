package main

import (
	"net/http"
)

func (api *API) createUser(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Pseudo string `json:"pseudo"`
		Bio    string `json:"bio"`
		Ville  string `json:"ville"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, err)
		return
	}
	user, err := api.users.Create(r.Context(), CreateUserInput{
		Pseudo: body.Pseudo,
		Bio:    body.Bio,
		Ville:  body.Ville,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, user)
}

func (api *API) getUser(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, err)
		return
	}
	user, err := api.users.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, user)
}

func (api *API) updateUser(w http.ResponseWriter, r *http.Request) {
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
		Pseudo string `json:"pseudo"`
		Bio    string `json:"bio"`
		Ville  string `json:"ville"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, err)
		return
	}
	user, err := api.users.Update(r.Context(), actorID, id, UpdateUserInput{
		Pseudo: body.Pseudo,
		Bio:    body.Bio,
		Ville:  body.Ville,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, user)
}

func (api *API) getSkills(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		writeError(w, err)
		return
	}
	skills, err := api.users.GetSkills(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, skills)
}

func (api *API) replaceSkills(w http.ResponseWriter, r *http.Request) {
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
	var skills []Skill
	if err := decodeJSON(r, &skills); err != nil {
		writeError(w, err)
		return
	}
	if err := api.users.ReplaceSkills(r.Context(), actorID, id, skills); err != nil {
		writeError(w, err)
		return
	}
	updated, err := api.users.GetSkills(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}
