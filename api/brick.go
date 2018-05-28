package main

import (
	"database/sql"
	"fmt"
	"log"
)

type Brick struct {
	ID                      string `json:"id,omitempty"`
	ImageStoragePath        string `json:"omit"`
	TreatedImageStoragePath string `json:"omit"`
	ThumbnailStoragePath    string `json:"omit"`
	ImageURL                string `json:"image_url,omitempty"`
	TreatedImageURL         string `json:"treated_url,omitempty"`
	ThumbnailImageURL       string `json:"thumbnail_url,omitempty"`
	CreationDate            string `json:"creationDate,omitempty"`
	ETag                    string `json:"etag,omitempty"`
}

func GetBrick(db *sql.DB, id string) (Brick, error) {
	var (
		imageStoragePath        string
		treatedImageStoragePath string
		thumbnailStoragePath    string
		eTag                    string
		creationDate            string
		brick                   Brick
	)

	stmtOut, err := db.Prepare("select ID, ImageStoragePath, TreatedImageStoragePath, ThumbnailStoragePath, ETag, CreationDate from bricks where ID = ?")
	if err != nil {
		panic(err.Error())
	}
	defer stmtOut.Close()

	rows, err := stmtOut.Query(id)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&id, &imageStoragePath, &treatedImageStoragePath, &thumbnailStoragePath, &eTag, &creationDate)
		if err != nil {
			log.Fatal(err)
		}
		brick = Brick{ID: id,
			ImageStoragePath:        imageStoragePath,
			TreatedImageStoragePath: treatedImageStoragePath,
			ThumbnailStoragePath:    thumbnailStoragePath,
			ImageURL:                fmt.Sprintf("/bricks/%s/original.png", id),
			TreatedImageURL:         fmt.Sprintf("/bricks/%s.png", id),
			ThumbnailImageURL:       fmt.Sprintf("/bricks/%s/thumbnail.png", id),
			ETag:                    eTag,
			CreationDate:            creationDate}
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	return brick, nil
}

func GetBricks(db *sql.DB) []Brick {
	var (
		id                      string
		imageStoragePath        string
		thumbnailStoragePath    string
		treatedImageStoragePath string
		eTag                    string
		creationDate            string
		bricks                  []Brick
	)

	rows, err := db.Query("select ID, ImageStoragePath, TreatedImageStoragePath, ThumbnailStoragePath, ETag, CreationDate from bricks")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&id, &imageStoragePath, &treatedImageStoragePath, &thumbnailStoragePath, &eTag, &creationDate)
		if err != nil {
			log.Fatal(err)
		}

		bricks = append(bricks,
			Brick{ID: id,
				ImageStoragePath:        imageStoragePath,
				TreatedImageStoragePath: treatedImageStoragePath,
				ThumbnailStoragePath:    thumbnailStoragePath,
				ImageURL:                fmt.Sprintf("/bricks/%s/original.png", id),
				TreatedImageURL:         fmt.Sprintf("/bricks/%s.png", id),
				ThumbnailImageURL:       fmt.Sprintf("/bricks/%s/thumbnail.png", id),
				ETag:                    eTag,
				CreationDate:            creationDate})
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	return bricks
}

func SaveBrick(db *sql.DB, brick Brick) {
	var count int
	rows, err := db.Query("select COUNT(*) as count from bricks where ID = ?", brick.ID)
	if err != nil {
		log.Fatal(err)
	}
	for rows.Next() {
		err := rows.Scan(&count)
		if err != nil {
			log.Fatal(err)
		}
	}

	if count > 0 {
		stmtUpd, err := db.Prepare("UPDATE bricks set ImageStoragePath = ?, TreatedImageStoragePath = ?, ThumbnailStoragePath=?, ETag = ?, CreationDate = ? where ID = ?")
		if err != nil {
			panic(err.Error()) // proper error handling instead of panic in your app
		}
		defer stmtUpd.Close() // Close the statement when we leave main() / the program terminates
		_, err = stmtUpd.Exec(brick.ImageStoragePath, brick.TreatedImageStoragePath, brick.ThumbnailStoragePath, brick.ETag, brick.CreationDate, brick.ID)
		if err != nil {
			panic(err.Error()) // proper error handling instead of panic in your app
		}
	} else {
		stmtIns, err := db.Prepare("INSERT INTO bricks VALUES( ?, ?, ?, ?, ?, ? )")
		if err != nil {
			panic(err.Error()) // proper error handling instead of panic in your app
		}
		defer stmtIns.Close() // Close the statement when we leave main() / the program terminates
		_, err = stmtIns.Exec(brick.ID, brick.ImageStoragePath, brick.TreatedImageStoragePath, brick.ThumbnailStoragePath, brick.ETag, brick.CreationDate)
		if err != nil {
			panic(err.Error()) // proper error handling instead of panic in your app
		}
	}
}
