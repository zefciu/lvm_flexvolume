package lvm_utils

import "bytes"
import "errors"
import "fmt"
import "os"
import "os/exec"
import "strings"
import "syscall"
import "time"

type LogFileWrap struct {
	LogFile *os.File
}

var LogFile = LogFileWrap{}

type Lv struct {
	Vg   string
	Name string
	Path string
}

func CallCmd(name string, args ...string) (string, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	var err error
	Log(fmt.Sprintf("%s %s", name, strings.Join(args, " ")))
	cmd := exec.Command(name, args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	done := make(chan error, 1)
	go func() {
		cmd.Run()
		done <- cmd.Wait()
	}()
	select {
	case <-time.After(10 * time.Second):
		err = cmd.Process.Signal(syscall.SIGINT)
		if err == nil {
			err = errors.New("Killed the command")
		}
	case err = <-done:
	}
	Log(stderr.String())
	Log(stdout.String())
	if err != nil {
		return "", errors.New(stderr.String() + err.Error())
	}
	return stdout.String(), nil

}

func init() {
	logFile, err := os.Create("/tmp/lvmdriver.log")
	if err != nil {
		panic(fmt.Sprintf("Cannot create logfile: %s", err.Error()))
	}
	LogFile.LogFile = logFile
}

func Cleanup() {
	LogFile.LogFile.Close()
}

func Log(data string) {
	_, err := LogFile.LogFile.Write([]byte(data + "\n"))
	if err != nil {
		panic(fmt.Sprintf("Cannot log data: %s", err.Error()))
	}

}

func GetLV(vgName, lvName string) (*Lv, error) {
	data, err := CallCmd(
		"lvs",
		vgName+"/"+lvName,
		"--separator", ";",
		"--no-heading",
		"-o", "lv_name,lv_path",
	)
	if err != nil || data == "" {
		return nil, nil
	}
	dataParsed := strings.SplitN(strings.Trim(data, " \n\t"), ";", 2)
	return &Lv{Vg: vgName, Name: dataParsed[0], Path: dataParsed[1]}, nil
}

func CreateLV(
	vg string, pool string, volId string, size uint64, fs string,
) (*Lv, error) {
	_, err := CallCmd(
		"lvcreate",
		"-v",
		fmt.Sprintf("-L%dB", size),
		// "-T", fmt.Sprintf("%s/%s", vg, pool),
		"-n", volId,
		vg,
	)
	if err != nil {
		return nil, err
	}
	lv, err := GetLV(vg, volId)
	if err != nil {
		return nil, err
	}
	_, err = CallCmd("mkfs", "-t", fs, lv.Path)
	if err != nil {
		return nil, err
	}
	return lv, nil
}

func EnsureDevice(
	vg string, pool string, volId string, size uint64, fs string,
) (*Lv, error, bool) {
	lv, err := GetLV(vg, volId)
	if err != nil {
		return nil, err, false
	}
	if lv != nil {
		return lv, nil, false
	}
	lv, err = CreateLV(vg, pool, volId, size, fs)
	if err != nil {
		return nil, err, false
	}
	return lv, nil, true
}
