package main

import (
	"os"
	"fmt"
	"gxdocker/container"

	"github.com/sirupsen/logrus"
)

func removeContainer(containerName string) {
	containerInfo, err := getContainerInfoByName(containerName)
	if err != nil {
		logrus.Errorf("get container %s info error %v", containerName, err)
		return
	}

	if containerInfo.Status != container.STOP {
		logrus.Errorf("Could not remove running container")
		return
	}
	dirURL := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	if err := os.RemoveAll(dirURL); err != nil {
		logrus.Errorf("remove file %s error %v", dirURL, err)
		return
	}
}
