package main

import (
	"encoding/json"
	"fmt"
	"gxdocker/cgroups"
	"gxdocker/cgroups/subsystems"
	"gxdocker/container"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

func Run(tty bool, comArray []string, res *subsystems.ResourceConfig, volume string, containerName string) {
	parent, writePipe := container.NewParentProcess(tty, volume, containerName)
	if parent == nil {
		logrus.Errorf("new parnet process error")
		return
	}
	if err := parent.Start(); err != nil {
		logrus.Error(err)
		fmt.Println("tttttttttttttttttttttttttttttttttttttttttttt")
	}
	fmt.Println("dfdffffffffffffffffffffffff", parent.Process.Pid, comArray, containerName)
	containerName, err := recordContainerInfo(parent.Process.Pid, comArray, containerName)
	if err != nil {
		logrus.Errorf("Record container info error %v", err)
		return
	}

	cgroupManager := cgroups.NewCgroupManager("mydocker-cgroup")
	defer cgroupManager.Destory()
	cgroupManager.Set(res)
	cgroupManager.Apply(parent.Process.Pid)
	fmt.Println("ggggggggggggggggggggggggggg", comArray)
	sendInitCommand(comArray, writePipe)

	//if tty {
	parent.Wait()
	deleteContainerInfo(containerName)
	// }

	mntURL := "/root/mnt/"
	rootURL := "/root/"

	fmt.Println("delete mnt   !!!")
	container.DeleteWorkSpace(rootURL, mntURL, volume)

	os.Exit(0)
}

func sendInitCommand(comArray []string, writePipe *os.File) {
	command := strings.Join(comArray, " ")
	logrus.Infof("command all is %v", command)
	writePipe.WriteString(command)
	writePipe.Close()
}

func deleteContainerInfo(containerId string) {
	dirURl := fmt.Sprintf(container.DefaultInfoLocation, containerId)
	if err := os.RemoveAll(dirURl); err != nil {
		logrus.Errorf("remove dir error")
	}
}

func recordContainerInfo(containerPID int, commandArray []string, containerName string) (string, error) {
	id := randStringBytes(10)
	createTime := time.Now().Format("2006-01-02 15:04:05")
	command := strings.Join(commandArray, "")

	if containerName == "" {
		containerName = id
	}

	containerInfo := &container.ContainerInfo{
		Id:          id,
		Pid:         strconv.Itoa(containerPID),
		CreatedTime: createTime,
		Command:     command,
		Status:      container.RUNNING,
		Name:        containerName,
	}

	jsonBytes, err := json.Marshal(containerInfo)
	if err != nil {
		logrus.Errorf("record container info error %v", err)
		return "", err
	}

	jsonStr := string(jsonBytes)

	dirUrl := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	fmt.Println("dirUrl : ", dirUrl)

	if err := os.MkdirAll(dirUrl, 0622); err != nil {
		logrus.Errorf("mkdir err %s err %v", dirUrl, err)
		return "", err
	}
	fileName := dirUrl + "/" + container.ConfigName

	file, err := os.Create(fileName)

	defer file.Close()

	if err != nil {
		logrus.Errorf("create file error")
		return "", err
	}

	if _, err := file.WriteString(jsonStr); err != nil {
		logrus.Errorf("file write string error")
		return "", err
	}

	return containerName, nil

}

func randStringBytes(n int) string {
	lettersBytes := "123456789"
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = lettersBytes[rand.Intn(len(lettersBytes))]
	}

	return string(b)
}
