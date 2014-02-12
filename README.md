# osx-tmpfs

A suid helper to create a secure ramdisk on OS X; the ramdisk it creates is mounted in-kernel, using wired memory. (This is why this program is suid.)

How to use:
    ./osx-tmpfs

Output, on success:
     {diskpath}
     {mountpoint}

There are no options that affect its behavior. Change the constants and recompile; the fewer options a suid binary has, the better.

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

See `man 1 hdiutil`
