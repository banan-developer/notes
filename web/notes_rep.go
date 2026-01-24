package main

import "database/sql"

type MySQLNoteRep struct {
	db *sql.DB
}

func NewMySQLNoteRepository(db *sql.DB) *MySQLNoteRep {
	return &MySQLNoteRep{db: db}
}

// создание заметки
func (r *MySQLNoteRep) Create(content string, userID int) (*Note, error) {
	result, err := r.db.Exec(
		"INSERT INTO notes (content, user_id) VALUES (?, ?)",
		content,
		userID,
	)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &Note{
		ID:      int(id),
		Content: content,
		UserID:  userID,
	}, nil
}

// получение всех заметок
func (r *MySQLNoteRep) GetallByUser(userID int) ([]Note, error) {
	rows, err := r.db.Query(
		"SELECT id, content FROM notes WHERE user_id = ? ORDER BY id DESC",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []Note

	for rows.Next() {
		var note Note
		if err := rows.Scan(&note.ID, &note.Content); err != nil {
			return nil, err
		}
		note.UserID = userID
		notes = append(notes, note)
	}

	return notes, nil
}

func (r *MySQLNoteRep) Delete(noteID int, userID int) error {
	_, err := r.db.Exec(
		"DELETE FROM notes WHERE id = ? AND user_id = ?",
		noteID,
		userID,
	)
	return err
}
