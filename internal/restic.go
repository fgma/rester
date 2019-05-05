package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	shlex "github.com/anmitsu/go-shlex"
	homedir "github.com/mitchellh/go-homedir"
)

type Restic struct {
	resticExecutable string
}

func NewRestic(resticExecutable string) Restic {
	r := Restic{
		resticExecutable: resticExecutable,
	}
	return r
}

func (r Restic) IsResticAvailable() bool {

	cmd := r.resticExecutable
	args := []string{"version"}

	_, err := exec.Command(cmd, args...).Output()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return false
	}

	return true
}

func (r Restic) IsRepositoryAvailable(repository Repository) error {

	cmd := r.prepareResticCommand(repository, make(map[string]string))
	cmd.Args = append(cmd.Args, "snapshots")

	return cmd.Run()
}

func (r Restic) RunBackup(backup Backup, repository Repository) error {

	environment := combineMaps(repository.Environment, backup.Environment)

	if err := r.IsRepositoryAvailable(repository); err != nil {
		runHandlerBackupFailure(backup, repository, environment)
		return err
	}

	runHandler(backup.Handler.Before, "before", environment, &backup, &repository)

	if err := r.runUnlock(repository); err != nil {
		r.dumpUnlockError(repository, err)
		runHandlerBackupFailure(backup, repository, environment)
		return err
	}

	cmd := r.prepareResticCommand(repository, backup.Environment)
	cmd.Args = append(cmd.Args, "backup")

	cmd.Args = append(cmd.Args, backup.Data...)

	for _, exclude := range backup.Exclude {
		cmd.Args = append(cmd.Args, fmt.Sprintf("--exclude=%s", exclude))
	}

	for _, tag := range backup.Tags {
		cmd.Args = append(cmd.Args, "--tag", tag)
	}

	if backup.OneFileSystem {
		cmd.Args = append(cmd.Args, "--one-file-system")
	}

	for _, flag := range backup.CustomFlags {
		cmd.Args = append(cmd.Args, flag)
	}

	var cmdStdin *exec.Cmd
	var pr, pw *os.File

	if backup.DataStdinCommand != "" {

		cmd.Args = append(cmd.Args, "--stdin", "--stdin-filename", backup.StdinFilename)

		var err error

		cmdStdin, err = prepareShellCommand(backup.DataStdinCommand, environment)
		if err != nil {
			fmt.Fprintf(
				os.Stderr, "Failed to prepare stdin shell command \"%s\": %s\n",
				backup.DataStdinCommand, err,
			)
			return err
		}

		pr, pw, err = os.Pipe()

		if err != nil {
			fmt.Fprintf(
				os.Stderr, "Failed to create pipe: %s\n",
				err,
			)
			return err
		}

		cmdStdin.Stdout = pw
		cmdStdin.Stderr = os.Stderr

		cmd.Stdin = pr
	}

	/*
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Args = append(cmd.Args, "--verbose=5")

		fmt.Println(cmd)
	*/

	if cmdStdin != nil {
		err := cmdStdin.Start()

		if err != nil {
			fmt.Fprintf(
				os.Stderr, "Failed to run stdin command: %s\n",
				err,
			)
			return err
		}
	}

	err := cmd.Start()
	if err != nil {
		fmt.Fprintf(
			os.Stderr, "Failed to run restic command: %s\n",
			err,
		)
		return err
	}

	if cmdStdin != nil {
		err = cmdStdin.Wait()
		pw.Close()

		if err != nil {
			fmt.Fprintf(
				os.Stderr, "Failed to wait for stdin command: %s\n",
				err,
			)
			return err
		}
	}

	err = cmd.Wait()
	pr.Close()

	if err != nil {
		fmt.Fprintf(
			os.Stderr, "Failed to wait for restic command: %s\n",
			err,
		)
		return err
	}

	runHandler(backup.Handler.After, "after", environment, &backup, &repository)

	if err != nil {
		runHandlerBackupFailure(backup, repository, environment)
	} else {
		runHandler(backup.Handler.Success, "success", environment, &backup, &repository)
	}

	return err
}

func (r Restic) RunCheck(repository Repository) error {

	if err := r.runUnlock(repository); err != nil {
		r.dumpUnlockError(repository, err)
		runHandlerCheckFailure(repository)
		return err
	}

	cmd := r.prepareResticCommand(repository, make(map[string]string))
	cmd.Args = append(cmd.Args, "check")

	if repository.Check.ReadDataPercentage >= 100 {
		cmd.Args = append(cmd.Args, "--read-data")
	} else if repository.Check.ReadDataPercentage > 0 {
		subsets := int(math.Ceil(100.0 / float64(repository.Check.ReadDataPercentage)))
		rand.Seed(time.Now().UTC().UnixNano())
		subsetToCheck := rand.Intn(subsets) + 1

		cmd.Args = append(cmd.Args, fmt.Sprintf("--read-data-subset=%d/%d", subsetToCheck, subsets))
	}

	err := cmd.Run()
	if err != nil {
		fmt.Fprintf(
			os.Stderr, "Failed to check repository [%s]: %s\n",
			repository.Name,
			err,
		)
		runHandlerCheckFailure(repository)
		return err
	} else {
		runHandler(repository.Handler.CheckSuccess, "check_success", repository.Environment, nil, &repository)
	}

	return nil
}

func (r Restic) RunForget(repository Repository) error {

	if err := r.runUnlock(repository); err != nil {
		r.dumpUnlockError(repository, err)
		runHandlerForgetFailure(repository)
		return err
	}

	cmd := r.prepareResticCommand(repository, make(map[string]string))
	cmd.Args = append(cmd.Args, "forget", "--prune")

	if repository.Policy.KeepLast > 0 {
		cmd.Args = append(cmd.Args, "--keep-last", fmt.Sprint(repository.Policy.KeepLast))
	}

	if repository.Policy.KeepHourly > 0 {
		cmd.Args = append(cmd.Args, "--keep-hourly", fmt.Sprint(repository.Policy.KeepHourly))
	}

	if repository.Policy.KeepDaily > 0 {
		cmd.Args = append(cmd.Args, "--keep-daily", fmt.Sprint(repository.Policy.KeepDaily))
	}

	if repository.Policy.KeepWeekly > 0 {
		cmd.Args = append(cmd.Args, "--keep-weekly", fmt.Sprint(repository.Policy.KeepWeekly))
	}

	if repository.Policy.KeepMonthly > 0 {
		cmd.Args = append(cmd.Args, "--keep-monthly", fmt.Sprint(repository.Policy.KeepMonthly))
	}

	if repository.Policy.KeepYearly > 0 {
		cmd.Args = append(cmd.Args, "--keep-yearly", fmt.Sprint(repository.Policy.KeepYearly))
	}

	if repository.Policy.KeepWithin != "" {
		cmd.Args = append(cmd.Args, "--keep-within", fmt.Sprint(repository.Policy.KeepWithin))
	}

	for _, tag := range repository.Policy.KeepTags {
		cmd.Args = append(cmd.Args, "--keep-tag", tag)
	}

	err := cmd.Run()
	if err != nil {
		fmt.Fprintf(
			os.Stderr, "Failed to forget for repository [%s]: %s\n",
			repository.Name,
			err,
		)
		runHandlerForgetFailure(repository)
		return err
	} else {
		runHandler(repository.Handler.ForgetSuccess, "forget_success", repository.Environment, nil, &repository)
	}

	return nil
}

func (r Restic) PrintSnapshots(repository Repository) error {

	if err := r.runUnlock(repository); err != nil {
		r.dumpUnlockError(repository, err)
		return err
	}

	cmd := r.prepareResticCommand(repository, make(map[string]string))
	cmd.Args = append(cmd.Args, "snapshots")

	cmdOut, err := cmd.Output()
	if err != nil {
		return err
	}

	fmt.Printf("Snapshots for %s (%s):\n\n", repository.Name, repository.URL)
	fmt.Print(string(cmdOut))

	return err
}

func (r Restic) Mount(repository Repository, mountPoint string) error {

	if fileInfo, err := os.Stat(mountPoint); err != nil || !fileInfo.IsDir() {
		return fmt.Errorf("%s is not a directory", mountPoint)
	}

	if err := r.runUnlock(repository); err != nil {
		r.dumpUnlockError(repository, err)
		return err
	}

	cmd := r.prepareResticCommand(repository, make(map[string]string))
	cmd.Args = append(cmd.Args, "mount", mountPoint)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func (r Restic) Init(repository Repository) error {

	cmd := r.prepareResticCommand(repository, make(map[string]string))
	cmd.Args = append(cmd.Args, "init")

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func (r Restic) CheckAge(backup Backup, repository Repository) (bool, bool, error) {

	environment := combineMaps(repository.Environment, backup.Environment)

	if err := r.runUnlock(repository); err != nil {
		r.dumpUnlockError(repository, err)
		runHandlerAgeError(backup, repository, environment)
		return false, false, err
	}

	lastBackupTimestamp, err := r.GetLastBackupTimestamp(backup, repository)

	if err != nil {
		return false, false, err
	}

	if (lastBackupTimestamp == time.Time{}) {
		return false, true, nil
	} else {
		age := time.Now().Sub(lastBackupTimestamp)

		if age > backup.Age.Error.Duration {
			runHandlerAgeError(backup, repository, environment)
			return false, true, nil
		} else if age > backup.Age.Warn.Duration {
			runHandler(backup.Handler.AgeWarn, "age_warn", environment, &backup, &repository)
			return true, false, nil
		}
	}

	return false, false, nil
}

func (r Restic) GetLastBackupTimestamp(backup Backup, repository Repository) (time.Time, error) {

	cmd := r.prepareResticCommand(repository, backup.Environment)
	cmd.Args = append(cmd.Args, "snapshots", "--json")

	var output bytes.Buffer
	cmd.Stdout = &output

	err := cmd.Run()
	if err != nil {
		fmt.Fprintf(
			os.Stderr, "Failed to get age for backup [%s] in repository [%s]: %s\n",
			backup.Name,
			repository.Name,
			err,
		)
		return time.Time{}, err
	}

	type Snapshot struct {
		Time     time.Time
		Hostname string
		Paths    []string
		Tags     []string
	}

	var snapshots []Snapshot

	if err := json.Unmarshal(output.Bytes(), &snapshots); err != nil {
		return time.Time{}, err
	}

	hostname, err := os.Hostname()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get hostname: %s\n", err)
	}

	lastBackupTimestamp := time.Time{}
	for _, s := range snapshots {

		if backup.DataStdinCommand != "" {
			if len(s.Paths) != 1 {
				continue
			}
			if !strings.HasSuffix(s.Paths[0], backup.StdinFilename) {
				continue
			}
		} else {
			if !comparePathList(s.Paths, backup.Data) {
				continue
			}
		}

		if s.Hostname != hostname || !compareStringList(s.Tags, backup.Tags) {
			continue
		}

		if s.Time.After(lastBackupTimestamp) {
			lastBackupTimestamp = s.Time
		}
	}

	return lastBackupTimestamp, nil
}

func (r Restic) runUnlock(repository Repository) error {

	cmd := r.prepareResticCommand(repository, make(map[string]string))
	cmd.Args = append(cmd.Args, "unlock")

	return cmd.Run()
}

func (r Restic) dumpUnlockError(repository Repository, err error) {
	fmt.Fprintf(
		os.Stderr, "Failed to unlock repository [%s]: %s\n",
		repository.Name,
		err,
	)
}

func (r Restic) prepareResticCommand(
	repo Repository, additionalEnvironment map[string]string,
) *exec.Cmd {
	environment := combineMaps(repo.Environment, additionalEnvironment)
	return r.PrepareResticEnvironmentCommand(
		r.resticExecutable, repo.URL, repo.Password, environment,
		repo.LimitDownload, repo.LimitUpload,
	)
}

func (r Restic) PrepareResticEnvironmentCommand(
	command string, repoURL string, password string, environment map[string]string,
	limitDownload int, limitUpload int,
) *exec.Cmd {
	cmd := exec.Command(command)

	if limitDownload > 0 {
		cmd.Args = append(cmd.Args, "--limit-download", strconv.Itoa(limitDownload))
	}
	if limitUpload > 0 {
		cmd.Args = append(cmd.Args, "--limit-upload", strconv.Itoa(limitUpload))
	}

	cmd.Env = append(
		os.Environ(),
		convertEnvironment(environment)...,
	)
	cmd.Env = append(cmd.Env, fmt.Sprintf("RESTIC_REPOSITORY=%s", repoURL))
	cmd.Env = append(cmd.Env, fmt.Sprintf("RESTIC_PASSWORD=%s", password))

	return cmd
}

func runHandler(command string, handlerName string, environment map[string]string, backup *Backup, repository *Repository) {

	if command == "" {
		return
	}

	commandToRun, err := homedir.Expand(command)
	if err != nil {
		fmt.Fprintf(
			os.Stderr, "Failed to expand homedir in command: %s\n",
			err,
		)
	}

	tmpl := template.New("cmd")
	tmpl, err = tmpl.Parse(commandToRun)
	if err != nil {
		fmt.Fprintf(
			os.Stderr, "Failed to expand template variables in command: %s\n",
			err,
		)
	} else {

		type TemplateArgs struct {
			BackupName     string
			RepositoryName string
			RepositoryURL  string
		}

		args := TemplateArgs{}
		if backup != nil {
			args.BackupName = backup.Name
		}
		if repository != nil {
			args.RepositoryName = repository.Name
			args.RepositoryURL = repository.URL
		}

		var buffer bytes.Buffer
		tmpl.Execute(&buffer, args)
		commandToRun = buffer.String()
	}

	cmd, err := prepareShellCommand(commandToRun, environment)

	if err != nil {
		fmt.Fprintf(
			os.Stderr, "Failed to run handler [%s] \"%s\": %s\n",
			handlerName, commandToRun, err,
		)
	}

	err = cmd.Run()

	if err != nil {
		fmt.Fprintf(
			os.Stderr, "Failed to run handler [%s] \"%s\": %s\n",
			handlerName, commandToRun, err,
		)
	}
}

func runHandlerCheckFailure(repository Repository) {
	runHandler(repository.Handler.CheckFailure, "check_failure", repository.Environment, nil, &repository)
}

func runHandlerBackupFailure(backup Backup, repository Repository, environment map[string]string) {
	runHandler(backup.Handler.Failure, "failure", environment, &backup, &repository)
}

func runHandlerForgetFailure(repository Repository) {
	runHandler(repository.Handler.ForgetFailure, "forget_failure", repository.Environment, nil, &repository)
}

func runHandlerAgeError(backup Backup, repository Repository, environment map[string]string) {
	runHandler(backup.Handler.AgeError, "age_error", environment, &backup, &repository)
}

func prepareShellCommand(command string, environment map[string]string) (*exec.Cmd, error) {

	args, err := shlex.Split(command, true)

	if err != nil {
		return nil, err
	}

	args0 := ""
	args1 := []string{}

	if len(args) > 0 {
		args0 = args[0]
	}

	if len(args) > 1 {
		args1 = args[1:]
	}

	cmd := exec.Command(args0, args1...)
	cmd.Env = append(
		os.Environ(),
		convertEnvironment(environment)...,
	)

	return cmd, err
}

func convertEnvironment(env map[string]string) []string {
	var result []string

	for k, v := range env {
		result = append(result, fmt.Sprintf("%s=%s", k, v))
	}

	return result
}

func combineMaps(a, b map[string]string) map[string]string {
	result := make(map[string]string)

	for k, v := range a {
		result[k] = v
	}

	for k, v := range b {
		result[k] = v
	}

	return result
}
