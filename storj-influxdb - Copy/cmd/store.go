package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

// storeCmd represents the store command.
var storeCmd = &cobra.Command{
	Use:   "store",
	Short: "Command to upload data to storjV3 network.",
	Long:  `Command to connect and transfer ALL tables from a desired InfluxDB instance to given Storj Bucket.`,
	Run:   nextcloudStore,
}

func init() {

	// Setup the store command with its flags.
	rootCmd.AddCommand(storeCmd)
	var defaultNextcloudFile string
	var defaultStorjFile string
	var filePath string
	storeCmd.Flags().BoolP("accesskey", "a", false, "Connect to storj using access key(default connection method is by using API Key).")
	storeCmd.Flags().BoolP("share", "s", false, "For generating share access of the uploaded backup file.")
	storeCmd.Flags().BoolP("debug", "d", false, "For debugging purpose only.")
	storeCmd.Flags().StringVarP(&defaultNextcloudFile, "nextcloud", "i", "././config/nextcloud_property_tab.json", "full filepath contaning Influxdb configuration.")
	storeCmd.Flags().StringVarP(&defaultStorjFile, "storj", "u", "././config/storj_config_v01.json", "full filepath contaning storj V3 configuration.")
	storeCmd.Flags().StringVarP(&filePath, "path", "f", "/", "for listing the files present on the path.")
}

func nextcloudStore(cmd *cobra.Command, args []string) {

	// Process arguments from the CLI.
	nextcloudConfigfilePath, _ := cmd.Flags().GetString("nextcloud")
	fullFileNameStorj, _ := cmd.Flags().GetString("storj")
	useAccessKey, _ := cmd.Flags().GetBool("accesskey")
	useAccessShare, _ := cmd.Flags().GetBool("share")
	useDebug, _ := cmd.Flags().GetBool("debug")
	filePath, _ := cmd.Flags().GetString("path")

	// Read InfluxDB instance's configurations from an external file and create an InfluxDB configuration object.
	//configNextcloud, err := LoadNextcloudProperty(influxConfigfilePath)

	// Connect to Nextcloud using the specified credientials
	nextcloudClient, err := ConnectToNextcloud(nextcloudConfigfilePath)
	if err != nil {
		log.Fatalf("Failed to establish connection with Nextcloud: %s\n", err)

	}
	// Read storj network configurations from and external file and create a storj configuration object.
	storjConfig := LoadStorjConfiguration(fullFileNameStorj)

	// Connect to storj network using the specified credentials.
	access, project := ConnectToStorj(fullFileNameStorj, storjConfig, useAccessKey)

	// Create back-up of InfluxDB database and get file names to be uploaded.
	err = GetFilesWithPaths(nextcloudClient, filePath)
	//CreateBackup(configInfluxDB)

	fmt.Printf("\nInitiating back-up.\n")
	// Fetch all backup files from InfluxDB instance and simultaneously store them into desired Storj bucket.
	for i := 0; i < len(AllFilesWithPaths); i++ {
		file := AllFilesWithPaths[i]
		nextcloudReader := GetReader(nextcloudClient, file)
		//uploadFileName := path.Join(AllFilesWithPaths[i], file)
		UploadData(project, storjConfig, file, nextcloudReader, AllFilesWithPaths[i])

	}
	fmt.Printf("\nBack-up complete.\n\n")

	// Download the uploaded data if debug is provided as argument.
	if useDebug {
		fmt.Printf("Initiating download.\n\n")
		for i := 0; i < len(AllFilesWithPaths); i++ {
			file := AllFilesWithPaths[i]
			//downloadFileName := path.Join(AllFilesWithPaths[i])
			DownloadData(project, storjConfig, file)
		}
		fmt.Printf("\nDownload completed and stored inside debug folder.\n")
	}

	// Create restricted shareable serialized access if share is provided as argument.
	if useAccessShare {
		ShareAccess(access, storjConfig)
	}
}
