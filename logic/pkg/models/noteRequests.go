package models

type DeleteNoteRequest struct {
	NoteID            string `json:"noteid"`
	DeletePermanently bool   `json:"deletePermanently"`
}
