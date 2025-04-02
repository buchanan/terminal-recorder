package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/syslog"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	pb "buchanan/recorder/pb"
	pty "buchanan/recorder/pty"

	timestamp "github.com/golang/protobuf/ptypes/timestamp"
	flag "github.com/spf13/pflag"
	viper "github.com/spf13/viper"
	terminal "golang.org/x/crypto/ssh/terminal"
)

// TODO have some persistence if data cannot be written to network out (buffer/reconnect)

var (
	copyBuf   int = 32 * 1024 //copyBuf is used by reader is the max length of a keystroke
	errBuf    bytes.Buffer
	errLog    *log.Logger = log.New(io.Writer(&errBuf), "RECORDER: ", 0)
	Userlogin *user.User
	Hostname  string
	output    io.Writer
	outputL   sync.Mutex
	startTime time.Time = time.Now()
	cleanup   []func() error
	winsize   pty.Winsize = pty.Winsize{
		Rows: 24,
		Cols: 80,
	}
)

func init() {
	var err error
	Hostname, err = os.Hostname()
	if err != nil {
		panic(err)
	}
	// Identify current user
	Userlogin, err = user.Current()
	if err != nil {
		panic(err)
	}

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
	viper.SetDefault("idle", 1)

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
	// flag.String("output", "", "path to save recoreding")
	// flag.Bool("stdin", false, "enable stdin recording, disabled by default")
	// flag.Int("idle", 1, "limit recorded idle time to given number of seconds")
	// flag.String("title", "", "title of the record")
	// flag.StringP("command", "c", "sh", "command to record, defaults to sh")
	// flag.Parse()
	viper.BindPFlags(flag.CommandLine)

	// Final overwrites

	// Open local outfile
	var outfileFlags int = os.O_WRONLY | os.O_CREATE | os.O_EXCL
	outfile, err := os.OpenFile(viper.GetString("recordPath"), outfileFlags, 0644)
	if err != nil {
		errLog.Printf("%s\r\n", err.Error())
	} else {
		cleanup = append(cleanup, outfile.Close)
	}

	// Connect to network outfile
	netout, err := net.Dial("tcp", "127.0.0.1:2110")
	if err != nil {
		panic(err)
	} else {
		cleanup = append(cleanup, netout.Close)
	}
	// TODO

	if outfile == nil && netout == nil {
		errLog.Printf("%s\r\n", "Cannot open local or network outfiles. Exiting")
		Cleanup()
		return
	} else {
		output = io.MultiWriter(outfile, netout)
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

	//var pseudoTerm *os.File

	// Craft header
	//TODO add environment to header
	encodedTheme, _ := json.Marshal(struct {
		FG      string `json:"fg"`
		BG      string `json:"bg"`
		Palette string `json:"palette"`
	}{
		FG:      "#d0d0d0",
		BG:      "#212121",
		Palette: "#151515:#ac4142:#7e8e50:#e5b567:#6c99bb:#9f4e85:#7dd6cf:#d0d0d0:#505050:#ac4142:#7e8e50:#e5b567:#6c99bb:#9f4e85:#7dd6cf:#f5f5f5",
	})
	if message, err := pb.CreateMessage(&pb.Header{
		Version: 1,
		Timestamp: &timestamp.Timestamp{
			Seconds: time.Now().UnixNano(),
		},
		Idle:    viper.GetFloat64("idle"),
		Command: viper.GetString("command"),
		Title:   viper.GetString("title"),
		// Pre-Encode theme
		Theme:    encodedTheme,
		Host:     Hostname,
		Username: Userlogin.Username,
		Terminal: &pb.Terminal{
			Offset: int64(0),
			Width:  uint32(winsize.Cols),
			Height: uint32(winsize.Rows),
		},
	}); err == nil {
		//Write header to outputs
		output.Write(message)
	}
	//Start command
	commandSlice := strings.Split(viper.GetString("command"), " ")
	command := exec.Command(commandSlice[0], commandSlice[1:]...)
	command.Env = os.Environ()
	if Userlogin != nil {
		command.Dir = Userlogin.HomeDir
	}

	if pseudoTerm, err := pty.StartWithSize(command, &winsize); err == nil {
		cleanup = append(cleanup, pseudoTerm.Close)
		Snoop(output, pseudoTerm)
	} else {
		errLog.Printf("%s\r\n", err.Error())
		Cleanup()
		fmt.Println("ERROR: unable to allocate pty")

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
}

func Cleanup() {
	for i := 0; i < len(cleanup); i++ {
		if err := cleanup[i](); err != nil {
			log.Printf("%s\r\n", err.Error())
		}
	}
	cleanup = nil
}

func ScanLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		return i + 1, data[0 : i+1], nil
	}
	if atEOF {
		return len(data), data, nil
	}
	return 0, nil, nil
}

func Snoop(output io.Writer, pseudoTerm *os.File) {
	Rin := &KeyRecorder{
		typeInput:    true,
		outputWriter: output,
	}
	Rout := &KeyRecorder{
		typeInput:    false,
		outputWriter: output,
	}

	in := io.MultiWriter(pseudoTerm, Rin)
	// in := pseudoTerm
	out := io.MultiWriter(os.Stdout, Rout)

	go io.CopyBuffer(in, os.Stdin, make([]byte, copyBuf))
	go io.CopyBuffer(out, pseudoTerm, make([]byte, copyBuf))

	// Notify if window size changes or shell dies
	sig := make(chan os.Signal, 2)
	signal.Notify(sig, syscall.SIGWINCH, syscall.SIGCLD)
	for {
		switch <-sig {
		case syscall.SIGWINCH:
			if ws, err := pty.GetsizeFull(os.Stdin); err != nil {
				errLog.Printf("%s\r\n", err.Error())
			} else {
				pty.Setsize(pseudoTerm, ws)
				//Write out terminal info
				if message, err := pb.CreateMessage(&pb.Terminal{
					Offset: time.Since(startTime).Nanoseconds(),
					Width:  uint32(winsize.Cols),
					Height: uint32(winsize.Rows),
				}); err == nil {
					outputL.Lock()
					output.Write(message)
					outputL.Unlock()
				}
			}
		default:
			return
		}
	}
}

type KeyRecorder struct {
	typeInput    bool
	outputWriter io.Writer
}

func (R *KeyRecorder) Write(data []byte) (int, error) {
	k := pb.Key{
		Key:    make([]byte, len(data)),
		Offset: time.Since(startTime).Nanoseconds(),
		Input:  R.typeInput,
	}
	copy(k.Key, data)

	message, err := pb.CreateMessage(&k)
	if err != nil {
		return 0, err
	}
	outputL.Lock()
	_, err = R.outputWriter.Write(message)
	if err != nil {
		return 0, err
	}
	outputL.Unlock()
	return len(data), err
}
