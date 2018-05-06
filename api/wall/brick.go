package wall

import(
	"log"
	"database/sql"
)

type Brick struct {
	ID string `json:"id,omitempty"`
	ImageStoragePath string `json:"imageStoragePath,omitempty"`
	CreationDate string `json: "creationDate,omitempty"`
	ETag string `json: "etag,omitempty"`
  }
  
  func get_brick(db *sql.DB, id string) Brick {
	var (
	imageStoragePath string
	eTag string
	creationDate string
	  brick Brick
	)
  
	stmtOut, err := db.Prepare("select ID, ImageStoragePath, ETag, CreationDate from bricks where ID = ?")
	if err != nil {
	  panic(err.Error())
	}
	defer stmtOut.Close()
  
	log.Println(db)
	rows, err := stmtOut.Query(id);
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
	  brick = Brick{ID: id, ImageStoragePath: imageStoragePath, ETag: eTag, CreationDate: creationDate};
	}
	err = rows.Err()
	if err != nil {
	  log.Fatal(err)
	}
	return brick;
  }
  
  func get_bricks(db *sql.DB) []Brick {
	var (
	  id string
	  imageStoragePath string
	  eTag string
	  creationDate string
	  bricks []Brick
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
	  err:= rows.Scan(&count)
	  if err != nil {
		log.Fatal(err)
	  }
	}   
  
	if (count > 0) {
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