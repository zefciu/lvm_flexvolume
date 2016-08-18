package lvm_utils

// #cgo LDFLAGS: -llvm2app
// #include <lvm2app.h>
// #include <stdlib.h>
// const char* getString(lvm_property_value_t pv) {
//     return pv.value.string;
// }
import "C"

import "errors"
import "os/exec"
import "unsafe"

var libHandle C.lvm_t

type Lv struct {
	lv C.lv_t
}

func init() {
	path := C.CString("etc/lvm")
	defer C.free(unsafe.Pointer(path))
	libHandle = C.lvm_init(path)
}

func GetVG(vgName string) (C.vg_t, error) {
	goVgName := C.CString(vgName)
	defer C.free(unsafe.Pointer(goVgName))
	w := C.CString("w")
	defer C.free(unsafe.Pointer(w))
	vg := C.lvm_vg_open(libHandle, goVgName, w, 0)
	if vg == nil {
		return nil, errors.New(C.GoString(C.lvm_errmsg(libHandle)))
	} else {
		return vg, nil
	}
}

func GetLV(vg C.vg_t, lvName string) (*Lv, error) {
	goLvName := C.CString(lvName)
	defer C.free(unsafe.Pointer(goLvName))
	lv := C.lvm_lv_from_name(vg, goLvName)
	if lv == nil {
		message := C.GoString(C.lvm_errmsg(libHandle))
		if message == "" {
			return nil, nil // LV doesn't exist
		} else {
			return nil, errors.New(message) // Some error
		}
	}
	return &Lv{lv}, nil
}

func CreateLV(
	vg C.vg_t, pool string, volId string, size uint64, fs string,
) (*Lv, error) {
	goPool := C.CString(pool)
	defer C.free(unsafe.Pointer(goPool))
	goVolId := C.CString(volId)
	defer C.free(unsafe.Pointer(goVolId))
	params := C.lvm_lv_params_create_thin(
		vg, goPool, goVolId, C.uint64_t(size),
	)
	if params == nil {
		return nil, errors.New(C.GoString(C.lvm_errmsg(libHandle)))
	}
	clv := C.lvm_lv_create(params)
	if clv == nil {
		return nil, errors.New(C.GoString(C.lvm_errmsg(libHandle)))
	}
	lv := Lv{clv}
	path := lv.Path()
	cmd := exec.Command("mkfs", "-t", fs, path)
	err := cmd.Run()
	if err != nil {
		return nil, err
	}
	return &lv, nil
}

func (lv Lv) Name() string {
	return C.GoString(C.lvm_lv_get_name(lv.lv))
}

func (lv Lv) Path() string {
	pathString := C.CString("lv_path")
	defer C.free(unsafe.Pointer(pathString))
	property := C.lvm_lv_get_property(lv.lv, pathString)
	return C.GoString(C.getString(property))
}

func EnsureDevice(
	vgName string, pool string, volId string, size uint64, fs string,
) (*Lv, error, bool) {
	vg, err := GetVG(vgName)
	if err != nil {
		return nil, err, false
	}
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
