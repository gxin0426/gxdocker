package container

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/sirupsen/logrus"
)

func RunContainerInitProcess() error {
	setupMount()
	cmdArray := readUserCommand()
	if cmdArray == nil || len(cmdArray) == 0 {
		return fmt.Errorf("Run container get user command error cmdArray is nil")
		
	}
	fmt.Println(cmdArray)
	fmt.Println("commarry : ", cmdArray[0])
	path, err := exec.LookPath(cmdArray[0])

	if err != nil {
		logrus.Errorf("exec loop path err %v", err)
		return err
	}
	logrus.Infof("find path %s", path)
	if err := syscall.Exec(path, cmdArray[0:], os.Environ()); err != nil {
		logrus.Errorf(err.Error())
	}
	return nil
}

func readUserCommand() []string {
	pipe := os.NewFile(uintptr(3), "pipe")
	msg, err := ioutil.ReadAll(pipe)

	if err != nil {
		logrus.Errorf("init read pipe error %v", err)
		return nil
	}
	msgStr := string(msg)
	fmt.Println(msgStr)
	return strings.Split(msgStr, " ")
}

func setupMount() {
	pwd, err := os.Getwd()
	// pwd := "/busybox"
	if err != nil {
		logrus.Errorf("get current location error %v", err)
		return
	}
	logrus.Infof("Current location is %s", pwd)
	pivotRoot(pwd)

	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")
	syscall.Mount("tmpfs", "/dev", "tmpfs", syscall.MS_NOSUID|syscall.MS_STRICTATIME, "mode=755")
}

func pivotRoot(root string) error {
	/*
		in order to the newroot and oldroot in different filesystem  we should mount root(string) again
		bind mount is a method which the same content change a mount point
	*/
	if err := syscall.Mount(root, root, "bind", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		return fmt.Errorf("Mount rootfs to itself errr %v", err)
	}
	pivotDir := filepath.Join(root, ".pivot_root")

	if err := os.RemoveAll(pivotDir); err != nil {
		fmt.Println(err)
	}

	if err := os.Mkdir(pivotDir, 0777); err != nil {
		return err
	}

	//systemd included linux mount namespace become shared by default  you should  explicit formulation declare
	// the new mount namespace is independent
	if err := syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, ""); err != nil {
		fmt.Println(err)
	}

	if err := syscall.PivotRoot(root, pivotDir); err != nil {
		return fmt.Errorf("pivot_root %v", err)
	}

	if err := syscall.Chdir("/"); err != nil {
		return fmt.Errorf("umount pivot_root dir %v", err)
	}

	pivotDir = filepath.Join("/", ".pivot_root")
	if err := syscall.Unmount(pivotDir, syscall.MNT_DETACH); err != nil {
		return fmt.Errorf("umount pivot_root dir %v", err)
	}

	return os.Remove(pivotDir)
}
