package backend

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// APIServer is an instance of the API service
type APIServer struct {
	dirPath     string
	binPath     string
	configsPath string
}

// MakeAPIServer is a factory method
func MakeAPIServer() (*APIServer, error) {
	binPath, e := filepath.Abs(os.Args[0])
	if e != nil {
		return nil, fmt.Errorf("could not get binPath of currently running binary: %s", e)
	}

	dirPath := filepath.Dir(binPath)
	configsPath := dirPath + "/ops/configs"

	return &APIServer{
		dirPath:     dirPath,
		binPath:     binPath,
		configsPath: configsPath,
	}, nil
}

func (s *APIServer) runKelpCommand(cmd string) ([]byte, error) {
	cmdString := fmt.Sprintf("%s %s", s.binPath, cmd)
	return runBashCommand(cmdString)
}

func runBashCommand(cmd string) ([]byte, error) {
	bytes, e := exec.Command("bash", "-c", cmd).Output()
	if e != nil {
		return nil, fmt.Errorf("could not run bash command '%s': %s", cmd, e)
	}
	return bytes, nil
}