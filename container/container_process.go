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
	RootUrl             string = "/root"
	MntUrl              string = "/root/mnt/%s/"
	WriteLayerUrl       string = "/root/writeLayer/%s/"
)

type ContainerInfo struct {
	Pid         string `json:"pid"`
	Id          string `json:"id"`
	Name        string `json:"name"`
	Command     string `json:"command"`
	CreatedTime string `json:"createTime"`
	Status      string `json:"status"`
	Volume      string `json:"volume"`
}

func NewParentProcess(tty bool, volume string, containerName string, imageName string) (*exec.Cmd, *os.File) {
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
	// mntURL := "/root/mnt"
	// rootURL := "/root/"
	NewWorkSpace(volume, imageName, containerName)
	cmd.Dir = fmt.Sprintf(MntUrl, containerName)
	return cmd, writePipe
}

func NewPipe() (*os.File, *os.File, error) {
	read, write, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}
	return read, write, nil
}

func NewWorkSpace(volume, imageName, containerName string) {
	CreateReadOnlyLayer(imageName)
	CreateWriteLayer(containerName)
	CreateMountPoint(containerName, imageName)

	if volume != "" {
		volumeURLs := volumeUrlExtract(volume)
		length := len(volumeURLs)
		if length == 2 && volumeURLs[0] != "" && volumeURLs[1] != "" {
			MountVolume(containerName, volumeURLs)
			logrus.Infof("%q", volumeURLs)
		} else {
			logrus.Infof("Volume paramter input is not correct")
		}
	}
}

func MountVolume(containerName string, volumesURLs []string) {
	parentUrl := volumesURLs[0]
	if err := os.MkdirAll(parentUrl, 0777); err != nil {
		logrus.Infof("mkdir parent dir %s error %v", parentUrl, err)
	}

	containerUrl := volumesURLs[1]
	fmt.Println("containerUrl", containerUrl)
	mntURL := fmt.Sprintf(MntUrl, containerName)
	fmt.Println("mntURL", mntURL)
	containerVolumeURL := mntURL + "/" + containerUrl
	fmt.Println("containerVolumeURL", containerVolumeURL)
	if err := os.MkdirAll(containerVolumeURL, 0777); err != nil {
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

func CreateReadOnlyLayer(imageName string) error {

	unTarFolderUrl := RootUrl + "/" + imageName + "/"
	fmt.Println("unTarFolderUrl", unTarFolderUrl)
	imageUrl := RootUrl + "/" + imageName + ".tar"
	fmt.Println(imageUrl)
	exists, err := PathExists(unTarFolderUrl)
	if err != nil {
		panic("fail to judge whether dir exist")
	}

	if exists == false {
		if err := os.MkdirAll(unTarFolderUrl, 0777); err != nil {
			panic("mkdir dir err")
		}
		if _, err := exec.Command("tar", "-xvf", imageUrl, "-C", unTarFolderUrl).CombinedOutput(); err != nil {
			panic("tar -xvf fail")
		}
	}
	return nil
}

func CreateWriteLayer(containerName string) {
	writeURL := fmt.Sprintf(WriteLayerUrl, containerName)
	if err := os.MkdirAll(writeURL, 0777); err != nil {
		fmt.Println(err)
		return
	}

}

func CreateMountPoint(containerName, imageName string) error {

	mntUrl := fmt.Sprintf(MntUrl, containerName)

	if err := os.MkdirAll(mntUrl, 0777); err != nil {
		fmt.Println("MkdirAll(mntUrl, 0777)", err)
		return err
	}

	tmpWriteLayer := fmt.Sprintf(WriteLayerUrl, containerName)
	tmpImageLocation := RootUrl + "/" + imageName

	dirs := "dirs=" + tmpWriteLayer + ":" + tmpImageLocation
	cmd := exec.Command("mount", "-t", "aufs", "-o", dirs, "none", mntUrl)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		logrus.Errorf("%v", err)
		return err
	}
	return nil
}

func DeleteWorkSpace(containerName, volume string) {

	if volume != "" {
		volumeURLs := volumeUrlExtract(volume)
		length := len(volumeURLs)
		if length == 2 && volumeURLs[0] != "" && volumeURLs[1] != "" {
			DeleteMountPointWithVolume(containerName, volumeURLs)
		} else {
			DeleteMountPoint(containerName)
		}

	} else {
		DeleteMountPoint(containerName)
	}
	DeleteWriteLayer(containerName)
}

func DeleteMountPointWithVolume(containerName string, volumeURLs []string) error {

	mntURL := fmt.Sprintf(MntUrl, containerName)
	containerUrl := mntURL + "/" + volumeURLs[1]
	if _, err := exec.Command("umount", containerUrl).CombinedOutput(); err != nil {
		logrus.Errorf("Umount volume fail")
		return err
	}
	if _, err := exec.Command("umount", mntURL).CombinedOutput(); err != nil {
		logrus.Errorf("umount mntUrl fail")
		return err
	}

	if err := os.RemoveAll(mntURL); err != nil {
		logrus.Infof("remove mountpoint dir %s err %v", mntURL, err)
		return err
	}
	return nil
}

func DeleteMountPoint(containerName string) error {

	mntURL := fmt.Sprintf(MntUrl, containerName)
	_, err := exec.Command("umount", mntURL).CombinedOutput()
	if err != nil {
		logrus.Errorf("Umount error %v", err)
		return err
	}
	if err := os.Remove(mntURL); err != nil {
		logrus.Errorln("RRRRRRRRRRRRR", err)
		return err
	}
	return nil
}

func DeleteWriteLayer(containerName string) {
	writeURL := fmt.Sprintf(WriteLayerUrl, containerName)
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
