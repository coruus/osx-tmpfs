package main

import (
	"crypto/rand"
	"encoding/base32"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strconv"

	"github.com/golang/glog"
)

const (
	ramdisk_uri = "ram://2048"
)

func randomBytes(c int) (b []byte, err error) {
	if err != nil {
		return b, err
	}
	return b, nil
}

func createPinnedRamdisk() (disk string, err error) {
	// Creating an in-kernel ramdisk supposedly ensures that the memory
	// is pinned; see man 1 hdiutil.
	// TODO: Why doesn't encryption work with ramdisks? "-encryption", enckey,
	hdik := exec.Command("/usr/sbin/hdik", "-nomount", ramdisk_uri)
	out, err := hdik.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("hdik error: %s\n%s", err, out)
	}
	ramdiskre, _ := regexp.Compile("/dev/disk[0-9]+")
	ramdisk := ramdiskre.Find(out)
	if ramdisk == nil {
		return "", fmt.Errorf("hdik didn't return a disk; output:\n %s", out)
	}
	return string(ramdisk), nil
}

func createFs(ramdisk string, uid int) (volname string, err error) {
	// Generate a random volume name.
	b := make([]byte, 20)
	_, err = rand.Read(b)
	if err != nil {
		return "", errors.New("couldn't generate random volume name")
	}
	volname = base32.StdEncoding.EncodeToString(b) + ".noindex"

	// Create the new filesystem
	// -P: set kHFSContentProtectionBit
	// -M 700: default mode=700
	// (newfs_hfs is relatively safe; it won't unmount and erase a mounted disk)
	newfs := exec.Command("/sbin/newfs_hfs", "-v", volname, "-U", strconv.Itoa(uid), "-G", "admin", "-M", "700", "-P", ramdisk)
	out, err := newfs.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("newfs_hfs error: %s", out)
	}
	return volname, nil
}

func main() {
	// Parse flags for glog
	//flag.Parse()

	uid := os.Getuid()

	// Check that the current directory is owned by the real uid.
	fi, err := os.Lstat(".")
	if err != nil {
		glog.Errorf("Error statting current directory: %s", err)
		os.Exit(255)
	}
	if !fi.IsDir() {
		// TODO: check ownership and go-rw
		glog.Errorf("Current directory not a directory.")
		os.Exit(255)
	}

	// Create the ramdisk
	ramdisk, err := createPinnedRamdisk()
	if err != nil {
		glog.Errorf("createPinnedRamdisk: %s", err)
		os.Exit(255)
	}

	// Create the filesystem
	volname, err := createFs(ramdisk, uid)
	if err != nil {
		glog.Errorf("createFs: %s", err)
		os.Exit(255)
	}

	// Create the mountpoint
	mountpoint := path.Join("./", volname)
	err = os.Mkdir(mountpoint, 700)
	if err != nil {
		glog.Errorf("couldn't create directory: %s", err)
		os.Exit(255)
	}
	err = os.Lchown(mountpoint, uid, -1)
	if err != nil {
		glog.Errorf("couldn't chown directory: %s", err)
		os.Exit(255)
	}
	// Mount the new volume on the mountpoint
	mount := exec.Command("/sbin/mount_hfs", "-u", strconv.Itoa(uid), "-m", "700", "-o", "noatime,nosuid,nobrowse", ramdisk, string(mountpoint))
	out, err := mount.CombinedOutput()
	if err != nil {
		glog.Errorf("couldn't mount new volume: %s\n%s", err, out)
		os.Exit(255)
	}
	glog.Infof("mount_hfs: %s", out)

	err = chprivDir(string(mountpoint), fmt.Sprintf("%d", uid))
	if err != nil {
		glog.Errorf("coudn't make dir safe\n%s", err)
		// TODO: should this cause a non-zero exit status?
	}

	// TODO: disable indexing, trash, and fsevents; probably
	// can drop privileges first
	// rm -rf .fseventsd .Trashes
	// touch .fseventsd/no_log .metadata_never_index .Trashes

	fmt.Printf("%s\n%s", ramdisk, mountpoint)
}
