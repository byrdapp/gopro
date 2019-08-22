package scripts

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"

	storage "github.com/blixenkrone/gopro/storage"
	firebase "github.com/blixenkrone/gopro/storage/firebase"
)

func createCSV(record []string, index int, info interface{}) {
	csvWriter := CreateCSVWriterFile("media_credit_usage_dev")
	_ = csvWriter
	for _, val := range record {
		record = append(record, val)
	}
}

// WithdrawalsToCSV asdasd
func WithdrawalsToCSV(db *firebase.Firebase) {
	withdrawals, err := db.GetWithdrawals()
	if err != nil {
		log.Fatalf("Error initializing db %s\n", err)
	}

	csvWriter := CreateCSVWriterFile("withdrawals_not_completed_" + os.Getenv("ENV") + ".csv")
	for idx, wd := range withdrawals {
		writeWithdrawalsToCSV(db, csvWriter, idx, wd)
	}
	defer fmt.Println("Done writing CSV!")
}

// WriteWithdrawalsToCSV Does everything inside the loop above
func writeWithdrawalsToCSV(db *firebase.Firebase, w *csv.Writer, index int, val *storage.Withdrawals) {
	fmt.Println("The userID: ", val.RequestUserID)
	profile, err := db.GetProfile(context.Background(), val.RequestUserID)
	var record []string
	record = append(record, strconv.FormatInt(int64(index), 10))
	record = append(record, ParseUnixAsDate(val.RequestDate))
	record = append(record, profile.DisplayName)
	record = append(record, strconv.FormatInt(val.RequestAmount, 10))
	record = append(record, ParseUnixAsDate(val.RequestCompletedDate))
	record = append(record, "Udbetalt? : "+strconv.FormatBool(val.RequestCompleted))
	err = w.Write(record)
	if err != nil {
		log.Printf("Error writing result: %s", err)
	}
	fmt.Printf("\n----\nConverting key: %v with profileName: %v\n", index, profile.DisplayName)
	w.Flush()
}

// ProfilesToCSV initiates data and loops the write process
func ProfilesToCSV(db *firebase.Firebase) {
	prfs, err := db.GetProfiles(db.Context)
	if err != nil {
		log.Fatalf("Error initializing db %s\n", err)
	}

	csvWriter := CreateCSVWriterFile("profile_withdrawable_" + os.Getenv("ENV") + ".csv")
	rowNames := []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}
	WriteColumnHeaders(rowNames, csvWriter)

	index := int64(1)
	for _, prf := range prfs {
		if prf.SalesAmount > 0 {
			writeProfilesToCSV(csvWriter, index, prf)
			index++
		}
	}
	defer fmt.Println("Done writing CSV!")
}

// WriteProfilesToCSV Does everything inside the loop above
func writeProfilesToCSV(csvWriter *csv.Writer, index int64, profile *storage.FirebaseProfile) {
	alreadyCashedAmount := (profile.SalesAmount - profile.WithdrawableAmount)
	var record []string
	record = append(record, strconv.FormatInt(index, 10))
	record = append(record, profile.DisplayName)
	record = append(record, profile.FirstName+" "+profile.LastName)
	record = append(record, profile.Email)
	record = append(record, strconv.FormatBool(profile.IsProfessional))
	// record = append(record, ParseFloatUnixDate(profile.CreateDate))
	record = append(record, strconv.FormatInt(profile.SalesQuantity, 10))
	record = append(record, strconv.FormatInt(profile.WithdrawableAmount, 10))
	record = append(record, strconv.FormatInt(alreadyCashedAmount, 10))
	record = append(record, strconv.FormatInt(profile.SalesAmount, 10))
	err := csvWriter.Write(record)
	if err != nil {
		log.Printf("Error writing result: %s", err)
	}
	fmt.Printf("\n----\nConverting key: %v with profileName: %v\n", index, profile.DisplayName)
	csvWriter.Flush()
}
