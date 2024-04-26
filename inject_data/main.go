package main

import (
	"log"
	"os"
	"path"
)

const DEFAULT_CACHEDIR_RELNAME = ".sdcc_dinj_cache/"
const DEFAULT_FILE_NAME = "nyc_yellowtaxis_feb2024.csv"
const DEFAULT_SKIP_CHECKSUM = false
const DEFAULT_SKIP_DOWNLOAD = false
const EXPECTED_FILE_SHA256_CHECKSUM = "D22A63D4EE390D4375F3EAC901FD5C4B5FDB938786E7E4D5294893B1B43B75E9"
const DEFAULT_URL = "https://drive.usercontent.google.com/download?id=1mqkh5NOnXcPbMaDtlQwbqAohh2hmwD9A&" +
	"export=download&authuser=0&confirm=t&uuid=e9f71f77-81ff-43a9-8a61-3b5c7e707aa1&" +
	"at=APZUnTVXgI56OrhPdA_2E6QDCYca%3A1714139639065"

type Config struct {
	filename     string
	cacheDirPath string
	checksum     string
	downloadUrl  string
	skipChecksum bool
	skipDownload bool
}

type Argument struct {
	name        string
	description string
	needsValue  bool
	handler     func(string)
}

var programConfig Config
var programArguments []Argument = []Argument{
	{
		name:        "--checksum",
		description: "Enable custom checksum verification",
		needsValue:  true,
		handler: func(val string) {
			programConfig.checksum = val
			programConfig.skipChecksum = false
			if len(programConfig.checksum) != 64 {
				log.Fatalf("SHA256 checksum strings " +
					"(base64 encoded) must be exactly 64 char long")
			}
		},
	},
	{
		name:        "--skip-checksum",
		description: "Disable checksum verification",
		needsValue:  false,
		handler: func(_ string) {
			programConfig.skipChecksum = true
		},
	},
	{
		name:        "--checksum-default",
		description: "Enable default checksum verification",
		needsValue:  false,
		handler: func(_ string) {
			programConfig.checksum = EXPECTED_FILE_SHA256_CHECKSUM
			programConfig.skipChecksum = false
		},
	},
	{
		name:        "--filename",
		description: "Set custom relative (to cachedir) filename to use",
		needsValue:  true,
		handler: func(val string) {
			programConfig.filename = path.Clean(val)
		},
	},
	{
		name:        "--filename-default",
		description: "Set default relative (to cachedir) filename to use",
		needsValue:  false,
		handler: func(val string) {
			programConfig.filename = DEFAULT_FILE_NAME
		},
	},
	{
		name:        "--cachedir",
		description: "Set custom cache directory",
		needsValue:  true,
		handler: func(val string) {
			programConfig.cacheDirPath = path.Clean(val)
		},
	},
	{
		name:        "--cachedir-default",
		description: "Set default cache directory",
		needsValue:  false,
		handler: func(val string) {
			programConfig.cacheDirPath = DEFAULT_CACHEDIR_RELNAME
		},
	},
	{
		name:        "--download",
		description: "Enable downloading required file by custom URL",
		needsValue:  true,
		handler: func(val string) {
			programConfig.downloadUrl = val
			programConfig.skipDownload = false
		},
	},
	{
		name:        "--download-default",
		description: "Enable downloading required file by default URL",
		needsValue:  false,
		handler: func(_ string) {
			programConfig.downloadUrl = DEFAULT_URL
			programConfig.skipDownload = false
		},
	},
	{
		name:        "--skip-download",
		description: "Disable downloading required file",
		needsValue:  false,
		handler: func(_ string) {
			programConfig.skipDownload = true
		},
	},
}

func getHomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("unable to get user home: %s\n", err.Error())
	}

	return home
}

func printHelpAndExit() {
	log.Printf(
		"default filename:\t%s\n"+
			"default cachedir:\t%s\n"+
			"default checksum value:\t%s\n"+
			"default download url:\t%s\n"+
			"default checksum validate:\t%t\n"+
			"default download file if missing:\t%t\n",
		programConfig.filename,
		programConfig.cacheDirPath,
		programConfig.checksum,
		programConfig.downloadUrl,
		!programConfig.skipChecksum,
		!programConfig.skipDownload)

	for _, arg := range programArguments {
		needsValueToStr := "no"
		if arg.needsValue {
			needsValueToStr = "yes"
		}

		log.Printf(" * %s:\tdesc: %s\tneedsValue: %s\n",
			arg.name, arg.description, needsValueToStr)
	}
	os.Exit(0)
}

func configureProgramByArgs() {
	dflCacheDir := getHomeDir() + "/" + DEFAULT_CACHEDIR_RELNAME

	programConfig.filename = DEFAULT_FILE_NAME
	programConfig.cacheDirPath = path.Clean(dflCacheDir)
	programConfig.checksum = EXPECTED_FILE_SHA256_CHECKSUM
	programConfig.downloadUrl = DEFAULT_URL
	programConfig.skipChecksum = DEFAULT_SKIP_CHECKSUM
	programConfig.skipDownload = DEFAULT_SKIP_DOWNLOAD

	args := os.Args

	for i := 0; i < len(args); i++ {
		if args[i] == "--help" {
			printHelpAndExit()
		}
	}

	for i := 0; i < len(args); i++ {
		matchingArgFound := false
		for _, arg := range programArguments {
			if args[i] == arg.name {
				matchingArgFound = true
				if arg.needsValue {
					i++
					if i == len(args) {
						log.Fatalf("argument %s requires value!\n", args[i-1])
					}
				}
				arg.handler(args[i])
				break
			}
		}
		if !matchingArgFound {
			log.Printf("ignoring unknown cmdline opt %s\n", args[i])
		}
	}
}

func main() {
	//--filename <filename> or --filename-default
	//--cachedir <cachedir> or --cachedir-default
	//--skip-checksum or --checksum-default or --checksum <custom_hexascii_sha256>
	//--skip-download or --download-default or --download <url>
}
