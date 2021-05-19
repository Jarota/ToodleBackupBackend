package scheduler

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"time"

	"github.com/jarota/ToodleBackupBackend/db"
	"github.com/jarota/ToodleBackupBackend/dropbox"
	"github.com/jarota/ToodleBackupBackend/toodledo"
	"github.com/jarota/ToodleBackupBackend/user"
	"go.mongodb.org/mongo-driver/bson"
)

const (
	taskFields = "folder, context, goal, location, tag, startdate, duedate, duedatemod, starttime, duetime, remind, repeat, status, star, priority, length, timer, added, note, parent, children, order, meta, previous, attachment, shared, addedby, via, attachments"
)

var client = db.ConnectToMongoDB()

// PollForPendingBackups continuously pings mongodb for users to backup
func PollForPendingBackups() {
	for {
		h, m, _ := time.Now().UTC().Clock()
		log.Printf("Polling database at %d:%d...", h, m)

		userCollection, err := db.GetCollection(client, "ToodleBackup", "Users")
		if err != nil {
			log.Fatal(err)
		}

		filter := bson.D{{Key: "time.hour", Value: h}, {Key: "time.minute", Value: m}}

		cursor, err := userCollection.Find(context.Background(), filter)
		if err != nil {
			log.Fatal("Error polling database for users to backup")
		}
		defer cursor.Close(context.Background())

		for cursor.Next(context.Background()) {
			var u user.User
			err := cursor.Decode(&u)
			if err != nil {
				log.Fatal("Error decoding user for backup")
			}

			if len(u.Clouds) > 0 && len(u.Toodledo.ToBackup) > 0 {
				go BackupUserData(&u)
			}
		}
		if err := cursor.Err(); err != nil {
			log.Fatal(err)
		}

		// Pause backing up for one minute
		time.Sleep(60 * time.Second)
	}

}

// BackupUserData backs up the user's data
func BackupUserData(user *user.User) {
	log.Printf("Backing up the user:  %s\n", user.Username)

	// First refresh toodledo access token
	toodleInfo, err := toodledo.GetToodledoTokens(user.Toodledo.Refresh, "refresh_token")
	// Update the user's toodleinfo in mongodb
	if err != nil {
		log.Fatal(err)
	}

	userCollection, err := db.GetCollection(client, "ToodleBackup", "Users")

	if err != nil {
		log.Fatal(err)
	}

	filter := bson.D{{Key: "username", Value: user.Username}}
	update := bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "toodledo", Value: *toodleInfo},
		}},
	}

	_, err = userCollection.UpdateOne(context.TODO(), filter, update)

	if err != nil {
		log.Fatal(err)
	}

	// Open a file with the current time as the name
	backupPath := user.Username + " " + time.Now().UTC().String()[:19] + ".xml"
	f, err := os.Create(backupPath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	now := time.Now().Unix()
	var offset int64 = 0
	n, err := f.WriteString(fmt.Sprintf("<xml>\n<title>Toodledo :: XML Backup</title>\n<link>http://www.toodledo.com/</link>\n<toodledoversion>20</toodledoversion>\n<description>Your Toodledo backup</description>\n<export_date>%v</export_date>\n", now))
	if err != nil {
		log.Fatal(err)
	}
	offset += int64(n)
	for _, s := range user.Toodledo.ToBackup {
		if s != "basic" {
			endpoint := "/3/" + s + "/get.php"
			data := append(retrieveFromToodledo(endpoint, toodleInfo.Token)[38:], []byte("\n")...) // Slice to skip <xml version> tag at beginning
			n, err := f.WriteAt(data, offset)
			if err != nil {
				log.Fatal(err)
			}
			offset += int64(n)
		}
	}
	_, err = f.WriteAt([]byte("</xml>"), offset)
	if err != nil {
		log.Fatal(err)
	}

	// Use the dropbox refresh token to retrieve an access token
	var dbxToken string
	for _, v := range user.Clouds {
		if v.Name == "Dropbox" {
			dbxToken = v.Token
		}
	}
	if len(dbxToken) != 0 {
		accessToken, _, err := dropbox.GetDropboxTokens(dbxToken, "refresh_token")
		if err != nil {
			log.Fatal(err)
		}

		// Call the dropbox python script with the backupPath and the access token
		cmd := exec.Command("python", "./backup.py", backupPath, accessToken)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			log.Fatal(err)
		}

	}

	err = os.Remove(backupPath)
	if err != nil {
		log.Fatal(err)
	}

}

func contains(l []string, val string) bool {
	for _, v := range l {
		if v == val {
			return true
		}
	}
	return false
}

func retrieveFromToodledo(endpoint string, token string) []byte {

	client := &http.Client{}

	apiURL := "https://api.toodledo.com"
	resource := endpoint
	u, _ := url.ParseRequestURI(apiURL)
	u.Path = resource
	urlStr := u.String()

	req, _ := http.NewRequest(http.MethodGet, urlStr, nil)
	params := req.URL.Query()
	if endpoint == "/3/tasks/get.php" {
		params.Add("fields", taskFields)
	}
	params.Add("access_token", token)
	params.Add("f", "xml")
	req.URL.RawQuery = params.Encode()

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	return bytes

}
