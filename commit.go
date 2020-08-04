package main

import (
	"fmt"
	"os/exec"

	"github.com/sirupsen/logrus"
)

func commitContainer(imageName string) {
	mntURL := "/root/mnt"
	imageTar := "/root/" + imageName + ".tar"
	fmt.Printf("%s", imageTar)
	if _, err := exec.Command("tar", "-cvf", imageTar, "-C", mntURL, ".").CombinedOutput(); err != nil {
		logrus.Errorf("tar folder %s err %v", mntURL, err)
	}
}
