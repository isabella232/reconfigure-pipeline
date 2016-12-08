package lastpass

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"os/exec"
	"strings"

	"regexp"

	"gopkg.in/yaml.v2"
)

type Processor interface {
	Process(config string) string
}

type processor struct {
}

func NewProcessor() Processor {
	return &processor{}
}

func (l *processor) Process(config string) string {
	re := regexp.MustCompile("lpass:///(.*)")

	processedConfig := re.ReplaceAllStringFunc(config, func(match string) string {
		credHandle, _ := url.Parse(match)
		return handle(credHandle)
	})

	return processedConfig
}

func handle(credHandle *url.URL) string {
	pathParts := strings.Split(credHandle.Path, "/")

	credential := getCredential(pathParts[1], pathParts[2])

	if credHandle.Fragment != "" {
		// Assume YAML contents, return element
		fragmentMap := map[string]string{}
		err := yaml.Unmarshal([]byte(credential), &fragmentMap)
		if err != nil {
			if err != nil {
				log.Fatal(err)
			}
		}
		credential = fragmentMap[credHandle.Fragment]
	}

	if strings.Contains(credential, "\n") {
		encoded, _ := json.Marshal(credential) // always a string
		return string(encoded)
	}

	return credential
}

func getCredential(credential, field string) string {
	fieldFlagMap := map[string]string{
		"Password": "--password",
		"Username": "--username",
		"URL":      "--url",
		"Notes":    "--notes",
	}

	fieldFlag := fieldFlagMap[field]
	if fieldFlag == "" {
		fieldFlag = "--field=" + field
	}

	cmd := exec.Command("lpass", "show", fieldFlag, credential)

	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	err = cmd.Start()
	if err != nil {
		log.Fatal(err)
	}

	output, err := ioutil.ReadAll(stdout)

	err = cmd.Wait()
	if err != nil {
		log.Fatal(err)
	}

	return string(output)
}
