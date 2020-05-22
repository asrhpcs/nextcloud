package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"storj.io/uplink"
)

// DEBUG allows more detailed working to be exposed through the terminal.
var DEBUG = false

// MAXRETRY defines number of times to try upload data to storj before throwing error
var MAXRETRY = 5

// ConfigStorj depicts keys to search for within the stroj_config.json file.
type ConfigStorj struct {
	APIKey               string `json:"apikey"`
	Satellite            string `json:"satellite"`
	Bucket               string `json:"bucket"`
	UploadPath           string `json:"uploadPath"`
	EncryptionPassphrase string `json:"encryptionpassphrase"`
	SerializedAccess     string `json:"serializedAccess"`
	AllowDownload        string `json:"allowDownload"`
	AllowUpload          string `json:"allowUpload"`
	AllowList            string `json:"allowList"`
	AllowDelete          string `json:"allowDelete"`
	NotBefore            string `json:"notBefore"`
	NotAfter             string `json:"notAfter"`
}

// LoadStorjConfiguration reads and parses the JSON file that contain Storj configuration information.
func LoadStorjConfiguration(fullFileName string) ConfigStorj {

	var configStorj ConfigStorj
	fileHandle, err := os.Open(filepath.Clean(fullFileName))
	if err != nil {
		log.Fatal("Could not load storj config file: ", err)
	}

	jsonParser := json.NewDecoder(fileHandle)
	if err = jsonParser.Decode(&configStorj); err != nil {
		log.Fatal(err)
	}

	// Close the file handle after reading from it.
	if err = fileHandle.Close(); err != nil {
		log.Fatal(err)
	}

	// Display storj configuration read from file.
	fmt.Println("\nRead Storj configuration from the ", fullFileName, " file")
	fmt.Println("\nAPI Key\t\t: ", configStorj.APIKey)
	fmt.Println("Satellite	: ", configStorj.Satellite)
	fmt.Println("Bucket		: ", configStorj.Bucket)

	// Convert the upload path to standard form.
	checkSlash := configStorj.UploadPath[len(configStorj.UploadPath)-1:]
	if checkSlash != "/" {
		configStorj.UploadPath = configStorj.UploadPath + "/"
	}

	fmt.Println("Upload Path\t: ", configStorj.UploadPath)
	fmt.Println("Serialized Access Key\t: ", configStorj.SerializedAccess)
	return configStorj
}

// ShareAccess generates and prints the shareable serialized access
// as per the restrictions provided by the user.
func ShareAccess(access *uplink.Access, configStorj ConfigStorj) {

	allowDownload, _ := strconv.ParseBool(configStorj.AllowDownload)
	allowUpload, _ := strconv.ParseBool(configStorj.AllowUpload)
	allowList, _ := strconv.ParseBool(configStorj.AllowList)
	allowDelete, _ := strconv.ParseBool(configStorj.AllowDelete)
	notBefore, _ := time.Parse("2006-01-02_15:04:05", configStorj.NotBefore)
	notAfter, _ := time.Parse("2006-01-02_15:04:05", configStorj.NotAfter)

	permission := uplink.Permission{
		AllowDownload: allowDownload,
		AllowUpload:   allowUpload,
		AllowList:     allowList,
		AllowDelete:   allowDelete,
		NotBefore:     notBefore,
		NotAfter:      notAfter,
	}

	// Create shared access.
	sharedAccess, err := access.Share(permission)
	if err != nil {
		log.Fatal("Could not generate shared access: ", err)
	}

	// Generate restricted serialized access.
	serializedAccess, err := sharedAccess.Serialize()
	if err != nil {
		log.Fatal("Could not serialize shared access: ", err)
	}
	fmt.Println("Shareable sererialized access: ", serializedAccess)
}

// ConnectToStorj reads Storj configuration from given file
// and connects to the desired Storj network.
// It then reads data property from an external file.
func ConnectToStorj(fullFileName string, configStorj ConfigStorj, accesskey bool) (*uplink.Access, *uplink.Project) {

	var access *uplink.Access
	var cfg uplink.Config

	// Configure the UserAgent
	cfg.UserAgent = "InfluxDB"
	ctx := context.Background()
	var err error

	if accesskey {
		fmt.Println("\nConnecting to Storj network using Serialized access.")
		// Generate access handle using serialized access.
		access, err = uplink.ParseAccess(configStorj.SerializedAccess)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		fmt.Println("\nConnecting to Storj network.")
		// Generate access handle using API key, satellite url and encryption passphrase.
		access, err = cfg.RequestAccessWithPassphrase(ctx, configStorj.Satellite, configStorj.APIKey, configStorj.EncryptionPassphrase)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Open a new porject.
	project, err := cfg.OpenProject(ctx, access)
	if err != nil {
		log.Fatal(err)
	}
	defer project.Close()

	// Ensure the desired Bucket within the Project
	_, err = project.EnsureBucket(ctx, configStorj.Bucket)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Successfully connected to Storj network.")
	return access, project
}

// UploadData uploads the backup file to storj network.
func UploadData(project *uplink.Project, configStorj ConfigStorj, uploadFileName string, reader io.Reader, filePath string) {

	ctx := context.Background()

	// To trim the first "/" form the file name
	checkSlash := uploadFileName[0:1]
	if checkSlash == "/" {
		uploadFileName = trimLeftSlash(uploadFileName)
	}

	// Standardize the upload path
	checkSlash = configStorj.UploadPath[len(configStorj.UploadPath)-1:]
	if checkSlash != "/" {
		configStorj.UploadPath = configStorj.UploadPath + "/"
	}

	// Create an upload handle.
	upload, err := project.UploadObject(ctx, configStorj.Bucket, configStorj.UploadPath+uploadFileName, nil)
	if err != nil {
		log.Fatal("Could not initiate upload : ", err)
	}
	fmt.Printf("\nUploading %s to %s.", configStorj.UploadPath+uploadFileName, configStorj.Bucket)

	var lastIndex int64
	var numOfBytesRead int
	lastIndex = 0
	var buf = make([]byte, 32768)

	// fileReader, err := os.Open(filepath.Clean(filePath))
	// if err != nil {
	// 	log.Fatal(err)
	// }
	buff := bytes.NewBuffer([]byte{})
	_, err = io.Copy(buff, reader)
	if err != nil {
		log.Fatal(err)
	}

	ioReader := bytes.NewReader(buff.Bytes())

	var err1 error
	// Loop to read the backup file in chunks and append the contents to the upload object.
	for err1 != io.EOF {
		sectionReader := io.NewSectionReader(ioReader, lastIndex, int64(cap(buf)))
		numOfBytesRead, err1 = sectionReader.ReadAt(buf, 0)
		if numOfBytesRead > 0 {
			reader := bytes.NewBuffer(buf[0:numOfBytesRead])
			// Try to upload data on storj n number of times
			retry := 0
			for retry < MAXRETRY {
				_, err = io.Copy(upload, reader)
				if err != nil {
					retry++
				} else {
					break
				}
			}
			if retry == MAXRETRY {
				log.Fatal("Could not upload data to storj: ", err)
			}

		}

		lastIndex = lastIndex + int64(numOfBytesRead)
	}

	// Commit the upload after copying the complete content of the backup file to upload object.
	fmt.Println("\nPlease wait while the upload is being committed to Storj.")
	err = upload.Commit()
	if err != nil {
		log.Fatal("Could not commit object upload : ", err)
	}

	// // Close file handle after reading from it.
	// if err = ioReader.Close(); err != nil {
	// 	log.Fatal(err)
	// }

}

// DownloadData function downloads the data from storj bucket after upload to verify data is uploaded successfully.
func DownloadData(project *uplink.Project, configStorj ConfigStorj, downloadFileName string) {

	ctx := context.Background()

	// To trim the first "/" form the file name
	checkSlash := downloadFileName[0:1]
	if checkSlash == "/" {
		downloadFileName = trimLeftSlash(downloadFileName)
	}

	// Standardize the upload path
	checkSlash = configStorj.UploadPath[len(configStorj.UploadPath)-1:]
	if checkSlash != "/" {
		configStorj.UploadPath = configStorj.UploadPath + "/"
	}

	directory, file := filepath.Split(downloadFileName)
	fmt.Printf("Directory : " + directory)
	fmt.Printf("file : " + file)
	var receivedContents = []byte{}
	var lastIndex int64
	var buf = make([]byte, 32768)
	lastIndex = 0

	// To retrieve information(mainly size) of the uploaded object.
	object, err := project.StatObject(ctx, configStorj.Bucket, configStorj.UploadPath+downloadFileName)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Downloading %s.\n", file)

	// Loop to read the object in chunks and store the read data in a byte array.
	for int64(len(receivedContents)) < object.System.ContentLength {
		var download *uplink.Download
		var retry = 0
		// Try to download data from storj n number of times
		for retry < MAXRETRY {
			download, err = project.DownloadObject(ctx, configStorj.Bucket, configStorj.UploadPath+downloadFileName, &uplink.DownloadOptions{Offset: lastIndex, Length: int64(cap(buf))})
			if err != nil {
				retry++
			} else {
				break
			}
		}
		if retry == MAXRETRY {
			log.Fatal("Could not download data form storj: ", err)
		}

		data, err2 := ioutil.ReadAll(download)
		if err2 != nil {
			break
		}

		// Append the read bytes to a single slice to write into the local file at once.
		receivedContents = append(receivedContents, data...)
		lastIndex = lastIndex + int64(len(data))
	}
	fmt.Printf("Directory : " + directory)
	// Create the debug directory if not exists.
	if _, err := os.Stat("./debug"); os.IsNotExist(err) {
		err1 := os.Mkdir("./debug", 0750)
		if err1 != nil {
			log.Fatal("Invalid Download Path: ", err1)
		}
	}
	if _, err := os.Stat("./debug" + "/" + directory); os.IsNotExist(err) {
		err1 := os.Mkdir("./debug"+"/"+directory, 0750)
		if err1 != nil {
			log.Fatal("Invalid Download Path: ", err1)
		}
	}

	// Create/open file in append mode.
	downloadFileDisk, err := os.OpenFile(filepath.Join("./debug", directory, file), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0750)
	if err != nil {
		log.Fatal(err)
	}

	// Write the contents retrieved from downloaded object to file on local disk.
	_, err = downloadFileDisk.Write(receivedContents)
	if err != nil {
		log.Fatal(err)
	}

	// Close the file handle after reading from it.
	if err = downloadFileDisk.Close(); err != nil {
		log.Fatal(err)
	}

}

func trimLeftSlash(file string) string {
	for i := range file {
		if i > 0 {
			return file[i:]
		}
	}
	return file[:0]
}
