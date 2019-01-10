package main

import (
	pb "./pb"
	"bytes"
	"encoding/json"
	"github.com/golang/protobuf/proto"
	timestamp "github.com/golang/protobuf/ptypes/timestamp"
	"github.com/kr/pty"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh/terminal"
	"io"
	"log"
	"log/syslog"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// TODO cook/buffer stdin until Enter/Return and store string for searches

var (
	errBuf    bytes.Buffer
	errLog    *log.Logger = log.New(io.Writer(&errBuf), "RECORDER: ", 0)
	Userlogin *user.User
	outfile   *os.File
	netout    *os.File
	startTime time.Time = time.Now()
	cleanup   []func() error
	winsize   pty.Winsize = pty.Winsize{
		Rows: 24,
		Cols: 80,
	}
)

func init() {
	// Identify current user
	Userlogin, _ := user.Current()

	// Try to open syslog
	if W, err := syslog.New(syslog.LOG_WARNING|syslog.LOG_DAEMON, "recorder"); err == nil {
		errLog = log.New(W, "", 0)
		cleanup = append(cleanup, W.Close)
		// syslog failed try to open homedir errlog
	} else if Userlogin != nil {
		if CheckDir(filepath.Join(Userlogin.HomeDir, ".recorder")) {
			if fh, err := os.OpenFile(filepath.Join(Userlogin.HomeDir, ".recorder", "error_log"), os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644); err == nil {
				errLog = log.New(fh, "", log.LstdFlags)
			}
		}
	}

	// Set Defaults
	if Userlogin != nil && CheckDir(filepath.Join(Userlogin.HomeDir, ".recorder")) {
		viper.SetDefault("recordPath", filepath.Join(Userlogin.HomeDir, ".recorder", strconv.FormatInt(startTime.UnixNano(), 10)+".cast"))
	} else if CheckDir("/var/log/recorder") {
		viper.SetDefault("recordPath", filepath.Join("/var/log/recorder", strconv.FormatInt(startTime.UnixNano(), 10)+".cast"))
	}
	viper.SetDefault("recordStdin", false)
	viper.SetDefault("command", "sh") // TODO lookup users default shell
	//viper.SetDefault("commandEnvironment", []string{"SHELL", "TERM"})
	viper.SetDefault("title", "")
	viper.SetDefault("idle", 0)

	//Read config from user first then global
	viper.SetConfigName("recorder.conf")
	if Userlogin != nil {
		viper.AddConfigPath(filepath.Join(Userlogin.HomeDir, ".recorder"))
	}
	viper.AddConfigPath("/etc/recorder/")
	if err := viper.ReadInConfig(); err != nil {
		errLog.Printf("Error reading config file: %s\r\n", err)
	}

	// Check ENVIRONMENT
	// viper.BindEnv("command", "SHELL") // dont use recorder

	// Parse flags
	//flag.String("output", "", "path to save recoreding")
	//flag.Bool("stdin", false, "enable stdin recording, disabled by default")
	flag.Int("idle", 0, "limit recorded idle time to given number of seconds")
	flag.String("title", "", "title of the record")
	flag.StringP("command", "c", "sh", "command to record, defaults to sh")
	flag.Parse()
	viper.BindPFlags(flag.CommandLine)

	// Final overwrites

	// Open local outfile
	var outfileFlags int = os.O_WRONLY | os.O_CREATE | os.O_EXCL
	if fh, err := os.OpenFile(viper.GetString("recordPath"), outfileFlags, 0644); err != nil {
		errLog.Printf("%s\r\n", err.Error())
	} else {
		outfile = fh
		cleanup = append(cleanup, fh.Close)
	}

	// Connect to network outfile
	// TODO

	if outfile == nil && netout == nil {
		errLog.Printf("%s\r\n", "Cannot open local or network outfiles. Exiting")
		Cleanup()
		return
	}

	// Initialize pseudoterminal
	if stdinFD := int(os.Stdin.Fd()); terminal.IsTerminal(stdinFD) {
		// Handle pty size.
		if stdinWS, err := pty.GetsizeFull(os.Stdin); err != nil {
			errLog.Printf("%s\r\n", err.Error())
		} else {
			winsize = *stdinWS
		}

		// Set stdin in raw mode.
		if oldState, err := terminal.MakeRaw(stdinFD); err != nil {
			errLog.Printf("%s\r\n", err.Error())
			Cleanup()
			return
		} else {
			_ = oldState
			cleanup = append(cleanup, func() error { terminal.Restore(stdinFD, oldState); return nil })
		}
	}

	// Write header
	if outfile != nil {
		var encodedTheme []byte
		if tb, err := json.Marshal(struct {
			FG      string `json:"fg"`
			BG      string `json:"bg"`
			Palette string `json:"palette"`
		}{FG: "#d0d0d0", BG: "#212121", Palette: "#151515:#ac4142:#7e8e50:#e5b567:#6c99bb:#9f4e85:#7dd6cf:#d0d0d0:#505050:#ac4142:#7e8e50:#e5b567:#6c99bb:#9f4e85:#7dd6cf:#f5f5f5"}); err == nil {
			encodedTheme = tb
		}
		fileHeader := pb.HEADER{
			Version:   1,
			Width:     uint32(winsize.Cols),
			Height:    uint32(winsize.Rows),
			Timestamp: &timestamp.Timestamp{Seconds: startTime.Unix()},
			Idle:      viper.GetFloat64("idle"),
			Command:   viper.GetString("command"),
			Title:     viper.GetString("title"),
			Theme:     encodedTheme,
		}
		if b, err := proto.Marshal(&fileHeader); err == nil {
			if _, err := outfile.Write(b); err != nil {
				panic(err)
			}
		} else {
			panic(err)
		}
		// json.NewEncoder(outfile).Encode(struct {
		// 	Version   int     `json:"version"`
		// 	Width     uint16  `json:"width"`
		// 	Height    uint16  `json:"height"`
		// 	Timestamp int64   `json:"timestamp"`
		// 	Idle      float64 `json:"idle_time_limit"`
		// 	Command   string  `json:"command"`
		// 	Title     string  `json:"title"`
		// 	//Env       map[string]string
		// 	Theme struct {
		// 		FG      string `json:"fg"`
		// 		BG      string `json:"bg"`
		// 		Palette string `json:"palette"`
		// 	}
		// }{
		// 	Version:   2,
		// 	Width:     winsize.Cols,
		// 	Height:    winsize.Rows,
		// 	Timestamp: startTime.Unix(),
		// 	Idle:      viper.GetFloat64("idle"),
		// 	Command:   viper.GetString("command"),
		// 	Title:     viper.GetString("title"),
		// 	//Env:       viper.GetStringMapString("commandEnvironment"),
		// 	Theme: struct {
		// 		FG      string `json:"fg"`
		// 		BG      string `json:"bg"`
		// 		Palette string `json:"palette"`
		// 	}{
		// 		FG:      "#d0d0d0",
		// 		BG:      "#212121",
		// 		Palette: "#151515:#ac4142:#7e8e50:#e5b567:#6c99bb:#9f4e85:#7dd6cf:#d0d0d0:#505050:#ac4142:#7e8e50:#e5b567:#6c99bb:#9f4e85:#7dd6cf:#f5f5f5",
		// 	},
		// })
	}
}

func CheckDir(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0755); err == nil {
			return true
		}
		return false
	}
	return true
}

func main() {
	defer Cleanup()
	os.Stdout.Write(errBuf.Bytes())

	var pseudoTerm *os.File
	if outfile != nil || netout != nil { // When would this happen TEST
		// TODO move write header to here
		commandSlice := strings.Split(viper.GetString("command"), " ")
		command := exec.Command(commandSlice[0], commandSlice[1:]...)
		command.Env = os.Environ()
		if Userlogin != nil {
			command.Dir = Userlogin.HomeDir
		}

		pt, err := pty.StartWithSize(command, &winsize)
		if err == nil {
			pseudoTerm = pt
			cleanup = append(cleanup, pt.Close)
		} else {
			errLog.Printf("%s\r\n", err.Error())
		}
	}
	if pseudoTerm == nil { // When would this happen TEST
		Cleanup()
		commandSlice := strings.Split(viper.GetString("command"), " ")
		command := exec.Command(commandSlice[0], commandSlice[1:]...)
		command.Env = os.Environ()
		if Userlogin != nil {
			command.Dir = Userlogin.HomeDir
		}
		command.Stdin = os.Stdin
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr
		command.Run()
		os.Exit(129)
	}

	// tee copied from io.tee
	var teeIN io.Reader = os.Stdin
	var teeOUT io.Reader = &tee{
		r:   pseudoTerm,
		w:   outfile,
		tag: "o",
	}
	// stdin true by default, use os.Stdin unless flag is passed
	if viper.GetBool("recordStdin") {
		teeIN = &tee{
			r:   os.Stdin,
			w:   outfile,
			tag: "i",
		}
	}

	// Copy stdin to the pty and the pty to stdout.
	go func() {
		io.Copy(pseudoTerm, teeIN)
	}()
	io.Copy(os.Stdout, teeOUT)
}

type tee struct {
	r   io.Reader
	w   io.Writer
	tag string
}

func (t *tee) Read(p []byte) (n int, err error) {
	n, err = t.r.Read(p)
	if n > 0 {
		keystroke := pb.COMMAND_KEY{
			Offset: time.Since(startTime).Seconds(),
			Tag:    t.tag,
			Key:    string(p[:n]),
		}
		if b, merr := proto.Marshal(&keystroke); merr == nil {
			_, err = t.w.Write(b)
		} else {
			err = merr
		}
		// err = json.NewEncoder(t.w).Encode([]interface{}{
		// 	time.Since(startTime).Seconds(),
		// 	t.tag,
		// 	string(p[:n]),
		// })
	}
	return n, err
}

func Cleanup() {
	for i := 0; i < len(cleanup); i++ {
		if err := cleanup[i](); err != nil {
			log.Printf("%s\r\n", err.Error())
		}
	}
	cleanup = nil
}
