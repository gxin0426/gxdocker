package container

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/sirupsen/logrus"
)

var (
	RUNNING             string = "running"
	STOP                string = "stopped"
	Exit                string = "exited"
	DefaultInfoLocation string = "/var/run/gxdocker/%s/"
	ConfigName          string = "config.json"
	ContainerLogFile    string = "container.log"
)

type ContainerInfo struct {
	Pid         string `json:"pid"`
	Id          string `json:"id"`
	Name        string `json:"name"`
	Command     string `json:"command"`
	CreatedTime string `json:"createTime"`
	Status      string `json:"status"`
}

func NewParentProcess(tty bool, volume string, containerName string) (*exec.Cmd, *os.File) {
	readPipe, writePipe, err := NewPipe()
	if err != nil {
		logrus.Errorf("New pipe err %v", err)
		return nil, nil
	}
	cmd := exec.Command("/proc/self/exe", "init")
	fmt.Println("newparentprocessde cmd:", cmd)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWNS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC,
	}
	if tty {
		cmd.Stdout = os.Stdout
		cmd.Stdin = os.Stdin
		cmd.Stderr = os.Stderr
	}

	dirURL := fmt.Sprintf(DefaultInfoLocation, containerName)
	if err := os.MkdirAll(dirURL, 0622); err != nil {
		logrus.Errorf("newparentprocess mkdir err %v", err)
		return nil, nil
	}
	stdLogFilePath := dirURL + ContainerLogFile
	stdLogFile, err := os.Create(stdLogFilePath)
	if err != nil {
		logrus.Errorf("create file %s err %v", stdLogFilePath, err)
		return nil, nil
	}

	cmd.Stdout = stdLogFile

	cmd.ExtraFiles = []*os.File{readPipe}
	mntURL := "/root/mnt"
	rootURL := "/root/"
	NewWorkSpace(rootURL, mntURL, volume)
	cmd.Dir = mntURL
	return cmd, writePipe
}

func NewPipe() (*os.File, *os.File, error) {
	read, write, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}
	return read, write, nil
}

func NewWorkSpace(rootURL string, mntURL string, volume string) {
	CreateReadOnlyLayer(rootURL)
	CreateWriteLayer(rootURL)
	CreateMountPoint(rootURL, mntURL)

	if volume != "" {
		volumeURLs := volumeUrlExtract(volume)
		length := len(volumeURLs)
		if length == 2 && volumeURLs[0] != "" && volumeURLs[1] != "" {
			MountVolume(rootURL, mntURL, volumeURLs)
			logrus.Infof("%q", volumeURLs)
		} else {
			logrus.Infof("Volume paramter input is not correct")
		}
	}
}

func MountVolume(rootURL string, mntURL string, volumesURLs []string) {
	parentUrl := volumesURLs[0]
	if err := os.Mkdir(parentUrl, 0777); err != nil {
		logrus.Infof("mkdir parent dir %s error %v", parentUrl, err)
	}

	containerUrl := volumesURLs[1]
	containerVolumeURL := mntURL + containerUrl
	if err := os.Mkdir(containerVolumeURL, 0777); err != nil {
		logrus.Infof("mkdir container dir %s error %v", containerVolumeURL, err)
	}

	dirs := "dirs=" + parentUrl
	cmd := exec.Command("mount", "-t", "aufs", "-o", dirs, "none", containerVolumeURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		logrus.Errorln("mount volume failed")
	}

}

func volumeUrlExtract(volume string) []string {
	var volumeURLs []string
	volumeURLs = strings.Split(volume, ":")
	return volumeURLs
}

func CreateReadOnlyLayer(rootURL string) {
	busyboxURL := rootURL + "busybox/"
	busyboxTarURL := rootURL + "busybox.tar"
	exists, err := PathExists(busyboxURL)
	if err != nil {
		panic("fail to judge whether dir exist")
	}

	if exists == false {
		if err := os.Mkdir(busyboxURL, 0777); err != nil {
			panic("mkdir dir err")
		}
		if _, err := exec.Command("tar", "-xvf", busyboxTarURL, "-C", busyboxURL).CombinedOutput(); err != nil {
			panic("tar -xvf fail")
		}
	}
}

func CreateWriteLayer(rootURL string) {
	writeURL := rootURL + "writeLayer/"
	if err := os.Mkdir(writeURL, 0777); err != nil {
		fmt.Println(err)
		return
	}

}

func CreateMountPoint(rootURL, mntURL string) {
	if err := os.Mkdir(mntURL, 0777); err != nil {
		fmt.Println(err)
		return
	}

	dirs := "dirs=" + rootURL + "writeLayer:" + rootURL + "busybox"
	cmd := exec.Command("mount", "-t", "aufs", "-o", dirs, "none", mntURL)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		logrus.Errorf("%v", err)
	}
}

func DeleteWorkSpace(rootURL string, mntURL string, volume string) {

	if volume != "" {
		volumeURLs := volumeUrlExtract(volume)
		length := len(volumeURLs)
		if length == 2 && volumeURLs[0] != "" && volumeURLs[1] != "" {
			DeleteMountPointWithVolume(rootURL, mntURL, volumeURLs)
		} else {
			DeleteMountPoint(mntURL)
		}

	} else {
		DeleteMountPoint(mntURL)
	}
	DeleteWriteLayer(rootURL)
}

func DeleteMountPointWithVolume(rootURL, mntURL string, volumeURLs []string) {
	contanerUrl := mntURL + volumeURLs[1]
	cmd := exec.Command("umount", contanerUrl)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		logrus.Errorln("unmount volume failed")
	}

	cmd = exec.Command("umount", mntURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		logrus.Errorln("umount mountpoint failed")
	}

	if err := os.RemoveAll(mntURL); err != nil {
		logrus.Infof("remove mountpoint dir %s err %v", mntURL, err)
	}

}

func DeleteMountPoint(mntURL string) {
	cmd := exec.Command("umount", mntURL)
	fmt.Println("dao zhe le ma ~!!!!")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	fmt.Println("dao zhe le ma ~!!!!2")
	if err := cmd.Run(); err != nil {
		fmt.Println("dao zhe le ma ~!!!!3")
		logrus.Errorln("FFFFFFFFFFFFFFFFFFFFF", err)
	}
	fmt.Println("dao zhe le ma ~!!!!4")
	if err := os.Remove(mntURL); err != nil {
		fmt.Println("dao zhe le ma ~!!!!5")
		logrus.Errorln("RRRRRRRRRRRRR", err)
	}
}

func DeleteWriteLayer(rootURL string) {
	writeURL := rootURL + "writeLayer/"
	if err := os.RemoveAll(writeURL); err != nil {
		logrus.Errorln(err)
	}
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return false, err
}
