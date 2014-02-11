package main

import (
	"fmt"
	"os/exec"
	"os/user"
	"github.com/golang/glog"
)

const (
	acluser = "user:%s:allow delete,readattr,writeattr,readextattr,writeextattr,readsecurity,writesecurity,chown,list,search,add_file,add_subdirectory,delete_child,read,write,execute,append,file_inherit,directory_inherit\neveryone:deny delete,readattr,writeattr,readextattr,writeextattr,readsecurity,writesecurity,chown,list,search,add_file,add_subdirectory,delete_child,read,write,execute,append,file_inherit,directory_inherit\n"
	perms = "go-rwx"
)

func ChPrivateDir(path string, uid string) (err error) {
	// Ensure that the permissions are correct.
	// -P: don't follow symlinks
	// (Go's os.Chmod always follows symlinks)
	out, err := exec.Command("/bin/chmod", "-P", "-R", perms, path).CombinedOutput()
	if err != nil {
		glog.Errorf("/bin/chmod failed:\n%s", string(out))
		return err
	}
	glog.Infof("%s", out)

	// chflags
	// -P: don't follow symlinks
	// hidden: hide from GUI (TODO: is this desirable)
	out, err = exec.Command("/usr/bin/chflags", "-P", "-R", "-v", "-v", "hidden", path).CombinedOutput()
	if err != nil {
		glog.Errorf("/usr/bin/chflags failed:\n%s", string(out))
		return err
	}
	glog.Infof("chflags %s", string(out))

	username, err := user.LookupId(uid)
	if err != nil {
		glog.Errorf("Couldn't set ACL because username lookup failed")
		// TODO: is this, in fact, true?
	}

	// Set the extended attributes; we pipe the ACLs into chmod -E
	cmd := exec.Command("/bin/chmod", "-P", "-R", "-E", path)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		glog.Errorf("error opening pipe")
		return err
	}
	fmt.Fprintf(stdin, acluser, "dlg")
	stdin.Close()
	glog.Infof("setting acl to: %s", fmt.Sprintf(acluser, "dlg"))
	out, err = cmd.CombinedOutput()
	if err != nil {
		glog.Errorf("error setting acl: %s", out)
		return err
	}
	glog.Infof("setting acl: %s", out)
	return nil
	// TODO: verify acls were, in fact, set correctly
	// TODO: add a no backup xattr
}
