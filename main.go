package main

import (
	"encoding/json"
	"flag"
	"io"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/kr/pty"
	"golang.org/x/crypto/ssh/terminal"
)

var (
	outPath      string
	stdin        bool
	appendOut    bool
	overwriteOut bool
	command      string
	commandEnv   map[string]string
	startTime    time.Time
	title        *string  = new(string)
	idle         *float64 = new(float64)
	errLog       *log.Logger
)

func readConfig(path string) bool {
	var config struct {
		Command    string
		CommandEnv []string
		Title      string
		Idle       float64
	}
	if s, err := os.Stat(path); err == nil && s.Mode().IsRegular() {
		if conFile, err := os.Open(path); err == nil {
			if err := json.NewDecoder(conFile).Decode(&config); err == nil {
				command = config.Command
				if config.CommandEnv != nil {
					commandEnv = make(map[string]string)
					for _, env := range config.CommandEnv {
						commandEnv[env] = os.Getenv(env)
					}
				}
				*title = config.Title
				*idle = config.Idle
				return true
			}
		}
	}
	return false
}

func openLog(path string) {
	if _, err := os.Stat(filepath.Dir(path)); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return
		}
	}
	fh, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err == nil {
		errLog = log.New(fh, "", log.LstdFlags)
	}
}

func init() {
	startTime = time.Now()
	var (
		//Username string = "Unknown"
		Userhome string = "/etc/records"
	)
	if currentUser, err := user.Current(); err == nil {
		//Username = currentUser.Username
		Userhome = currentUser.HomeDir
	}

	if errLog == nil {
		openLog("/var/log/records")
	}
	if errLog == nil {
		openLog(filepath.Join(Userhome, ".records", "error.log"))
	}

	// Find config file
	if readConfig(filepath.Join(Userhome, ".records")) {
		log.Println("Reading config from", filepath.Join(Userhome, ".records"))
	} else if readConfig(filepath.Join("/etc/records/records.conf")) {
		log.Println("Reading config from /etc/records/records.conf")
	} else {
		//log.Println("No config files found using defaults")
	}

	// Parse cli parameters
	//flag.StringVar(&outPath, "output", filepath.Join("/var/log/records/", Username, string(startTime.Unix())+".cast"), "path to save recoreding, defaults to /var/log/records")
	flag.StringVar(&outPath, "output", filepath.Join(Userhome, ".records", strconv.FormatInt(startTime.Unix(), 10)+".cast"), "path to save recoreding")
	flag.BoolVar(&stdin, "stdin", false, "enable stdin recording, disabled by default")
	flag.BoolVar(&appendOut, "append", false, "append to existing recording")
	flag.BoolVar(&overwriteOut, "overwrite", false, "overwrite the file if it already exists")
	cmdArg := flag.String("c", "sh", "command to record, defaults to sh")
	cmdEnv := flag.String("env", "SHELL,TERM", "list of environment variables to capture, defaults to SHELL,TERM")
	flag.StringVar(title, "t", "", "title of the asciicast")
	flag.Float64Var(idle, "i", 0, "limit recorded idle time to given number of seconds")
	flag.Parse()

	// Overwrite config with passed in parameters
	if command == "" && cmdArg != nil && *cmdArg != "" {
		command = *cmdArg
	} else {
		log.Println("Command not specified")
		os.Exit(131)
	}

	if commandEnv == nil {
		commandEnv = make(map[string]string)
		for _, key := range strings.Split(*cmdEnv, ",") {
			commandEnv[key] = os.Getenv(key)
		}
	}
}

func main() {
	// Create arbitrary command.
	commandArgs := strings.Split(command, " ")
	c := exec.Command(commandArgs[0], commandArgs[1:]...)

	var ptmx *os.File
	winsize := pty.Winsize{
		Rows: 24,
		Cols: 80,
	}
	if stdinFD := int(os.Stdin.Fd()); terminal.IsTerminal(stdinFD) {
		// Handle pty size.
		if stdinWS, err := pty.GetsizeFull(os.Stdin); err != nil {
			log.Println(err)
			os.Exit(129)
		} else {
			winsize = *stdinWS
		}

		// Set stdin in raw mode.
		if oldState, err := terminal.MakeRaw(stdinFD); err != nil {
			log.Println(err)
			os.Exit(129)
		} else {
			defer terminal.Restore(stdinFD, oldState) // Best effort.
		}
	}
	if pt, err := pty.StartWithSize(c, &winsize); err != nil {
		log.Println(err)
		os.Exit(129)
	} else {
		// Make sure to close the pty at the end.
		defer pt.Close()
		ptmx = pt
	}

	// OpenOutFile
	if _, err := os.Stat(filepath.Dir(outPath)); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
			log.Println(err)
			os.Exit(130)
		}
	}
	var outfile *os.File
	if overwriteOut {
		if fh, err := os.OpenFile(outPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644); err != nil {
			log.Println(err)
			os.Exit(130)
		} else {
			outfile = fh
			defer fh.Close()
		}
	} else if appendOut {
		if fh, err := os.OpenFile(outPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644); err != nil {
			log.Println(err)
			os.Exit(130)
		} else {
			outfile = fh
			defer fh.Close()
		}
	} else {
		if fh, err := os.OpenFile(outPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644); err != nil {
			log.Println(err)
			os.Exit(130)
		} else {
			outfile = fh
			defer fh.Close()
		}
	}

	// Write header
	if !appendOut {
		json.NewEncoder(outfile).Encode(
			struct {
				Version   int      `json:"version"`
				Width     uint16   `json:"width"`
				Height    uint16   `json:"height"`
				Timestamp int64    `json:"timestamp"`
				Idle      *float64 `json:"idle_time_limit"`
				Command   string   `json:"command"`
				Title     *string  `json:"title"`
				Env       map[string]string
				Theme     struct {
					FG      string `json:"fg"`
					BG      string `json:"bg"`
					Palette string `json:"palette"`
				}
			}{
				Version:   2,
				Width:     winsize.Cols,
				Height:    winsize.Rows,
				Timestamp: startTime.Unix(),
				Idle:      idle,
				Command:   command,
				Title:     title,
				Env:       commandEnv,
				Theme: struct {
					FG      string `json:"fg"`
					BG      string `json:"bg"`
					Palette string `json:"palette"`
				}{
					FG:      "#d0d0d0",
					BG:      "#212121",
					Palette: "#151515:#ac4142:#7e8e50:#e5b567:#6c99bb:#9f4e85:#7dd6cf:#d0d0d0:#505050:#ac4142:#7e8e50:#e5b567:#6c99bb:#9f4e85:#7dd6cf:#f5f5f5",
				},
			})
	}

	// TeeReader copied from io.TeeReader
	var teeIN io.Reader = &teeReader{
		r:   os.Stdin,
		w:   outfile,
		tag: "i",
	}
	var teeOUT io.Reader = &teeReader{
		r:   ptmx,
		w:   outfile,
		tag: "o",
	}
	// stdin true by default, use os.Stdin unless flag is passed
	if stdin {
		teeIN = os.Stdin
	}

	// Copy stdin to the pty and the pty to stdout.
	go func() {
		io.Copy(ptmx, teeIN)
	}()
	io.Copy(os.Stdout, teeOUT)
}

type teeReader struct {
	r   io.Reader
	w   io.Writer
	tag string
}

func (t *teeReader) Read(p []byte) (n int, err error) {
	n, err = t.r.Read(p)
	if n > 0 {
		err = json.NewEncoder(t.w).Encode([]interface{}{
			time.Since(startTime).Seconds(),
			t.tag,
			string(p[:n]),
		})
	}
	return n, err
}
