package main

import (
	"encoding/json"
	"fmt"
	"gxdocker/container"
	"io/ioutil"
	"strconv"
	"syscall"

	"github.com/sirupsen/logrus"
)

func getContainerInfoByName(containerName string) (*container.ContainerInfo, error) {
	dirURL := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	configFilePath := dirURL + container.ConfigName
	contentBytes, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		logrus.Errorf("read file err %v", err)
		return nil, err
	}
	var containerInfo container.ContainerInfo
	if err := json.Unmarshal(contentBytes, &containerInfo); err != nil {
		logrus.Errorf("getContainerInfoByName err")
		return nil, err
	}
	return &containerInfo, nil
}

func stopContainer(containerName string) {
	pid, err := getContainerPidByName(containerName)
	if err != nil {
		logrus.Errorf("get container pid by name")
		return
	}

	pidInt, err := strconv.Atoi(pid)
	if err != nil {
		logrus.Errorf("conver pid to int errr")
		return
	}

	if err := syscall.Kill(pidInt, syscall.SIGTERM); err != nil {
		logrus.Errorf("stop container error %v", err)
		return
	}

	containerInfo, err := getContainerInfoByName(containerName)
	if err != nil {
		logrus.Errorf("get getcontainerinfoByname error %v", err)
		return
	}

	containerInfo.Status = container.STOP

	containerInfo.Pid = " "
	newContentBytes, err := json.Marshal(containerInfo)
	if err != nil {
		logrus.Errorf("json marshal %s error %v", containerName, err)
		return
	}
	dirURL := fmt.Sprintf(container.DefaultInfoLocation, containerName)

	configFilePath := dirURL + container.ConfigName

	if err := ioutil.WriteFile(configFilePath, newContentBytes, 0622); err != nil {
		logrus.Errorf("write file %s error %v", configFilePath, err)
	}

}
