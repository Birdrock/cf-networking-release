package cf_cli_adapter

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type Adapter struct {
	CfCliPath string
}

func (a *Adapter) CreateOrg(name string) error {
	fmt.Printf("running: %s create-org %s\n", a.CfCliPath, name)
	cmd := exec.Command(a.CfCliPath, "create-org", name)
	return runCommandWithTimeout(cmd)
}

func (a *Adapter) DeleteOrg(name string) error {
	fmt.Printf("running: %s delete-org -f %s\n", a.CfCliPath, name)
	cmd := exec.Command(a.CfCliPath, "delete-org", "-f", name)
	return runCommandWithTimeout(cmd)
}

func (a *Adapter) CreateSpace(spaceName, orgName string) error {
	fmt.Printf("running: %s create-space %s -o %s\n", a.CfCliPath, spaceName, orgName)
	cmd := exec.Command(a.CfCliPath, "create-space", spaceName, "-o", orgName)
	return runCommandWithTimeout(cmd)
}

func (a *Adapter) TargetOrg(name string) error {
	fmt.Printf("running: %s target -o  %s\n", a.CfCliPath, name)
	cmd := exec.Command(a.CfCliPath, "target", "-o", name)
	return runCommandWithTimeout(cmd)
}

func (a *Adapter) TargetSpace(name string) error {
	fmt.Printf("running: %s target -s  %s\n", a.CfCliPath, name)
	cmd := exec.Command(a.CfCliPath, "target", "-s", name)
	return runCommandWithTimeout(cmd)
}

func (a *Adapter) SetApiWithSsl(api string) error {
	fmt.Printf("running: %s api  %s\n", a.CfCliPath, api)
	cmd := exec.Command(a.CfCliPath, "api", api)
	return runCommandWithTimeout(cmd)
}

func (a *Adapter) SetApiWithoutSsl(api string) error {
	fmt.Printf("running: %s api  %s --skip-ssl-validation\n", a.CfCliPath, api)
	cmd := exec.Command(a.CfCliPath, "api", api, "--skip-ssl-validation")
	return runCommandWithTimeout(cmd)
}

func (a *Adapter) Auth(user, password string) error {
	fmt.Printf("running: %s auth <user> <pass> \n", a.CfCliPath)
	cmd := exec.Command(a.CfCliPath, "auth", user, password)
	return runCommandWithTimeout(cmd)
}

func (a *Adapter) Push(name, directory, manifestFile string) error {
	fmt.Printf("running: %s push %s -p %s -f %s\n", a.CfCliPath, name, directory, manifestFile)
	bytes, err := exec.Command(a.CfCliPath,
		"push", name,
		"-p", directory,
		"-f", manifestFile).CombinedOutput()
	fmt.Printf("output: %s\n", string(bytes))
	return err
}

func (a *Adapter) Delete(appName string) error {
	fmt.Printf("running: %s delete -f %s\n", a.CfCliPath, appName)
	cmd := exec.Command(a.CfCliPath, "delete", "-f", appName)
	return runCommandWithTimeout(cmd)
}

func (a *Adapter) Scale(name string, instances int) error {
	instancesStr := fmt.Sprintf("%d", instances)
	fmt.Printf("running: %s scale %s -i %s\n", a.CfCliPath, name, instancesStr)
	cmd := exec.Command(a.CfCliPath, "scale", name, "-i", instancesStr)
	return runCommandWithTimeout(cmd)
}

func (a *Adapter) AppGuid(name string) (string, error) {
	fmt.Printf("running: %s app %s --guid\n", a.CfCliPath, name)
	bytes, err := exec.Command(a.CfCliPath, "app", name, "--guid").CombinedOutput()
	return strings.TrimSpace(string(bytes)), err
}

func (a *Adapter) SpaceGuid(name string) (string, error) {
	fmt.Printf("running: %s space %s --guid\n", a.CfCliPath, name)
	bytes, err := exec.Command(a.CfCliPath, "space", name, "--guid").CombinedOutput()
	return strings.TrimSpace(string(bytes)), err
}

type Apps struct {
	TotalResults int `json:"total_results"`
}

func (a *Adapter) OrgGuid(name string) (string, error) {
	fmt.Printf("running: %s org %s --guid\n", a.CfCliPath, name)
	bytes, err := exec.Command(a.CfCliPath, "org", name, "--guid").CombinedOutput()
	return strings.TrimSpace(string(bytes)), err
}

func (a *Adapter) Curl(method, path, inputFile string) ([]byte, error) {
	if inputFile != "" {
		fmt.Println("running:", a.CfCliPath, "curl", "-X", method, "-d", fmt.Sprintf("@%s", inputFile), path)
		return exec.Command(a.CfCliPath, "curl", "-X", method, "-d", fmt.Sprintf("@%s", inputFile), path).CombinedOutput()
	}

	fmt.Printf("running: %s curl -X %s \"%s\"\n", a.CfCliPath, method, path)
	return exec.Command(a.CfCliPath, "curl", "-X", method, path).CombinedOutput()
}

func (a *Adapter) AppCount(orgGuid string) (int, error) {
	fmt.Printf("running: %s curl \"/v2/apps?q=organization_guid%%20IN%%20%s\"\n", a.CfCliPath, orgGuid)
	bytes, err := exec.Command(a.CfCliPath, "curl", fmt.Sprintf("/v2/apps?q=organization_guid%%20IN%%20%s", orgGuid)).CombinedOutput()
	apps := &Apps{}
	if err := json.Unmarshal(bytes, apps); err != nil {
		return -1, err
	}
	return apps.TotalResults, err
}

func (a *Adapter) CheckApp(guid string) ([]byte, error) {
	fmt.Printf("running: %s curl \"/v2/apps/%s/summary\"\n", a.CfCliPath, guid)
	bytes, err := exec.Command(a.CfCliPath, "curl", fmt.Sprintf("/v2/apps/%s/summary", guid)).CombinedOutput()
	return bytes, err
}

func (a *Adapter) AddNetworkPolicy(sourceApp, destApp string, port int, protocol string) error {
	portStr := fmt.Sprintf("%d-%d", port, port)
	commandArgs := []string{"add-network-policy", sourceApp, destApp, "--port", portStr, "--protocol", "tcp"}
	fmt.Printf("running: cf %v \n", commandArgs)
	cmd := exec.Command("cf", commandArgs...)
	return runCommandWithTimeout(cmd)
}

func (a *Adapter) RemoveNetworkPolicy(sourceApp, destApp string, port int, protocol string) error {
	portStr := fmt.Sprintf("%d-%d", port, port)
	commandArgs := []string{"remove-network-policy", sourceApp, destApp, "--port", portStr, "--protocol", "tcp"}
	fmt.Printf("running: cf %v \n", commandArgs)
	cmd := exec.Command("cf", commandArgs...)
	return runCommandWithTimeout(cmd)
}

func (a *Adapter) CreateQuota(name, memory string, instanceMemory, routes, serviceInstances, appInstances, routePorts int) error {
	instanceMemoryStr := fmt.Sprintf("%d", instanceMemory)
	routesStr := fmt.Sprintf("%d", routes)
	serviceInstancesStr := fmt.Sprintf("%d", serviceInstances)
	appInstancesStr := fmt.Sprintf("%d", appInstances)
	routePortsStr := fmt.Sprintf("%d", routePorts)
	fmt.Printf("running cf create-org-quota %s -m %s -i %s -r %s -s %s -a %s --reserved-route-ports %s\n", name, memory, instanceMemoryStr, routesStr, serviceInstancesStr, appInstancesStr, routePortsStr)
	cmd := exec.Command("cf", "create-org-quota", name, "-m", memory, "-i", instanceMemoryStr, "-r", routesStr, "-s", serviceInstancesStr, "-a", appInstancesStr, "--reserved-route-ports", routePortsStr)
	return runCommandWithTimeout(cmd)
}

func (a *Adapter) SetQuota(org, quota string) error {
	fmt.Printf("running cf set-org-quota %s %s\n", org, quota)
	cmd := exec.Command("cf", "set-org-quota", org, quota)
	return runCommandWithTimeout(cmd)
}

func (a *Adapter) CreateSecurityGroup(name, filepath string) error {
	fmt.Printf("running cf create-security-group %s %s\n", name, filepath)
	cmd := exec.Command("cf", "create-security-group", name, filepath)
	return runCommandWithTimeout(cmd)
}

type ASG struct {
	Resources []struct {
		Entity struct {
			Rules []struct {
				Destination string `json:"destination"`
				Ports       string `json:"ports"`
				Protocol    string `json:"protocol"`
			} `json:"rules"`
		} `json:"entity"`
	} `json:"resources"`
}

func (a *Adapter) SecurityGroup(name string) (string, error) {
	fmt.Printf("running: %s curl \"/v2/security_groups?q=name%%3A%s\n", a.CfCliPath, name)
	bytes, err := exec.Command(a.CfCliPath, "curl", fmt.Sprintf("/v2/security_groups?q=name%%3A%s", name)).CombinedOutput()
	asg := &ASG{}
	if err := json.Unmarshal(bytes, asg); err != nil {
		return "", err
	}
	if len(asg.Resources) == 0 {
		return "", errors.New("no asgs with the name " + name)
	}
	rules, err := json.Marshal(asg.Resources[0].Entity.Rules)
	if err != nil {
		return "", err
	}
	return string(rules), err
}

func (a *Adapter) BindSecurityGroup(name, org, space string) error {
	fmt.Printf("running cf bind-security-group %s %s --space %s\n", name, org, space)
	cmd := exec.Command("cf", "bind-security-group", name, org, "--space", space)
	return runCommandWithTimeout(cmd)
}

func (a *Adapter) UnbindSecurityGroup(name, org, space string) error {
	fmt.Printf("running cf unbind-security-group %s %s %s\n", name, org, space)
	cmd := exec.Command("cf", "unbind-security-group", name, org, space)
	return runCommandWithTimeout(cmd)
}

func (a *Adapter) DeleteSecurityGroup(name string) error {
	fmt.Printf("running cf delete-security-group -f %s \n", name)
	cmd := exec.Command("cf", "delete-security-group", "-f", name)
	return runCommandWithTimeout(cmd)
}

func (a *Adapter) DeleteQuota(quota string) error {
	fmt.Printf("running cf delete-org-quota %s -f\n", quota)
	cmd := exec.Command("cf", "delete-org-quota", quota, "-f")
	return runCommandWithTimeout(cmd)
}

type CmdErr struct {
	Out     string
	Err     string
	Message string
}

func (e *CmdErr) Error() string {
	return fmt.Sprintf("%s:\n\nOut:\n%s\n\nErr:%s\n", e.Message, e.Out, e.Err)
}

func runCommandWithTimeout(cmd *exec.Cmd) error {
	outBuffer := &bytes.Buffer{}
	errBuffer := &bytes.Buffer{}
	wrapErr := func(msg string) error {
		return &CmdErr{
			Out:     outBuffer.String(),
			Err:     errBuffer.String(),
			Message: msg,
		}
	}
	cmd.Stdout = outBuffer
	cmd.Stderr = errBuffer
	if err := cmd.Start(); err != nil {
		return err
	}
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-time.After(2 * time.Minute):
		if err := cmd.Process.Kill(); err != nil {
			return wrapErr(fmt.Sprintf("command timed out and could not be killed: %s", err))
		}
		return wrapErr("command timed out")

	case err := <-done:
		if err != nil {
			return wrapErr(err.Error())
		}
	}
	return nil
}
