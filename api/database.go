package main

import (
	"database/sql"
	"log"
	"os"
	"strings"
	"time"
)

var RootConnectionString = strings.Join([]string{os.Getenv("MYSQL_USERNAME"), ":", os.Getenv("MYSQL_PASSWORD"), "@tcp(", os.Getenv("MYSQL_HOST"), ":", os.Getenv("MYSQL_PORT"), ")/"}, "")
var DbConnectionString = strings.Join([]string{os.Getenv("MYSQL_USERNAME"), ":", os.Getenv("MYSQL_PASSWORD"), "@tcp(", os.Getenv("MYSQL_HOST"), ":", os.Getenv("MYSQL_PORT"), ")/wall"}, "")

var revisions = []interface{}{Revision1}

func ProvisionDatabase() error {
	rootConn, err := sql.Open("mysql", RootConnectionString)
	if err != nil {
		log.Panic(err)
	}
	defer rootConn.Close()

	Initialize(rootConn)

	conn, err := sql.Open("mysql", DbConnectionString)
	if err != nil {
		log.Panic(err)
	}
	defer conn.Close()

	revision := GetRevision(conn)
	log.Printf("Database is at revision %d\n", revision)

	revision++
	for revision <= len(revisions) {
		log.Printf("Applying revision %d\n", revision)
		revisions[revision-1].(func(*sql.DB))(conn)
		revision++
	}

	return nil
}

func Initialize(conn *sql.DB) error {
	log.Println("Initializing database..")
	attempts := 0
	var err error
	for attempts < 10 {
		err = mysql_prepare_exec(conn, `CREATE DATABASE IF NOT EXISTS wall`)
		if err == nil {
			break
		}
		log.Println("Connection to database failed")
		attempts++
		time.Sleep(10 * time.Second)
	}
	if err != nil {
		log.Panic(err)
	}

	err = mysql_exec(conn, `USE wall`)
	if err != nil {
		log.Panic(err)
	}

	err = mysql_prepare_exec(conn, `CREATE TABLE IF NOT EXISTS `+"`bricks`"+` (
    `+"`ID`"+` VARCHAR(16) NOT NULL,
		`+"`ImageStoragePath`"+` VARCHAR(1024),
		`+"`ThumbnailStoragePath`"+` VARCHAR(1024),
    `+"`ETag`"+` VARCHAR(1024),
    `+"`CreationDate`"+` DATETIME,
    PRIMARY KEY (`+"`ID`"+`)
	);`)
	if err != nil {
		log.Panic(err)
	}

	return nil
}

func GetRevision(conn *sql.DB) int {
	log.Println("Getting current database revision..")
	rows, err := conn.Query(`SELECT 1 FROM db_revisions LIMIT 1;`)
	if err != nil {
		return 0
	}
	defer rows.Close()

	var revision int
	rows2, err := conn.Query(`select max(revision) from db_revisions where AppliedSuccessfully=true`)
	if err != nil {
		log.Panic(err)
	}
	defer rows2.Close()

	if rows2 != nil && rows2.Next() {
		err = rows2.Scan(&revision)
		if err != nil {
			log.Panic(err)
		}
		return revision
	}

	return 0
}

func mysql_prepare_exec(db *sql.DB, command string) error {
	stmtOut, err := db.Prepare(command)
	if err != nil {
		log.Println(err.Error())
		return err
	}
	defer stmtOut.Close()

	stmtOut.Exec()
	return nil
}

func mysql_exec(db *sql.DB, command string) error {
	_, err := db.Exec(command)
	if err != nil {
		log.Panic(err)
	}
	return nil
}

func Revision1(conn *sql.DB) {
	log.Println("Applying database revision 1..")
	st1, err := conn.Prepare(`CREATE TABLE IF NOT EXISTS ` + "`db_revisions`" + ` (
    ` + "`revision`" + ` SMALLINT NOT NULL,
		` + "`DateApplied`" + ` DATETIME,
		` + "`AppliedSuccessfully`" + ` BOOL,
    PRIMARY KEY (` + "`revision`" + `)
	);`)
	if err != nil {
		log.Panic(err)
	}
	defer st1.Close()
	_, err = st1.Exec()
	if err != nil {
		log.Panic(err)
	}

	revisionRow, err := conn.Query(`select * from db_revisions where revision = 1`)
	if !revisionRow.Next() {
		st2, err := conn.Prepare(`INSERT INTO db_revisions VALUES (?,?,?)`)
		if err != nil {
			log.Panic(err)
		}
		defer st2.Close()
		_, err = st2.Exec(1, time.Now().UTC().String(), false)
		if err != nil {
			log.Panic(err)
		}
	}

	st3, err := conn.Prepare(`ALTER TABLE bricks
ADD COLUMN TreatedImageStoragePath VARCHAR(1024) AFTER ImageStoragePath`)
	if err != nil {
		log.Panic(err)
	}
	defer st3.Close()
	_, err = st3.Exec()
	if err != nil {
		log.Panic(err)
	}

	st4, err := conn.Prepare(`UPDATE db_revisions set AppliedSuccessfully=? WHERE revision=?`)
	if err != nil {
		log.Panic(err)
	}
	defer st4.Close()
	_, err = st4.Exec(true, 1)
	if err != nil {
		log.Panic(err)
	}
}
