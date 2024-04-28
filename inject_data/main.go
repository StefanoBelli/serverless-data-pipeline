package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

const DEFAULT_CSV_SEPARATOR = ","

const DEFAULT_API_ENDPOINT = "API gateway endpoint"

const DEFAULT_EVERY_MS = "3000"
const DEFAULT_DIRTY_DATA = "true"
const DEFAULT_DIRTY_THRESHOLD = "0.8"

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

	generator struct {
		dirtyThresh float32
		dirtyData   bool
		everyMs     int
	}

	injector struct {
		apiEndpoint string
	}

	csv struct {
		separator string
	}
}

type Argument struct {
	name        string
	description string
	needsValue  bool
	handler     func(string)
	defValue    string
}

var programConfig Config
var programArguments []Argument = []Argument{
	{
		name:        "--checksum",
		description: "Enable custom checksum value verification (SHA256)",
		needsValue:  true,
		defValue:    EXPECTED_FILE_SHA256_CHECKSUM,
		handler: func(val string) {
			programConfig.checksum = val
			programConfig.skipChecksum = false
			if len(programConfig.checksum) != 64 {
				log.Fatalf("SHA256 checksum hex-encoded strings " +
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
		defValue:    DEFAULT_FILE_NAME,
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
		defValue:    DEFAULT_CACHEDIR_RELNAME,
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
		defValue:    DEFAULT_URL,
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
	{
		name:        "--every-ms",
		description: "Generate an entry every X ms",
		needsValue:  true,
		defValue:    DEFAULT_EVERY_MS,
		handler: func(value string) {
			evMs, err := strconv.ParseInt(value, 10, 32)
			if err != nil {
				log.Fatalln("unable to parse int")
			}

			programConfig.generator.everyMs = int(evMs)
		},
	},
	{
		name:        "--every-ms-default",
		description: "Generate an entry every default ms delay",
		needsValue:  false,
		handler: func(value string) {
			evMs, _ := strconv.ParseInt(DEFAULT_EVERY_MS, 10, 32)
			programConfig.generator.everyMs = int(evMs)
		},
	},
	{
		name:        "--dirty-data",
		description: "Enable generation of dirty data (random)",
		needsValue:  false,
		defValue:    DEFAULT_DIRTY_DATA,
		handler: func(_ string) {
			programConfig.generator.dirtyData = true
		},
	},
	{
		name:        "--no-dirty-data",
		description: "Disable generation of dirty data (random)",
		needsValue:  false,
		handler: func(_ string) {
			programConfig.generator.dirtyData = false
		},
	},
	{
		name:        "--dirty-thresh",
		description: "Set thresh. to dirty data. Check against PRNG-generated num in (0,1).",
		needsValue:  true,
		defValue:    DEFAULT_DIRTY_THRESHOLD,
		handler: func(value string) {
			dThr, err := strconv.ParseFloat(value, 32)
			if err != nil {
				log.Fatalln("unable to parse float")
			}

			programConfig.generator.dirtyThresh = float32(dThr)
		},
	},
	{
		name:        "--dirty-thresh-default",
		description: "Set thresh. to dirty data (default). Check against PRNG-generated num in (0,1).",
		needsValue:  false,
		handler: func(value string) {
			dThr, _ := strconv.ParseFloat(DEFAULT_DIRTY_THRESHOLD, 32)
			programConfig.generator.dirtyThresh = float32(dThr)
		},
	},
	{
		name:        "--api-endpoint",
		description: "Set API gateway endpoint to send data to",
		needsValue:  true,
		defValue:    DEFAULT_API_ENDPOINT,
		handler: func(value string) {
			programConfig.injector.apiEndpoint = value
		},
	},
	{
		name:        "--api-endpoint-default",
		description: "Set default API gateway endpoint to send data to",
		needsValue:  false,
		handler: func(value string) {
			programConfig.injector.apiEndpoint = DEFAULT_API_ENDPOINT
		},
	},
	{
		name:        "--csv-separator",
		description: "Set custom csv column separator",
		needsValue:  true,
		defValue:    DEFAULT_CSV_SEPARATOR,
		handler: func(value string) {
			if len(value) != 1 {
				log.Fatalf("%s is not a valid separator - must be 1 chr long\n",
					value)
			}
			programConfig.csv.separator = value
		},
	},
	{
		name:        "--csv-separator-default",
		description: "Set default csv column separator",
		needsValue:  false,
		handler: func(value string) {
			programConfig.csv.separator = DEFAULT_CSV_SEPARATOR
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
	for _, arg := range programArguments {
		fmt.Printf(" * %s - %s, needs value: %t\n",
			arg.name, arg.description, arg.needsValue)
		if arg.defValue != "" {
			fmt.Printf("  \tdefault: %s\n", arg.defValue)
		}
	}

	os.Exit(0)
}

func loadDefaults() {
	dirtyData, err := strconv.ParseBool(DEFAULT_DIRTY_DATA)
	if err != nil {
		log.Fatalln("unable to parse bool (loading defaults)")
	}

	dirtyThresh, err := strconv.ParseFloat(DEFAULT_DIRTY_THRESHOLD, 32)
	if err != nil {
		log.Fatalln("unable to parse float (loading defaults)")
	}

	evMs, err := strconv.ParseInt(DEFAULT_EVERY_MS, 10, 32)
	if err != nil {
		log.Fatalln("unable to parse int (loading defaults)")
	}

	dflCacheDir := getHomeDir() + "/" + DEFAULT_CACHEDIR_RELNAME

	programConfig.filename = DEFAULT_FILE_NAME
	programConfig.cacheDirPath = dflCacheDir
	programConfig.checksum = EXPECTED_FILE_SHA256_CHECKSUM
	programConfig.downloadUrl = DEFAULT_URL
	programConfig.skipChecksum = DEFAULT_SKIP_CHECKSUM
	programConfig.skipDownload = DEFAULT_SKIP_DOWNLOAD
	programConfig.injector.apiEndpoint = DEFAULT_API_ENDPOINT
	programConfig.generator.dirtyData = dirtyData
	programConfig.generator.dirtyThresh = float32(dirtyThresh)
	programConfig.generator.everyMs = int(evMs)

	programConfig.csv.separator = DEFAULT_CSV_SEPARATOR
	if len(programConfig.csv.separator) != 1 {
		log.Fatalf("%s is not a valid separator - must be 1 chr long\n",
			programConfig.csv.separator)
	}
}
func configureProgramByArgs() {
	loadDefaults()

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

	if err = os.WriteFile(path, bodyBytes, 0644); err != nil {
		log.Fatalf("unable to write file: %s\n",
			err.Error())
	}
}

func main() {
	configureProgramByArgs()

	err := os.MkdirAll(programConfig.cacheDirPath, 0700)
	if err != nil {
		log.Fatalf("unable to create directory %s: %s\n",
			programConfig.cacheDirPath, err.Error())
	}

	log.Printf("cachedir %s ok\n", programConfig.cacheDirPath)

	var fullFilePath string

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
					fullFilePath = myFilePath
					break // SUCCESS
				}
			} else {
				fullFilePath = myFilePath
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

	injectErr := inject(fullFilePath)
	if injectErr != nil {
		log.Fatalf("injector got an error: %s\n",
			injectErr.Error())
	}
}
