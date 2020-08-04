package subsystems

import (
	"strconv"
	"os"
	"fmt"
	"path"
	"io/ioutil"
)


type MemorySubSystem struct{

}

func (m *MemorySubSystem) Set(cgroupPath string, res *ResourceConfig)error{
	if subsysCgroupPath, err := GetCgroupPath(m.Name(), cgroupPath, true); err == nil {
		if res.MemoryLimit != ""{
			if err := ioutil.WriteFile(path.Join(subsysCgroupPath, "memory.limit_in_bytes"), []byte(res.MemoryLimit), 0644); err != nil {
				return fmt.Errorf("set cgroup memory fail %v", err)
			}
		}
		return nil
	}else{
		return err
	}
}

func (m *MemorySubSystem) Name() string{
	return "memory"
}


func (m *MemorySubSystem)Remove(cgroupPath string) error{
	if subsysCgroupPath, err := GetCgroupPath(m.Name(), cgroupPath, false); err == nil {
		return os.RemoveAll(subsysCgroupPath)
	}else{
		return err
	}

}

func (m *MemorySubSystem) Apply(cgroupPath string, pid int) error{
	if subsysCgroupPath, err := GetCgroupPath(m.Name(), cgroupPath, false); err == nil {
		if err := ioutil.WriteFile(path.Join(subsysCgroupPath, "tasks"), []byte(strconv.Itoa(pid)), 0644); err != nil {
			return fmt.Errorf("set proc fail %v", err)
		}
		return nil
	}else{
		return fmt.Errorf("get cgroup %s err : %v", cgroupPath, err)
	}
}