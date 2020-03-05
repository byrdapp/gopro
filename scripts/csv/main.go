package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"

	storage "github.com/blixenkrone/byrd/byrd-pro-api/internal/storage"
	firebase "github.com/blixenkrone/byrd/byrd-pro-api/internal/storage/firebase"
	"github.com/blixenkrone/byrd/byrd-pro-api/pkg/logger"
)

// func createCSV(record []string, index int, info interface{}) {
// 	csvWriter := CreateCSVWriterFile("media_credit_usage_dev")
// 	_ = csvWriter
// 	for _, val := range record {
// 		record = append(record, val)
// 	}
// }

var (
	log = logger.NewLogger()
	fb  storage.FBService
)

func main() {
	if err := godotenv.Load(); err != nil {
		panic(err)
	}

	fbsrv, err := firebase.NewFB()
	if err != nil {
		log.Fatalf("Error starting firebase: %s", err)
	}
	fb = fbsrv
	// WithdrawalsToCSV()
	ProfilesToCSV()
}

// WithdrawalsToCSV .
func WithdrawalsToCSV() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*40)
	defer cancel()
	withdrawals, err := fb.GetWithdrawals(ctx)
	if err != nil {
		log.Fatalf("error getting withdrawals %s\n", err)
	}

	csvWriter := CreateCSVWriterFile("withdrawals_" + os.Getenv("ENV") + ".csv")
	for idx, wd := range withdrawals {
		writeWithdrawalsToCSV(csvWriter, idx, wd)
	}
	defer fmt.Println("Done writing CSV!")
}

// WriteWithdrawalsToCSV Does everything inside the loop above
func writeWithdrawalsToCSV(w *csv.Writer, index int, val *storage.Withdrawals) {
	fmt.Println("The userID: ", val.RequestUserID)
	profile, err := fb.GetProfile(context.Background(), val.RequestUserID)
	if err != nil {
		log.Fatal(err)
	}
	var record []string
	record = append(record, strconv.FormatInt(int64(index), 10))
	record = append(record, ParseUnixAsDate(val.RequestDate))
	record = append(record, profile.DisplayName)
	record = append(record, strconv.FormatInt(int64(val.RequestAmount), 10))
	record = append(record, ParseUnixAsDate(int64(val.RequestCompletedDate)))
	record = append(record, "Udbetalt? : "+strconv.FormatBool(val.RequestCompleted))
	err = w.Write(record)
	if err != nil {
		log.Printf("Error writing result: %s", err)
	}
	fmt.Printf("\n----\nConverting key: %v with profileName: %v\n", index, profile.DisplayName)
	w.Flush()
}

// ProfilesToCSV initiates data and loops the write process
func ProfilesToCSV() {
	prfs, err := fb.GetProfiles(context.Background())
	if err != nil {
		log.Fatalf("Error getting profiles from db %s\n", err)
	}

	csvWriter := CreateCSVWriterFile("profile_withdrawable_" + os.Getenv("ENV") + ".csv")
	rowNames := []string{"#", "dName", "fullName", "email", "pro user", "# of sales", "withdrawable amount", "already withdrawed", "total revenue", ""}
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

// writeProfilesToCSV Does everything inside the loop above
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

// CreateCSVWriterFile asd
func CreateCSVWriterFile(fileName string) *csv.Writer {
	file, err := os.Create(fileName)
	if err != nil {
		log.Println(err)
	}
	w := csv.NewWriter(file)
	return w
}

// WriteColumnHeaders sizes of columns
func WriteColumnHeaders(columns []string, csvWriter *csv.Writer) {
	err := csvWriter.Write(columns)
	if err != nil {
		log.Fatalf("Error writing columns %s", err)
	}
}

// ParseUnixAsDate asd
func ParseUnixAsDate(val int64) string {
	if val == 0 {
		return "%"
	}
	date := time.Unix((val / 1000), 0)
	fmt.Println(date)
	return date.UTC().String()
}

// ParseFloatUnixDate float64 instead of int64
func ParseFloatUnixDate(t *time.Time) string {
	return t.UTC().String()
}
