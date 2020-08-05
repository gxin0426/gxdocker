package main

import (
	"fmt"
	"gxdocker/container"
	"os/exec"

	"github.com/sirupsen/logrus"
)

func commitContainer(containerName, imageName string) {
	mntURL := fmt.Sprintf(container.MntUrl, containerName)
	imageTar := container.RootUrl + "/" + imageName + ".tar"
	fmt.Printf("%s", imageTar)
	if _, err := exec.Command("tar", "-cvf", imageTar, "-C", mntURL, ".").CombinedOutput(); err != nil {
		logrus.Errorf("tar folder %s err %v", mntURL, err)
	}
}
