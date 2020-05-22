package cmd

import (
	//Standard Packages
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	//External Packages
	"gitlab.bertha.cloud/partitio/Nextcloud-Partitio/gonextcloud"
)

// ConfigNextcloud defines the variables and types for login.
type ConfigNextcloud struct {
	URL      string `json:"url"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// //Reader to read nextcloud file in chunks
// type Reader struct {
// 	read string
// 	done bool
// }

// LoadNextcloudProperty reads and parses the JSON file
// that contains a Nextcloud instance's properties
// and returns all the properties as an object.
func LoadNextcloudProperty(fullFileName string) (ConfigNextcloud, error) { // fullFileName for fetching database credentials from  given JSON filename.
	var configNextcloud ConfigNextcloud
	// Open and read the file
	fileHandle, err := os.Open(fullFileName)
	if err != nil {
		return configNextcloud, err
	}
	defer fileHandle.Close()

	jsonParser := json.NewDecoder(fileHandle)
	jsonParser.Decode(&configNextcloud)
	// Display Information about Nextcloud Instance.
	fmt.Println("Read Nextcloud configuration from the ", fullFileName, " file")
	fmt.Println("URL\t", configNextcloud.URL)
	fmt.Println("Username \t", configNextcloud.Username)
	fmt.Println("Password \t", configNextcloud.Password)
	return configNextcloud, nil
}

// //Read function reads the file through buffer
// func (r *Reader) Read(buff []byte) (n int, err error) {
// 	if r.done {
// 		return 0, io.EOF
// 	}
// 	for i, b := range []byte(r.read) {
// 		buff[i] = b
// 	}
// 	r.done = true
// 	return len(r.read), nil
// }

//ConnectToNextcloud : Connects to Nextcloud by creating a new authenticated Nextcloud client.
// It returns a Nextcloud client instance to perform file transfer(back-up)
func ConnectToNextcloud(fullFileName string) (*gonextcloud.Client, error) {
	configNextcloud, err := LoadNextcloudProperty(fullFileName)
	if err != nil {
		log.Printf("Loading NextcloudProperty: %s\n", err)
		return nil, err
	}
	fmt.Println("Connecting to Nextcloud...")
	nextcloudClient, err := gonextcloud.NewClient(configNextcloud.URL)
	if err != nil {
		fmt.Println("Client creation error:", err)
	}
	if err = nextcloudClient.Login(configNextcloud.Username, configNextcloud.Password); err != nil {
		fmt.Println("Login Error", err)
	}
	defer nextcloudClient.Logout()
	return nextcloudClient, err // return the NextcloudClient created to perform the download and store actions
}

//ListDirectory : Prints List of the directories and files in the nextcloud
func ListDirectory(nextcloudClient *gonextcloud.Client, path string) {
	folders, err := nextcloudClient.WebDav().ReadDir(path) //  path = "//"- for Root Directory listing and " Folder name" - folder listing
	if err != nil {
		log.Println("ReadDir error: ", err)
	}
	for _, file := range folders {
		if file.IsDir() { //If Folder, get all files.
			fmt.Println("-----------------------------")
			fmt.Println("Dir : ", " "+path+file.Name())
			ListDirectory(nextcloudClient, path+file.Name()+"/")

		} else { //For File
			fmt.Println("File : ", " "+path+file.Name())
		}
	}
}

// AllFilesWithPaths is a list to store complete
// path of all the files in the Nextcloud
// It will be used for direct transfer of files from Nextcloud to Storj
var AllFilesWithPaths []string

// GetFilesWithPaths retrieve the files' names with the exact
// file structure from Nextcloud Server to the System
func GetFilesWithPaths(nextcloudClient *gonextcloud.Client, path string) error {
	// If path given is of a folder
	if path[len(path)-1] == '/' {
		folders, err := nextcloudClient.WebDav().ReadDir(path)
		if err != nil {
			fmt.Println("GetFilesWithPaths Error: ", err)
			return err
		}
		for _, file := range folders {
			if file.IsDir() { //For folder, get all files in it
				GetFilesWithPaths(nextcloudClient, path+file.Name()+"/")
			} else {
				AllFilesWithPaths = append(AllFilesWithPaths, path+file.Name())
			}
		}
	} else { //if path given is of a file at root level
		AllFilesWithPaths = append(AllFilesWithPaths, path)
	}
	return nil
}

// GetReader returns a Reader of corresponding file whose path is specified by the user.
//  io.ReadCloser type of object returned is used to perform transfer of file to Storj
func GetReader(nextcloudClient *gonextcloud.Client, path string) io.ReadCloser {
	nextcloudReader, err := nextcloudClient.WebDav().ReadStream(path)
	if err != nil {
		fmt.Println("Read file error: ", err)
	}
	return nextcloudReader
}
