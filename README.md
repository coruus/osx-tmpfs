# osx-tmpfs

A suid helper to create a secure ramdisk on OS X.

How to use:
  ./osx-tmpfs

Output, on success:
  {diskpath}\n{mountpoint}\n

There are no options that affect its behavior; you can control logging
by the standard `glog` options.

An outline of how it works:
  - Generate a random encryption key
  - Attach a ram:// virtual disk in-kernel encrypted under that key
	- Create an HFS+ filesystem on the disk
	- Create a randomly named mountpoint in the current directory with
		the `.noindex` extension
	- Mount the filesystem
	- Change ownership to the real uid; set restrictive permissions
	  and extended attributes
	- Output the name of the ramdisk and the mountpoint

According to some OS X documentation (and experiments on previous
OS X versions), using `hdik` rather than `hdiutil` to create the ramdisk causes the ramdisk's memory to be pinned.

TODO: verify that this is indeed the case on Mavericks
