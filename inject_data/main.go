package main

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
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
		description: "Enable custom checksum value verification (SHA256)",
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
		description: "Enable default checksum value verification (SHA256)",
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
			programConfig.filename = val
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
			programConfig.cacheDirPath = val
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

func getFwdPathSep(path string) string {
	return strings.ReplaceAll(path, "\\", "/")
}

func getHomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("unable to get user home: %s\n", err.Error())
	}

	return home
}

func printHelpAndExit() {
	log.Printf("default filename: %s\n", programConfig.filename)
	log.Printf("default cachedir: %s\n", programConfig.cacheDirPath)
	log.Printf("default checksum value: %s\n", programConfig.checksum)
	log.Printf("default download url: %s\n", programConfig.downloadUrl)
	log.Printf("default checksum validate: %t\n", !programConfig.skipChecksum)
	log.Printf("default download file if missing: %t\n", !programConfig.skipDownload)

	for _, arg := range programArguments {
		log.Printf(" * %s - %s, value required: %t\n",
			arg.name, arg.description, arg.needsValue)
	}

	os.Exit(0)
}

func configureProgramByArgs() {
	dflCacheDir := getHomeDir() + "/" + DEFAULT_CACHEDIR_RELNAME

	programConfig.filename = DEFAULT_FILE_NAME
	programConfig.cacheDirPath = dflCacheDir
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

	for i := 1; i < len(args); i++ {
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

	uniformConfigParameters()
}

func uniformConfigParameters() {
	programConfig.checksum = strings.ToUpper(programConfig.checksum)
	programConfig.cacheDirPath, _ = filepath.Abs(programConfig.cacheDirPath)
	programConfig.filename = getFwdPathSep(programConfig.filename)
	programConfig.cacheDirPath = getFwdPathSep(programConfig.cacheDirPath)
}

func checkFile() (string, bool) {
	fpath := programConfig.cacheDirPath + "/" + programConfig.filename

	if filepath.IsAbs(programConfig.filename) {
		dir, _ := path.Split(programConfig.filename)
		if path.Clean(dir) != programConfig.cacheDirPath {
			log.Fatalln("basedir is different from cachedir")
		}

		fpath = programConfig.filename
	} else if strings.Contains(programConfig.filename, "/") {
		log.Println("if filename value is something like \"./file.txt\" or \".\\file.txt\"")
		log.Println("then try replacing it like \"file.txt\" (basic path checks implemented)")
		log.Fatalln("no subdirectories in cachedir allowed")
	}

	log.Printf("checking file %s...", fpath)

	_, err := os.Stat(fpath)

	return fpath, !os.IsNotExist(err)
}

func sha256Checksum(path string) string {
	fileBytes, err := os.ReadFile(path)
	if err != nil {
		return "failfailfail"
	}

	hasher := sha256.New()
	if _, err := hasher.Write(fileBytes); err != nil {
		return "failfailfail"
	}

	hexStr := hex.EncodeToString(hasher.Sum(nil))
	return strings.ToUpper(hexStr)
}

func fileDownload(path string) {
	res, err := http.Get(programConfig.downloadUrl)
	if err != nil {
		log.Fatalf("unable to perform HTTP GET: %s\n",
			err.Error())
	}

	body := res.Body
	defer body.Close()

	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		log.Fatalf("unable to read body buffer: %s\n",
			err.Error())
	}

	if err = os.WriteFile(path, bodyBytes, 0); err != nil {
		log.Fatalf("unable to write file: %s\n",
			err.Error())
	}
}

func main() {
	configureProgramByArgs()

	err := os.MkdirAll(programConfig.cacheDirPath, 0)
	if err != nil {
		log.Fatalf("unable to create directory %s: %s\n",
			programConfig.cacheDirPath, err.Error())
	}

	log.Printf("cachedir %s ok\n", programConfig.cacheDirPath)

	attempt := 1
	for attempt <= 2 {

		myFilePath, myFileExists := checkFile()

		if myFileExists {
			if !programConfig.skipChecksum {
				if sha256Checksum(myFilePath) != programConfig.checksum {
					log.Println("user requested checksum verification")
					log.Println("checksums do not match")
					if !programConfig.skipDownload {
						if os.Remove(myFilePath) != nil {
							log.Fatalf("unable to remove file %s: %s\n",
								myFilePath, err.Error())
						}

						attempt++
					} else {
						log.Fatalln("terminating now") //EXITING
					}
				} else {
					log.Println("checksums match")
					break // SUCCESS
				}
			} else {
				break // SUCCESS
			}
		} else {
			if programConfig.skipDownload {
				log.Println("user requested not to download anything")
				log.Fatalln("terminating now")
			} else {
				log.Println("downloading file...")
				fileDownload(myFilePath)
				if programConfig.skipChecksum {
					attempt++
				}
			}
		}
	}

	if attempt > 2 {
		log.Println("max attempts reached")
		log.Fatalln("terminating now")
	}
}
