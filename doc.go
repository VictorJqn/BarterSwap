// Package main implémente BarterSwap, une API REST d'échange de compétences
// entre particuliers basée sur un système de crédit-temps.
//
// L'architecture sépare trois couches au sein d'un unique package Go :
// handlers HTTP (présentation), services (logique métier) et sqlStore (PostgreSQL).
//
// L'authentification est simulée par le header X-User-ID sur les routes protégées.
package main
