package main

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"golang.org/x/image/draw"
)

var shared_db *sql.DB

type Brick struct {
	ID                   string `json:"id,omitempty"`
	ImageStoragePath     string `json:"url,omitempty"`
	ThumbnailStoragePath string `json:"thumbnail_url,omitempty"`
	CreationDate         string `json: "creationDate,omitempty"`
	ETag                 string `json: "etag,omitempty"`
}

type ImageDataPackage struct {
	ImageData string `json:"imagedata,omitempty"`
}

type PublishBrickResponse struct {
	Result string `json:"result,omitempty"`
}

func get_brick(db *sql.DB, id string) Brick {
	var (
		imageStoragePath     string
		thumbnailStoragePath string
		eTag                 string
		creationDate         string
		brick                Brick
	)

	stmtOut, err := db.Prepare("select ID, ImageStoragePath, ThumbnailStoragePath, ETag, CreationDate from bricks where ID = ?")
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
		err := rows.Scan(&id, &imageStoragePath, &thumbnailStoragePath, &eTag, &creationDate)
		if err != nil {
			log.Fatal(err)
		}
		brick = Brick{ID: id, ImageStoragePath: imageStoragePath, ThumbnailStoragePath: thumbnailStoragePath, ETag: eTag, CreationDate: creationDate}
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	return brick
}

func get_bricks(db *sql.DB) []Brick {
	var (
		id                   string
		imageStoragePath     string
		thumbnailStoragePath string
		eTag                 string
		creationDate         string
		bricks               []Brick
	)

	rows, err := db.Query("select ID, ImageStoragePath, ThumbnailStoragePath, ETag, CreationDate from bricks")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&id, &imageStoragePath, &thumbnailStoragePath, &eTag, &creationDate)
		if err != nil {
			log.Fatal(err)
		}

		bricks = append(bricks, Brick{ID: id, ImageStoragePath: imageStoragePath, ThumbnailStoragePath: thumbnailStoragePath, ETag: eTag, CreationDate: creationDate})
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	return bricks
}

func upsert_brick(db *sql.DB, brick Brick) {
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
		stmtUpd, err := db.Prepare("UPDATE bricks set ImageStoragePath = ?, ThumbnailStoragePath=?, ETag = ?, CreationDate = ? where ID = ?")
		if err != nil {
			panic(err.Error()) // proper error handling instead of panic in your app
		}
		defer stmtUpd.Close() // Close the statement when we leave main() / the program terminates
		_, err = stmtUpd.Exec(brick.ImageStoragePath, brick.ThumbnailStoragePath, brick.ETag, brick.CreationDate, brick.ID)
		if err != nil {
			panic(err.Error()) // proper error handling instead of panic in your app
		}
	} else {
		stmtIns, err := db.Prepare("INSERT INTO bricks VALUES( ?, ?, ?, ?, ? )")
		if err != nil {
			panic(err.Error()) // proper error handling instead of panic in your app
		}
		defer stmtIns.Close() // Close the statement when we leave main() / the program terminates
		_, err = stmtIns.Exec(brick.ID, brick.ImageStoragePath, brick.ThumbnailStoragePath, brick.ETag, brick.CreationDate)
		if err != nil {
			panic(err.Error()) // proper error handling instead of panic in your app
		}
	}
}

func mysql_prepare_exec(db *sql.DB, command string) {
	stmtOut, err := db.Prepare(command)
	if err != nil {
		panic(err.Error())
	}
	defer stmtOut.Close()

	stmtOut.Exec()
}

func mysql_exec(db *sql.DB, command string) {
	_, err := db.Exec(command)
	if err != nil {
		panic(err.Error())
	}
}

func create_database() {
	s := []string{os.Getenv("MYSQL_USERNAME"), ":", os.Getenv("MYSQL_PASSWORD"), "@tcp(", os.Getenv("MYSQL_HOST"), ":", os.Getenv("MYSQL_PORT"), ")/"}
	log.Print(strings.Join(s, ""))
	conn, err := sql.Open("mysql", strings.Join(s, ""))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	mysql_prepare_exec(conn, `CREATE DATABASE IF NOT EXISTS wall`)
	mysql_exec(conn, `USE wall`)

	mysql_prepare_exec(conn, `CREATE TABLE IF NOT EXISTS `+"`bricks`"+` (
    `+"`ID`"+` VARCHAR(16) NOT NULL,
		`+"`ImageStoragePath`"+` VARCHAR(1024),
		`+"`ThumbnailStoragePath`"+` VARCHAR(1024),
    `+"`ETag`"+` VARCHAR(1024),
    `+"`CreationDate`"+` DATETIME,
    PRIMARY KEY (`+"`ID`"+`)
  );`)
	conn.Close()
}

func main() {
	create_database()
	s := []string{os.Getenv("MYSQL_USERNAME"), ":", os.Getenv("MYSQL_PASSWORD"), "@tcp(", os.Getenv("MYSQL_HOST"), ":", os.Getenv("MYSQL_PORT"), ")/wall"}

	db, err := sql.Open("mysql", strings.Join(s, ""))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	shared_db = db

	router := mux.NewRouter()
	router.HandleFunc("/hello", GetHello).Methods("GET")
	router.HandleFunc("/bricks", GetBricks).Methods("GET")
	router.HandleFunc("/bricks/{id:[0-9_,-]+}", PutBrick).Methods("PUT")
	router.HandleFunc("/bricks/{id:[0-9_,-]+}", GetBrickImage).Methods("GET")
	router.HandleFunc("/bricks/{id:[0-9_,-]+}_t.jpg", GetBrickThumbnail).Methods("GET")

	log.Fatal(http.ListenAndServe(":8000", handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedMethods([]string{"POST", "PUT"}),
		handlers.AllowedHeaders([]string{"Content-Type", "X-Requested-With"}),
	)(router)))
}

func GetBricks(w http.ResponseWriter, req *http.Request) {
	var bricks = get_bricks(shared_db)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bricks)
}

func PutBrick(w http.ResponseWriter, req *http.Request) {
	var imageDataObject ImageDataPackage
	body, err := ioutil.ReadAll(io.LimitReader(req.Body, 1048576))
	if err != nil {
		log.Fatal(err)
	}
	if err := req.Body.Close(); err != nil {
		log.Fatal(err)
	}
	if err := json.Unmarshal(body, &imageDataObject); err != nil {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(422) // unprocessable entity
		if err := json.NewEncoder(w).Encode(err); err != nil {
			log.Fatal(err)
		}
	}

	i := strings.Index(imageDataObject.ImageData, ",")
	if i < 0 {
		log.Fatal("no comma")
	}
	imageData, err := base64.StdEncoding.DecodeString(imageDataObject.ImageData[i+1:])

	vars := mux.Vars(req)

	imagePath_Full := fmt.Sprintf("/bricks/%s", vars["id"])
	imagePath_Thumbnail := fmt.Sprintf("/bricks/%s_t.jpg", vars["id"])
	f, err := os.Create("/bricks/" + vars["id"])
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	n2, err := f.Write(imageData)
	log.Println(fmt.Sprintf("wrote %d bytes to %s\n", n2, "/bricks/"+vars["id"]))

	imageReader, err := os.Open(imagePath_Full)
	if err != nil {
		log.Fatal(err)
	}

	full_image, err := png.Decode(imageReader)
	if err != nil {
		log.Fatalf("failed to open image: %v", err)
	}

	thumbnail_image := image.NewRGBA(image.Rect(0, 0, 240, 135))
	draw.BiLinear.Scale(thumbnail_image, thumbnail_image.Bounds(), full_image, full_image.Bounds(), draw.Over, nil)

	tw, err := os.Create(imagePath_Thumbnail)
	if err != nil {
		log.Fatal(err)
	}
	defer tw.Close()
	err = png.Encode(tw, thumbnail_image)
	if err != nil {
		log.Fatal(err)
	}

	var brick = Brick{ID: vars["id"]}
	brick.CreationDate = fmt.Sprintf("%d", time.Now().UTC())
	brick.ETag = fmt.Sprintf("%d", time.Now().UTC().Unix())
	brick.ImageStoragePath = imagePath_Full
	brick.ThumbnailStoragePath = imagePath_Thumbnail

	defer req.Body.Close()
	upsert_brick(shared_db, brick)
	w.WriteHeader(http.StatusOK)

	result := PublishBrickResponse{Result: "OK"}
	json.NewEncoder(w).Encode(result)

}

func GetHello(w http.ResponseWriter, r *http.Request) {
	var bricks = get_bricks(shared_db)
	json.NewEncoder(w).Encode(bricks)
}

func GetBrickImage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/png")
	vars := mux.Vars(r)
	var brick = get_brick(shared_db, vars["id"])

	imagedata, err := ioutil.ReadFile(brick.ImageStoragePath)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	log.Println(brick.ImageStoragePath)
	w.Write(imagedata)
}

func GetBrickThumbnail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/png")
	vars := mux.Vars(r)
	var brick = get_brick(shared_db, vars["id"])

	imagedata, err := ioutil.ReadFile(brick.ThumbnailStoragePath)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	log.Println(brick.ImageStoragePath)
	w.Write(imagedata)
}
