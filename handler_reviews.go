package main

import (
	"net/http"
)

func (api *API) createReview(w http.ResponseWriter, r *http.Request) {
	authorID, err := requireUserID(r)
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
		Note        int    `json:"note"`
		Commentaire string `json:"commentaire"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, err)
		return
	}
	review, err := api.reviews.Create(r.Context(), authorID, id, CreateReviewInput{
		Note:        body.Note,
		Commentaire: body.Commentaire,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, review)
}
