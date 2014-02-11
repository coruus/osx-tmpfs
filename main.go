package main

import (
	"crypto/rand"
	"encoding/base32"
	"encoding/hex"
	"errors"
	"flag"
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
	b = make([]byte, c)
	_, err = rand.Read(b)
	if err != nil {
		return b, err
	}
	return b, nil
}

func createPinnedRamdisk(enckey string) (disk string, err error) {
	// Creating an in-kernel ramdisk supposedly ensures that the memory
	// is pinned; see man 1 hdiutil.
	// TODO: Why doesn't encryption work with ramdisks? "-encryption", enckey,
	hdik := exec.Command("/usr/sbin/hdik", "-nomount", ramdisk_uri)
	out, err := hdik.CombinedOutput()
	if err != nil {
		return "", errors.New(fmt.Sprintf("hdik error: %s\n%s", err, out))
	}
	ramdiskre, _ := regexp.Compile("/dev/disk[0-9]+")
	ramdisk := ramdiskre.Find(out)
	if ramdisk == nil {
		return "", errors.New(fmt.Sprintf("hdik didn't return a disk; output:\n %s", out))
	}
	return string(ramdisk), nil
}

func main() {
	// Parse flags for glog
	//flag.Parse()

	b, err := randomBytes(20)
	if err != nil {
		glog.Errorf("couldn't generate random volume name")
		os.Exit(255)
	}
	volname := base32.StdEncoding.EncodeToString(b) + ".noindex"
	b, err = randomBytes(36)
	if err != nil {
		glog.Errorf("couldn't generate random encryption key")
		os.Exit(255)
	}
	enckey := hex.EncodeToString(b)

	glog.Infof("volname %s", volname)
	glog.Infof("enckey %s", enckey)

	ramdisk, err := createPinnedRamdisk(enckey)
	if err != nil {
		glog.Fatalf("createPinnedRamdisk %s", err)
	}

	uid := os.Getuid()
	glog.Infof("uid = %d", uid)

	// Create the new filesystem
	// -P: set kHFSContentProtectionBit
	// -M 700: default mode=700
	newfs := exec.Command("/sbin/newfs_hfs", "-v", volname, "-U", strconv.Itoa(uid), "-G", "admin", "-M", "700", "-P", string(ramdisk))
	out, err := newfs.CombinedOutput()
	if err != nil {
		glog.Fatalf("newfs_hfs error: %s", out)
	}
	glog.Infof("newfs_hfs out %s", out)

	// Create the mountpoint
	mountpoint := path.Join("./", volname)
	err = os.Mkdir(mountpoint, 700)
	if err != nil {
		glog.Fatalf("couldn't create directory: %s", err)
	}
	// TODO: check that the current directory is safe first
	err = os.Lchown(mountpoint, uid, -1)
	if err != nil {
		glog.Fatalf("couldn't chown directory: %s", err)
	}

	//
	mount := exec.Command("/sbin/mount_hfs", "-u", strconv.Itoa(uid), "-m", "700", "-o", "noatime,nosuid,nobrowse", ramdisk, string(mountpoint))
	out, err = mount.CombinedOutput()
	if err != nil {
		glog.Fatalf("couldn't mount new volume: %s\n%s", err, out)
	}
	glog.Infof("mount_hfs: %s", out)

	err = ChflagsPrivateDir(string(mountpoint))
	if err != nil {
		glog.Errorf("coudn't make dir safe\n%s", err)
	}

	fmt.Printf("%s\n%s\n", ramdisk, mountpoint)
}
