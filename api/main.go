package main

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
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
)

var shared_db *sql.DB

type Brick struct {
	ID               string `json:"id,omitempty"`
	ImageStoragePath string `json:"imageStoragePath,omitempty"`
	CreationDate     string `json: "creationDate,omitempty"`
	ETag             string `json: "etag,omitempty"`
}

type ImageDataPackage struct {
	ImageData string `json:"imagedata,omitempty"`
}

func get_brick(db *sql.DB, id string) Brick {
	var (
		imageStoragePath string
		eTag             string
		creationDate     string
		brick            Brick
	)

	stmtOut, err := db.Prepare("select ID, ImageStoragePath, ETag, CreationDate from bricks where ID = ?")
	if err != nil {
		panic(err.Error())
	}
	defer stmtOut.Close()

	log.Println(db)
	rows, err := stmtOut.Query(id)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&id, &imageStoragePath, &eTag, &creationDate)
		if err != nil {
			log.Fatal(err)
		}
		log.Println(id)
		brick = Brick{ID: id, ImageStoragePath: imageStoragePath, ETag: eTag, CreationDate: creationDate}
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	return brick
}

func get_bricks(db *sql.DB) []Brick {
	var (
		id               string
		imageStoragePath string
		eTag             string
		creationDate     string
		bricks           []Brick
	)

	log.Println(db)
	rows, err := db.Query("select ID, ImageStoragePath, ETag, CreationDate from bricks")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&id, &imageStoragePath, &eTag, &creationDate)
		if err != nil {
			log.Fatal(err)
		}
		log.Println(id)
		bricks = append(bricks, Brick{ID: id, ImageStoragePath: imageStoragePath, ETag: eTag, CreationDate: creationDate})
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
		stmtUpd, err := db.Prepare("UPDATE bricks set ImageStoragePath = ?, ETag = ?, CreationDate = ? where ID = ?")
		if err != nil {
			panic(err.Error()) // proper error handling instead of panic in your app
		}
		defer stmtUpd.Close() // Close the statement when we leave main() / the program terminates
		_, err = stmtUpd.Exec(brick.ImageStoragePath, brick.ETag, brick.CreationDate, brick.ID)
		if err != nil {
			panic(err.Error()) // proper error handling instead of panic in your app
		}
	} else {
		stmtIns, err := db.Prepare("INSERT INTO bricks VALUES( ?, ?, ?, ? )")
		if err != nil {
			panic(err.Error()) // proper error handling instead of panic in your app
		}
		defer stmtIns.Close() // Close the statement when we leave main() / the program terminates
		_, err = stmtIns.Exec(brick.ID, brick.ImageStoragePath, brick.ETag, brick.CreationDate)
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
    `+"`ETag`"+` VARCHAR(1024),
    `+"`CreationDate`"+` DATETIME,
    PRIMARY KEY (`+"`ID`"+`)
  );`)
	conn.Close()
}

func main() {
	create_database()
	s := []string{os.Getenv("MYSQL_USERNAME"), ":", os.Getenv("MYSQL_PASSWORD"), "@tcp(", os.Getenv("MYSQL_HOST"), ":", os.Getenv("MYSQL_PORT"), ")/wall"}
	log.Print(strings.Join(s, ""))
	db, err := sql.Open("mysql", strings.Join(s, ""))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	shared_db = db

	router := mux.NewRouter()
	router.HandleFunc("/hello", GetHello).Methods("GET")
	router.HandleFunc("/bricks", GetBricks).Methods("GET")
	router.HandleFunc("/bricks/{id:[0-9,-]+}", PutBrick).Methods("PUT")
	router.HandleFunc("/bricks/{id:[0-9,-]+}", GetBrickImage).Methods("GET")

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

	f, err := os.Create("/bricks/" + vars["id"])
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	n2, err := f.Write(imageData)
	log.Println(fmt.Sprintf("wrote %d bytes to %s\n", n2, "/bricks/"+vars["id"]))

	var brick = Brick{ID: vars["id"]}
	brick.CreationDate = fmt.Sprintf("%d", time.Now().UTC())
	brick.ETag = fmt.Sprintf("%d", time.Now().UTC().Unix())
	brick.ImageStoragePath = "/bricks/" + vars["id"]

	defer req.Body.Close()
	upsert_brick(shared_db, brick)
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
		log.Fatal(err)
	}

	log.Println(brick.ImageStoragePath)
	w.Write(imagedata)
}
