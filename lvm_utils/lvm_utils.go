package lvm_utils

// #cgo LDFLAGS: -llvm2app
// #include <lvm2app.h>
// #include <stdlib.h>
import "C"

import "errors"
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
    if (vg == nil) {
	    return nil, errors.New(C.GoString(C.lvm_errmsg(libHandle)))
    } else {
        return vg, nil
    }
}

func GetLV(vg C.vg_t, lvName string) (*Lv, error) {
    goLvName := C.CString(lvName)
    defer C.free(unsafe.Pointer(goLvName))
    lv := C.lvm_lv_from_name(vg, goLvName);
    if (lv == nil) {
        message := C.GoString(C.lvm_errmsg(libHandle))
        if message == "" {
            return nil, nil // LV doesn't exist
        } else {
            return nil, errors.New(message) // Some error
        }
    }
    return &Lv{lv}, nil
}

func CreateLV(vg C.vg_t, pool string, volId string, size uint64) (*Lv, error) {
    goPool :=C.CString(pool)
    defer C.free(unsafe.Pointer(goPool))
    goVolId :=C.CString(volId)
    defer C.free(unsafe.Pointer(goVolId))
    params := C.lvm_lv_params_create_thin(
        vg, goPool, goVolId, C.uint64_t(size),
    )
    if params == nil {
        return nil, errors.New(C.GoString(C.lvm_errmsg(libHandle)))
    }
    lv := C.lvm_lv_create(params)
    if lv == nil {
        return nil, errors.New(C.GoString(C.lvm_errmsg(libHandle)))
    }
    return &Lv{lv}, nil
}

func (lv Lv) Name() (string) {
    return C.GoString(C.lvm_lv_get_name(lv.lv))
}


func EnsureDevice(
    vgName string, pool string, volId string, size uint64,
) (*Lv, error) {
    vg, err := GetVG(vgName)
    if err != nil {
        return nil, err
    }
    lv, err := GetLV(vg, volId)
    if err != nil {
        return nil, err
    }
    if lv != nil {
        return lv, nil
    }
    lv, err = CreateLV(vg, pool, volId, size)
    if err != nil {
        return nil, err
    }
    return lv, nil
}
