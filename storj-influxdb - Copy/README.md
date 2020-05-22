# storj-influxdb

# Developed using uplink RC v1.0.5

To build from scratch, [install Go](https://golang.org/doc/install#install).

## Download Project
* To download the project inside GOPATH use the following command:
```
go get github.com/utropicmedia/storj-influxdb
```
**NOTE**: To know your GOPATH, run ```go env gopath``` command. Navigate to the downloaded folder inside GOPATH and build the project.

* To download the project at a desired location, navigate to the desired location and run the following command:
```
git clone https://github.com/utropicmedia/storj-influxdb.git
```

## Set-up Files
* Create a `db_property.json` file, with following contents about a IndluxDB instance:
	* hostName :- Host Name connect to InfluxDB
	* port :- Port connect to InfluxDB
	* username :- User Name of InfluxDB
	* password :- Password of InfluxDB
	* database :- InfluxDB Database Name
	* influxdExeutablePath:- Path to the influxd executable including the executble name
```
{
	"hostname": "change-me-to-hostname",
	"port":     "change-me-to-port-number",
	"username": "change-me-to-username",
	"password": "change-me-to-password",
	"database": "change-me-to-database-name"
	"influxdExeutablePath": "change-me-to-path-of-influxd-executable-including-executable-name"
}
```

* Create a `storj_config.json` file, with Storj network's configuration information in JSON format:
	* apiKey:- API Key created in Storj Satellite GUI
	* satelliteURL:- Storj Satellite URL
	* encryptionPassphrase:- Storj Encryption Passphrase.
	* bucketName:- Name of the bucket to upload data into.
	* uploadPath:- Path on Storj Bucket to store data (optional) or "/"
	* serializedAccess:- Serialized access shared while uploading data used to access bucket without API Key
	* allowDownload:- Set true to create serialized access with restricted download
	* allowUpload:- Set true to create serialized access with restricted upload
	* allowList:- Set true to create serialized access with restricted list access
	* allowDelete:- Set true to create serialized access with restricted delete
	* notBefore:- Set time that is always before notAfter
	* notAfter:- Set time that is always after notBefore

```
{
	"apikey":     "change-me-to-the-api-key-created-in-satellite-gui",
	"satellite":  "us-central-1.tardigrade.io:7777",
	"bucket":     "change-me-to-desired-bucket-name",
	"uploadPath": "optionalpath/requiredfilename",
	"encryptionpassphrase": "you'll never guess this",
	"serializedAccess": "change-me-to-the-api-key-created-in-encryption-access-apiKey",
	"allowDownload":	"true/false-to-allow-download",
	"allowUpload":	"true/false-to-allow-upload",
	"allowList":	"true/false-to-list-buckets",
	"allowDelete":	"true/false-to-allow-delete",
	"notBefore":	"0/time-set-always-before-notAfter",
	"notAfter":	"0/time-set-always-after-notBefore"
}
```

* Store both these files in a `config` folder. Sample files are already provided. Make respective changes and you are good to go. File name command-line arguments are optional. Default locations are used.

## Description
* Overview
	Storj-InfluxDB Connector connects to the user InfluxDB database using the credentials provided by them in the database configuration file, takes backup of the specified database and uploads the backup files on Storj network inside the path specified by the user in the Storj configutation file.
* User can run the following command:
	* `store`:- This command works as follows:
		* Connect to the specidifed database using the user specified credential in the database configuration file (default: db_property).
		* Take backup of the database using a "backup" command provided by influxdb.
		* Connect to storj network using the serialized access specified in the database configuration file (default: storj_config).
		* Iterate through the backup files and upload each file in parts of buffer size 32KB to Storj network.
		Following are the flags that can be used with the `store` command:
			* `accesskey`:- Connects to Storj network using instead of Serialized Access Key instead of API key, satellite url and encryption passphrase .
			* `shared`:- Generates a restricted shareable serialized access with the restrictions specified in the storj configuration file.
			* `debug`:- Download the uploaded backup files to local disk inside project_folder/debug folder.

## Build Once
* To build the project, run the following command:
```
go build
```

**NOTE**: Please make sure influxd server is running in order to connect to database successfully.

## Run the command-line tool
* Once you have built the project run the following commands as per your requirement:

* Get help
```
$ ./storj-influxdb --help
```

* Check version
```
$ ./storj-influxdb --version
```

* Create and read backup files from InfluxDB instance and upload them to Storj bucket using API Key from `storj_config.json`.
```
$ ./storj-influxdb store --influx <path_to_influx_config_file> --storj <path_to_storj_config_file>
```

* Create and read backup files from InfluxDB instance and upload them to Storj bucket using Access Key from `storj_config.json`.
```
$ ./storj-influxdb store --accesskey
```

* Create and read backup files from InfluxDB instance and upload them to Storj bucket using API Key and generate a Shareable Access Key based on restrictions in `storj_config.json`.
```
$ ./storj-influxdb store --share
```

* Create and read backup files from InfluxDB instance and upload them to given Storj bucket using API Key from given Storj config file, further the uploaded file is downloaded from Storj to `./debug` folder.
```
$ ./storj-influxdb store --debug --influx <path_to_influx_config_file> --storj <path_to_storj_config_file>
```
**NOTE**: To restore database from the downloaded backup files after running `store` command with `--debug` flag, you can run the following command:
```
influxd restore -portable -db <old-database-name> -newdb <new-database-name> -host localhost:8088 <path_to_downloaded_backup_files>
```

##  Testing
* The project has been tested on the following operating systems:
```
	* Windows
		* Version: 10 Pro
		* Processor: Intel(R) Core(TM) i3-5005U CPU @ 2.00GHz 2.00GHz

	* macOS Catalina
		* Version: 10.15.4
		* Processor: 2.5 GHz Dual-Core Intel Core i5

	* ubuntu
		* Version: 16.04 LTS
		* Processor: AMD A6-7310 APU with AMD Radeon R4 Graphics × 4
```
