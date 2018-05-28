package main

import (
	"bytes"
	"database/sql"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

var shared_db *sql.DB

type ImageDataPackage struct {
	ImageData string `json:"imagedata,omitempty"`
}

type PublishBrickResponse struct {
	Result string `json:"result,omitempty"`
}

func init() {
	gob.Register(Brick{})
}

func main() {
	err := ProvisionDatabase()
	if err != nil {
		log.Panic(err)
	}
	log.Println("Database initialized.")

	db, err := sql.Open("mysql", DbConnectionString)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	shared_db = db

	router := mux.NewRouter()
	router.HandleFunc("/hello", ApiGetHello).Methods("GET")
	router.HandleFunc("/bricks", ApiGetBricks).Methods("GET")
	router.HandleFunc("/bricks/{id:[0-9_,-]+}", ApiPutBrick).Methods("PUT")
	router.HandleFunc("/bricks/{id:[0-9_,-]+}/original.png", ApiGetBrickImage).Methods("GET")
	router.HandleFunc("/bricks/{id:[0-9_,-]+}.png", ApiGetTreatedBrickImage).Methods("GET")
	router.HandleFunc("/bricks/{id:[0-9_,-]+}/thumbnail.png", ApiGetBrickThumbnail).Methods("GET")

	log.Fatal(http.ListenAndServe(":8000", handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedMethods([]string{"POST", "PUT"}),
		handlers.AllowedHeaders([]string{"Content-Type", "X-Requested-With"}),
	)(router)))

}

func ApiGetBricks(w http.ResponseWriter, req *http.Request) {
	var metadata = GetMetadata(shared_db)
	w.Header().Set("Content-Type", "application/json")
	if metadata.Valid {
		w.Header().Set("ETag", metadata.String)

		if req.Header.Get("If-None-Match") == metadata.String {
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}

	var bricks = GetBricks(shared_db)
	json.NewEncoder(w).Encode(bricks)
}

func ApiPutBrick(w http.ResponseWriter, req *http.Request) {
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
			log.Println(err.Error())
		}
		return
	}

	i := strings.Index(imageDataObject.ImageData, ",")
	if i < 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	imageData, err := base64.StdEncoding.DecodeString(imageDataObject.ImageData[i+1:])

	vars := mux.Vars(req)

	imagePath_Original_Full := fmt.Sprintf("/bricks/%s.png", vars["id"])
	imagePath_Treated_Full := fmt.Sprintf("/bricks/%s_treated.png", vars["id"])
	imagePath_Thumbnail := fmt.Sprintf("/bricks/%s_t.png", vars["id"])

	f, err := os.Create(imagePath_Original_Full)
	if err != nil {
		log.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer f.Close()
	_, err = f.Write(imageData)
	if err != nil {
		log.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	fullOriginalImage, err := imaging.Open(imagePath_Original_Full)
	if err != nil {
		log.Println("failed to open image: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	treatedImage := imaging.Convolve3x3(
		fullOriginalImage,
		[9]float64{
			-1.5, -1, 0,
			-1, 1, 1,
			0, 1, 1.5,
		},
		nil,
	)
	treatedImage = imaging.AdjustContrast(treatedImage, 20)
	treatedImage = imaging.Sharpen(treatedImage, 0.5)
	treatedImage = imaging.Blur(treatedImage, 0.3)
	err = imaging.Save(treatedImage, imagePath_Treated_Full)
	if err != nil {
		log.Panicf("failed to save image: %v", err)
	}

	thumbnail := imaging.Resize(treatedImage, 240, 0, imaging.Lanczos)
	err = imaging.Save(thumbnail, imagePath_Thumbnail)
	if err != nil {
		log.Fatalf("failed to save image: %v", err)
	}

	var brick = Brick{ID: vars["id"]}
	brick.CreationDate = fmt.Sprintf("%d", time.Now().UTC())
	brick.ETag = fmt.Sprintf("%d", time.Now().UTC().Unix())
	brick.ImageStoragePath = imagePath_Original_Full
	brick.TreatedImageStoragePath = sql.NullString{String: imagePath_Treated_Full, Valid: true}
	brick.ThumbnailStoragePath = imagePath_Thumbnail

	defer req.Body.Close()
	SaveBrick(shared_db, brick)
	w.WriteHeader(http.StatusOK)

	var bricks = GetBricks(shared_db)
	b := bytes.Buffer{}
	e := gob.NewEncoder(&b)
	err = e.Encode(bricks)
	if err != nil {
		fmt.Println(`failed gob Encode`, err)
	}
	etag := fmt.Sprintf(`W/%08X`, crc32.ChecksumIEEE(b.Bytes()))
	log.Printf("gob: %s\n", etag)

	err = UpdateMetadata(shared_db, etag)
	if err != nil {
		log.Printf("Failed to update metadata: %s\n", err.Error())
	}

	result := PublishBrickResponse{Result: "OK"}
	json.NewEncoder(w).Encode(result)

}

func ApiGetHello(w http.ResponseWriter, r *http.Request) {
	var bricks = GetBricks(shared_db)
	json.NewEncoder(w).Encode(bricks)
}

func ApiGetBrickImage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/png")
	vars := mux.Vars(r)
	brick, err := GetBrick(shared_db, vars["id"])

	imagedata, err := ioutil.ReadFile(brick.ImageStoragePath)
	if err != nil {
		log.Printf("%s: %s\n", brick.ImageStoragePath, err.Error())
		w.WriteHeader(http.StatusNotFound)
		return
	}

	log.Println(brick.ImageStoragePath)
	w.Write(imagedata)
}

func ApiGetTreatedBrickImage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/png")
	vars := mux.Vars(r)
	brick, err := GetBrick(shared_db, vars["id"])
	if !brick.TreatedImageStoragePath.Valid {
		log.Printf("Request for treated image for brick %s but brick doesn't have a treated image\n", vars["id"])
		w.WriteHeader(http.StatusNotFound)
		return
	}

	imagedata, err := ioutil.ReadFile(brick.TreatedImageStoragePath.String)
	if err != nil {
		log.Println(err.Error())
		w.WriteHeader(http.StatusNotFound)
		return
	}

	log.Println(brick.ImageStoragePath)
	w.Write(imagedata)
}

func ApiGetBrickThumbnail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/png")
	vars := mux.Vars(r)
	brick, err := GetBrick(shared_db, vars["id"])
	if err != nil {
		log.Printf("%s: %s\n", brick.ThumbnailStoragePath, err.Error())
		w.WriteHeader(http.StatusNotFound)
		return
	}

	imagedata, err := ioutil.ReadFile(brick.ThumbnailStoragePath)
	if err != nil {
		log.Printf("%s: %s\n", brick.ThumbnailStoragePath, err.Error())
		w.WriteHeader(http.StatusNotFound)
		return
	}

	log.Println(brick.ImageStoragePath)
	w.Write(imagedata)
}

func TreatImage(inputFilePath string, outputFilePath string) error {
	src, err := imaging.Open(inputFilePath)
	if err != nil {
		log.Panicf("failed to open %s: %v\n", inputFilePath, err)
	}
	img1 := imaging.Convolve3x3(
		src,
		[9]float64{
			-1.5, -1, 0,
			-1, 1, 1,
			0, 1, 1.5,
		},
		nil,
	)
	img1 = imaging.AdjustContrast(img1, 20)
	img1 = imaging.Sharpen(img1, 0.5)
	//	img1 = imaging.Blur(img1, 0.3)
	err = imaging.Save(img1, outputFilePath)
	if err != nil {
		log.Panicf("failed to save image: %v", err)
	}
	return nil
}
