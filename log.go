package main

import (
	"fmt"
	"gxdocker/container"
	"io/ioutil"
	"os"

	"github.com/sirupsen/logrus"
)

func logContainer(containerName string) {
	dirURL := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	logFileLocation := dirURL + container.ContainerLogFile
	file, err := os.Open(logFileLocation)
	defer file.Close()
	if err != nil {
		logrus.Errorf("log container open file error")
		return
	}

	content, err := ioutil.ReadAll(file)
	if err != nil {
		logrus.Errorf("log container read file error")
		return
	}

	fmt.Fprintf(os.Stdout, string(content))
}
