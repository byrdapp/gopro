package scripts

import (
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/byblix/gopro/storage"
	firebase "github.com/byblix/gopro/storage/firebase"
)

const noSSLstring = "http://res.cloudinary.com"
const hasSSLstring = "https://res.cloudinary.com"

// ChangeProfileUserPicture selfexplanatory
func ChangeProfileUserPicture() error {
	defer fmt.Println("Done")
	db, err := firebase.InitFirebaseDB()
	if err != nil {
		panic(err)
	}
	prfs, err := db.GetProfiles()
	if err != nil {
		return err
	}
	fmt.Println("Checking profiles")
	for _, prf := range prfs {
		str, change := changeDetection(prf)
		if change && str != "" {
			fmt.Printf("%s picture will be changed to: %s\n", prf.DisplayName, str)
			if err := db.UpdateData(prf.UserID, "userPicture", str); err != nil {
				log.Fatalf("Error with saving data %s", err)
			}
		}
	}

	return nil
}

func changeDetection(p *storage.Profile) (string, bool) {
	if contains := strings.Contains(p.UserPicture, noSSLstring); contains == true {
		fmt.Printf("%s is without SSL\n", p.DisplayName)
		corr := correctImageString(p)
		return corr, contains
	}
	return "", false
}

func correctImageString(p *storage.Profile) string {
	breakpoint := len(noSSLstring)
	sliced := p.UserPicture[breakpoint:]
	connected := hasSSLstring + sliced
	return connected
}

// DeleteUnusedAuthProfiles from Auth in Firebase
func DeleteUnusedAuthProfiles() error {
	var wg sync.WaitGroup
	db, err := firebase.InitFirebaseDB()
	profiles, err := db.GetAuth()
	if err != nil {
		return err
	}

	for _, p := range profiles {
		if p.UID != "wYVQ5ebjgyXVV1mecLuJi3nDAAD2" && p.UID != "0v6MD4T2PmVu237HkJTauoOOGIt1" {
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := db.DeleteAuthUserByUID(p.UID)
				if err != nil {
					log.Panicf("Error deleting: %s", err)
				}
				fmt.Printf("Deleted user: %s\n", p.UID)
			}()
			wg.Wait()
		}
	}

	return nil
}
