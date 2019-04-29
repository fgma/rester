package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"runtime"

	jsonutil "github.com/vrischmann/jsonutil"
)

type Policy struct {
	KeepLast    uint     `json:"keep_last,omitempty"`
	KeepHourly  uint     `json:"keep_hourly,omitempty"`
	KeepDaily   uint     `json:"keep_daily,omitempty"`
	KeepWeekly  uint     `json:"keep_weekly,omitempty"`
	KeepMonthly uint     `json:"keep_monthly,omitempty"`
	KeepYearly  uint     `json:"keep_yearly,omitempty"`
	KeepWithin  string   `json:"keep_within,omitempty"`
	KeepTags    []string `json:"keep_tags,omitempty"`
}

type Check struct {
	ReadDataPercentage uint `json:"read_data_percentage,omitempty"`
}

type RepositoryHandler struct {
	ForgetSuccess string `json:"forget_success,omitempty"`
	ForgetFailure string `json:"forget_failure,omitempty"`
	CheckSuccess  string `json:"check_success,omitempty"`
	CheckFailure  string `json:"check_failure,omitempty"`
}

type repositoryDefaultable struct {
	Policy        Policy            `json:"policy,omitempty"`
	Handler       RepositoryHandler `json:"handler,omitempty"`
	LimitDownload int               `json:"limit_download,omitempty"`
	LimitUpload   int               `json:"limit_upload,omitempty"`
}

type Repository struct {
	Name        string            `json:"name,omitempty"`
	URL         string            `json:"url,omitempty"`
	Password    string            `json:"password,omitempty"`
	Environment map[string]string `json:"environment,omitempty"`
	Check       Check             `json:"check,omitempty"`
	repositoryDefaultable
}

type BackupHandler struct {
	Before   string `json:"before,omitempty"`
	After    string `json:"after,omitempty"`
	Success  string `json:"success,omitempty"`
	Failure  string `json:"failure,omitempty"`
	AgeWarn  string `json:"age_warn,omitempty"`
	AgeError string `json:"age_error,omitempty"`
}

type BackupAge struct {
	Warn  jsonutil.Duration `json:"warn,omitempty"`
	Error jsonutil.Duration `json:"error,omitempty"`
}

type backupDefaultable struct {
	Handler BackupHandler `json:"handler,omitempty"`
	Age     BackupAge     `json:"age,omitempty"`
}

type Backup struct {
	Name             string            `json:"name,omitempty"`
	Repository       string            `json:"repository,omitempty"`
	Data             []string          `json:"data,omitempty"`
	DataStdinCommand string            `json:"data_stdin_command,omitempty"`
	StdinFilename    string            `json:"stdin_filename,omitempty"`
	Exclude          []string          `json:"exclude,omitempty"`
	OneFileSystem    bool              `json:"one_file_system,omitempty"`
	Tags             []string          `json:"tags,omitempty"`
	Environment      map[string]string `json:"environment,omitempty"`
	CustomFlags      []string          `json:"custom_flags,omitempty"`
	backupDefaultable
}

type Defaults struct {
	Repositories repositoryDefaultable `json:"repositories,omitempty"`
	Backups      backupDefaultable     `json:"backups,omitempty"`
}

type Config struct {
	ResticExecutable string       `json:"restic_executable,omitempty"`
	Defaults         Defaults     `json:"defaults,omitempty"`
	Repositories     []Repository `json:"repositories,omitempty"`
	Backups          []Backup     `json:"backups,omitempty"`
}

func (c *Config) GetRepositoryByName(name string) *Repository {
	for _, repo := range c.Repositories {
		if repo.Name == name {
			return &repo
		}
	}

	return nil
}

func (c *Config) GetBackupByName(name string) *Backup {
	for _, backup := range c.Backups {
		if backup.Name == name {
			return &backup
		}
	}

	return nil
}

type ValidationError struct {
	s string
}

func (e ValidationError) Error() string {
	return e.s
}

func Load(configFile string) (Config, error) {

	jsonFile, err := os.Open(configFile)

	if err != nil {
		return Config{}, err
	}

	defer jsonFile.Close()

	return LoadFromReader(jsonFile)
}

func LoadFromReader(reader io.Reader) (Config, error) {
	bytes, _ := ioutil.ReadAll(reader)

	var config = Config{
		ResticExecutable: "restic",
	}

	if err := json.Unmarshal(bytes, &config); err != nil {
		return Config{}, err
	}

	fillFromDefaults(&config)

	if err := validate(&config); err != nil {
		return Config{}, err
	}

	return config, nil
}

func fillFromDefaults(config *Config) {
	for i := range config.Repositories {
		if config.Repositories[i].Handler.CheckFailure == "" {
			config.Repositories[i].Handler.CheckFailure = config.Defaults.Repositories.Handler.CheckFailure
		}
		if config.Repositories[i].Handler.CheckSuccess == "" {
			config.Repositories[i].Handler.CheckSuccess = config.Defaults.Repositories.Handler.CheckSuccess
		}
		if config.Repositories[i].Handler.ForgetFailure == "" {
			config.Repositories[i].Handler.ForgetFailure = config.Defaults.Repositories.Handler.ForgetFailure
		}
		if config.Repositories[i].Handler.ForgetSuccess == "" {
			config.Repositories[i].Handler.ForgetSuccess = config.Defaults.Repositories.Handler.ForgetSuccess
		}
		if config.Repositories[i].Policy.KeepLast == 0 {
			config.Repositories[i].Policy.KeepLast = config.Defaults.Repositories.Policy.KeepLast
		}
		if config.Repositories[i].Policy.KeepHourly == 0 {
			config.Repositories[i].Policy.KeepHourly = config.Defaults.Repositories.Policy.KeepHourly
		}
		if config.Repositories[i].Policy.KeepDaily == 0 {
			config.Repositories[i].Policy.KeepDaily = config.Defaults.Repositories.Policy.KeepDaily
		}
		if config.Repositories[i].Policy.KeepWeekly == 0 {
			config.Repositories[i].Policy.KeepWeekly = config.Defaults.Repositories.Policy.KeepWeekly
		}
		if config.Repositories[i].Policy.KeepMonthly == 0 {
			config.Repositories[i].Policy.KeepMonthly = config.Defaults.Repositories.Policy.KeepMonthly
		}
		if config.Repositories[i].Policy.KeepYearly == 0 {
			config.Repositories[i].Policy.KeepYearly = config.Defaults.Repositories.Policy.KeepYearly
		}
		if config.Repositories[i].Policy.KeepWithin == "" {
			config.Repositories[i].Policy.KeepWithin = config.Defaults.Repositories.Policy.KeepWithin
		}
		if len(config.Repositories[i].Policy.KeepTags) == 0 {
			config.Repositories[i].Policy.KeepTags = config.Defaults.Repositories.Policy.KeepTags
		}
		if config.Repositories[i].LimitDownload == 0 {
			config.Repositories[i].LimitDownload = config.Defaults.Repositories.LimitDownload
		}
		if config.Repositories[i].LimitUpload == 0 {
			config.Repositories[i].LimitUpload = config.Defaults.Repositories.LimitUpload
		}
	}
	for i := range config.Backups {
		if config.Backups[i].Handler.Before == "" {
			config.Backups[i].Handler.Before = config.Defaults.Backups.Handler.Before
		}
		if config.Backups[i].Handler.After == "" {
			config.Backups[i].Handler.After = config.Defaults.Backups.Handler.After
		}
		if config.Backups[i].Handler.Success == "" {
			config.Backups[i].Handler.Success = config.Defaults.Backups.Handler.Success
		}
		if config.Backups[i].Handler.Failure == "" {
			config.Backups[i].Handler.Failure = config.Defaults.Backups.Handler.Failure
		}
		if config.Backups[i].Handler.AgeWarn == "" {
			config.Backups[i].Handler.AgeWarn = config.Defaults.Backups.Handler.AgeWarn
		}
		if config.Backups[i].Handler.AgeError == "" {
			config.Backups[i].Handler.AgeError = config.Defaults.Backups.Handler.AgeError
		}
		if (config.Backups[i].Age.Warn == jsonutil.Duration{}) {
			config.Backups[i].Age.Warn = config.Defaults.Backups.Age.Warn
		}
		if (config.Backups[i].Age.Error == jsonutil.Duration{}) {
			config.Backups[i].Age.Error = config.Defaults.Backups.Age.Error
		}
	}
}

func validate(config *Config) error {

	if config.ResticExecutable == "" {
		return ValidationError{"Restic executablepath is empty."}
	}

	repoNames := make(map[string]bool)

	for _, v := range config.Repositories {
		if err := validateRepository(&v); err != nil {
			return err
		}

		if _, ok := repoNames[v.Name]; ok {
			return ValidationError{fmt.Sprintf("Repository name %s is used multiple times.", v.Name)}
		}
		repoNames[v.Name] = true
	}

	for _, v := range config.Backups {
		if err := validateBackup(&v, repoNames); err != nil {
			return err
		}
	}

	return nil
}

func validateRepository(repo *Repository) error {

	if repo.Name == "" {
		return ValidationError{"Repository has no name."}
	}

	if repo.URL == "" {
		return ValidationError{"Repository has no URL."}
	}

	if repo.Password == "" {
		return ValidationError{"Repository has no password."}
	}

	if repo.Check.ReadDataPercentage < 0 || repo.Check.ReadDataPercentage > 100 {
		return ValidationError{"Repository check read data percentage outside expected range [0,100]"}
	}

	return nil
}

func validateBackup(backup *Backup, repoNames map[string]bool) error {

	if backup.Name == "" {
		return ValidationError{"Backup has no name."}
	}

	if backup.Repository == "" {
		return ValidationError{"Backup has no repository."}
	}

	if _, ok := repoNames[backup.Repository]; !ok {
		return ValidationError{fmt.Sprintf("Backup repository %s not defined.", backup.Repository)}
	}

	if len(backup.Data) > 0 && backup.DataStdinCommand != "" {
		return ValidationError{"Backup can't use data from filesystem and stdin."}
	}

	if len(backup.Data) == 0 && backup.DataStdinCommand == "" {
		return ValidationError{"Backup needs something to backup."}
	}

	if backup.DataStdinCommand != "" && backup.StdinFilename == "" {
		return ValidationError{"Backup from stdin needs a stdin filename."}
	}

	if backup.Age.Error.Nanoseconds() < backup.Age.Warn.Nanoseconds() {
		return ValidationError{"Backup age error limit < warn limit."}
	}

	if runtime.GOOS == "windows" && backup.OneFileSystem {
		fmt.Println("Warning: restic option --one-file-system does not work as expected on windows yet.")
	}

	return nil
}
