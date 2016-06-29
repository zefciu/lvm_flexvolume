package main

import "errors"
import "fmt"
import "encoding/json"
import "os"
import "regexp"
import "strconv"
import "strings"

import "github.com/lvm_flexvolume/lvm_utils"

var SizeRE = regexp.MustCompile(
	"^([0-9]+)((?:k|K|m|M|g|G|t|T|p|P)?)$",
)

type AttachOpts struct {
    FsType string `json:"kubernetes.io/fsType"`
    Readwrite string `json:"kubernetes.io/readwrite"`
    RawSize string `json:"size"`
    VolumeID string `json:"volumeID"`
    VolumeGroup string `json:"volumegroup"`
    Pool string `json:"pool"`
}

func (opts AttachOpts) Size() (uint64, error) {
	groups := SizeRE.FindStringSubmatch(opts.RawSize)
	if len(groups) != 3 {
		return 0, errors.New("Invalid size format")
	}
	number, _ := strconv.ParseUint(groups[1], 10, 64) // Already validated by re
	switch strings.ToLower(groups[2]) {
		case "", "k":
			return number * (1 << 10), nil
		case "m":
			return number * (1 << 20), nil
		case "g":
			return number * (1 << 30), nil
		case "t":
			return number * (1 << 40), nil
		case "p":
			return number * (1 << 50), nil
        default:
            panic("Unreachable!")
	}
}

func PrintResult(status string, message string, device string) {
    data := map[string]string{
        "status": status,
        "message": message,
    }
    if device != "" {
        data["device"] = device
    }
    result, _ := json.Marshal(data)
    fmt.Printf("%s\n", result)
}

func Init() {
    PrintResult("Success", "Initialized", "")
}

func Attach(jsonArgStr string) {
    var jsonArgs AttachOpts
    _ = json.Unmarshal([]byte(jsonArgStr), &jsonArgs)
    size, err := jsonArgs.Size()
    if err != nil {
        PrintResult("Failure", err.Error(), "")
    }
    device, err := lvm_utils.EnsureDevice(
		jsonArgs.VolumeGroup,
		jsonArgs.Pool,
		jsonArgs.VolumeID,
		size,
	)
    if err != nil {
        PrintResult("Failure", err.Error(), "")
        return
    }
    PrintResult("Success", "Volume attached", device.Name())
}

func main() {
    switch os.Args[1] {
    case "init":
        Init()
    case "attach":
        Attach(os.Args[2])
    // case "detach":
    //     Nop()
    // case "mount":
    //     Mount(os.Args[2], os.Args[3], os.Args[4])
    // case "unmount":
    //     Unmount(os.Args[2])
    default:
        PrintResult("Failure", "Invalid command", "")
    }
}
