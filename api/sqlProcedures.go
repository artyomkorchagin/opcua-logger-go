package api

import (
	"database/sql"
	"log"
	"main/types"

	_ "github.com/microsoft/go-mssqldb"
)

var (
	query string

//		nodeAddress string
//	 tagid       string
)

func RequestNodeAdressesFromTag(tag types.Tag) types.DeviceLog {
	device := types.DeviceLog{}
	if tag.Enabled == 1 {
		device.Address = tag.Address
		device.Tagid = tag.ID
	}
	return device
}

func ConnectToDB() *sql.DB {

	dbconn, err := types.NewDBConnection()
	if err != nil {
		log.Fatal("Error in initializing config for connecting to DB: ", err)
	}

	dsn := "server=" + dbconn.Server + ";user id=" + dbconn.Userid + ";password=" + dbconn.Password + ";database=" + dbconn.Database
	db, err := sql.Open("mssql", dsn)
	if err != nil {
		log.Fatal("Error in connecting to DB")

	}
	return db
}

func UpdateTagsTable(db *sql.DB, t types.Tag) {
	query := "EXECUTE UpdateOrInsertTags @ID = ?, @NAME = ?, @DESC = ?, @ADDRESS = ?, @ENABLED = ?"
	_, err := db.Exec(query, t.ID, t.Name, t.Description, t.Address, t.Enabled)
	if err != nil {
		log.Fatal("failed updating Tags: ", err)
	}
}

// Может потом пригодится
// func RequestNodeAdressesFromDB(db *sql.DB) []types.DeviceLog {

// 	var ArrayOfAdresses []types.DeviceLog

// 	query = "EXECUTE GetTagAddresses"
// 	rows, err := db.Query(query)

// 	if err != nil {
// 		log.Fatal("failed fetching node adresses: ", err)
// 	}

// 	for rows.Next() {
// 		if err := rows.Scan(&tagid, &nodeAddress); err != nil {
// 			log.Fatal(err)
// 		}
// 		ArrayOfAdresses = append(ArrayOfAdresses, types.NewDeviceLog("", tagid, nodeAddress, "", "", ""))
// 	}

//		return ArrayOfAdresses
//	}

func InsertNewDataEntry(db *sql.DB, dl types.DeviceLog) {

	query = "EXECUTE RefreshLatestEntryAccumulation @TagID = ?"

	_, err := db.Exec(query, dl.Tagid)
	if err != nil {
		log.Fatal("failed refresh: ", err)
	}

	query = "EXECUTE InsertAccumulation @Timestamp = ?, @TagID = ?, @Value = ?, @Latest = ?, @Quality = ?"
	_, err = db.Exec(query, dl.Timestamp, dl.Tagid, dl.Value, "1", dl.Quality)
	if err != nil {
		log.Fatal("failed insert: ", err)
	}

}
