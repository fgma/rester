package internal

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	reader := strings.NewReader(`{
		"repositories": [
			{
				"name": "test1",
				"url": "/home/test/repos/test1",
				"password": "1"
			},
			{
				"name": "test2",
				"url": "/home/test/repos/test2",
				"password": "2",
				"handler": {
					"check_success": "notify_send -t 1000 check done",
					"check_failure": "notify_send check failed",
					"forget_success": "notify_send -t 1000 forget done",
					"forget_failure": "notify_send forget failed"
				},
				"check": {
					"read_data_percentage": 5
				}
			}
		],
		"backups": [
			{
				"name": "some_data",
				"repositories": [ "test1" ],
				"data": [
					"/etc/",
					"/var/lib/",
					"/tmp/test3/"
				],
				"exclude": [ "*.tmp", "*.abc" ],
				"one_file_system": true,            
				"tags": [ "config", "etc" ],
				"custom_flags": [ "--ignore-inode" ],
				"handler": {
					"before": "notify_send -t 1000 Backup before",
					"after": "notify_send -t 1000 Backup after",
					"success": "notify_send -t 1000 Backup done",
					"failure": "notify_send Backup failed"
				}
			},
			{
				"name": "mysql",
				"repositories": [ "test2" ],
				"data_stdin_command": "mysqldump",
				"stdin_filename": "mysqldump.sql",
				"tags": [ "db", "mysql" ]
			},
			{
				"name": "some_other_data",
				"repositories": [ "test2", "test1" ],
				"data": [ "/etc/" ],
				"exclude": [ "*.tmp", "*.bcd" ],
				"one_file_system": false,            
				"tags": [ "etc" ]
			}
		]
	}`)

	config, error := LoadFromReader(reader)
	if error != nil {
		t.Errorf("Failed to parse: %s", error)
	}

	//
	// repositories
	//
	assert.Equal(t, 2, len(config.Repositories), "invalid repository count")
	assert.Equal(t, "test1", config.Repositories[0].Name)
	assert.Equal(t, "/home/test/repos/test1", config.Repositories[0].URL)
	assert.Equal(t, "1", config.Repositories[0].Password)
	assert.Equal(t, uint(0), config.Repositories[0].Check.ReadDataPercentage)

	assert.Equal(t, "test2", config.Repositories[1].Name)
	assert.Equal(t, "/home/test/repos/test2", config.Repositories[1].URL)
	assert.Equal(t, "2", config.Repositories[1].Password)
	assert.Equal(t, uint(5), config.Repositories[1].Check.ReadDataPercentage)
	assert.Equal(t, "notify_send -t 1000 check done", config.Repositories[1].Handler.CheckSuccess)
	assert.Equal(t, "notify_send check failed", config.Repositories[1].Handler.CheckFailure)
	assert.Equal(t, "notify_send -t 1000 forget done", config.Repositories[1].Handler.ForgetSuccess)
	assert.Equal(t, "notify_send forget failed", config.Repositories[1].Handler.ForgetFailure)

	//
	// backups:
	//
	assert.Equal(t, 3, len(config.Backups), "invalid backup count")

	//
	// backup 0
	//
	assert.Equal(t, "some_data", config.Backups[0].Name)
	assert.Equal(t, 1, len(config.Backups[0].Repositories))
	assert.Equal(t, "test1", config.Backups[0].Repositories[0])

	assert.Equal(t, 3, len(config.Backups[0].Data))
	assert.Equal(t, "/etc/", config.Backups[0].Data[0])
	assert.Equal(t, "/var/lib/", config.Backups[0].Data[1])
	assert.Equal(t, "/tmp/test3/", config.Backups[0].Data[2])

	assert.Equal(t, 2, len(config.Backups[0].Exclude))
	assert.Equal(t, "*.tmp", config.Backups[0].Exclude[0])
	assert.Equal(t, "*.abc", config.Backups[0].Exclude[1])

	assert.Equal(t, true, config.Backups[0].OneFileSystem)

	assert.Equal(t, 2, len(config.Backups[0].Tags), "invalid tag count")
	assert.Equal(t, "config", config.Backups[0].Tags[0])
	assert.Equal(t, "etc", config.Backups[0].Tags[1])

	assert.Equal(t, 1, len(config.Backups[0].CustomFlags), "invalid custom flag")
	assert.Equal(t, "--ignore-inode", config.Backups[0].CustomFlags[0])

	assert.Equal(t, "notify_send -t 1000 Backup before", config.Backups[0].Handler.Before)
	assert.Equal(t, "notify_send -t 1000 Backup after", config.Backups[0].Handler.After)
	assert.Equal(t, "notify_send -t 1000 Backup done", config.Backups[0].Handler.Success)
	assert.Equal(t, "notify_send Backup failed", config.Backups[0].Handler.Failure)

	//
	// backup 1
	//
	assert.Equal(t, "mysql", config.Backups[1].Name)
	assert.Equal(t, 1, len(config.Backups[1].Repositories))
	assert.Equal(t, "test2", config.Backups[1].Repositories[0])
	assert.Equal(t, "mysqldump", config.Backups[1].DataStdinCommand)
	assert.Equal(t, "mysqldump.sql", config.Backups[1].StdinFilename)
	assert.Equal(t, false, config.Backups[1].OneFileSystem)

	assert.Equal(t, 2, len(config.Backups[1].Tags), "invalid tag count")
	assert.Equal(t, "db", config.Backups[1].Tags[0])
	assert.Equal(t, "mysql", config.Backups[1].Tags[1])

	assert.Equal(t, "", config.Backups[1].Handler.Before)
	assert.Equal(t, "", config.Backups[1].Handler.After)
	assert.Equal(t, "", config.Backups[1].Handler.Success)
	assert.Equal(t, "", config.Backups[1].Handler.Failure)

	//
	// backup 2
	//
	assert.Equal(t, "some_other_data", config.Backups[2].Name)
	assert.Equal(t, 2, len(config.Backups[2].Repositories))
	assert.Equal(t, "test2", config.Backups[2].Repositories[0])
	assert.Equal(t, "test1", config.Backups[2].Repositories[1])

	assert.Equal(t, 1, len(config.Backups[2].Data))
	assert.Equal(t, "/etc/", config.Backups[2].Data[0])

	assert.Equal(t, 2, len(config.Backups[2].Exclude))
	assert.Equal(t, "*.tmp", config.Backups[2].Exclude[0])
	assert.Equal(t, "*.bcd", config.Backups[2].Exclude[1])

	assert.Equal(t, false, config.Backups[2].OneFileSystem)

	assert.Equal(t, 1, len(config.Backups[2].Tags), "invalid tag count")
	assert.Equal(t, "etc", config.Backups[2].Tags[0])

	assert.Equal(t, "", config.Backups[2].Handler.Before)
	assert.Equal(t, "", config.Backups[2].Handler.After)
	assert.Equal(t, "", config.Backups[2].Handler.Success)
	assert.Equal(t, "", config.Backups[2].Handler.Failure)

}

func TestLoadConfigDataAndStdinShoulFail(t *testing.T) {
	reader := strings.NewReader(`{
		"repositories": [
			{
				"name": "test1",
				"url": "/home/test/repos/test1",
				"password": "1"
			}
		],
		"backups": [
			{
				"name": "mysql",
				"repositories": [ "test1" ],
				"data": [ "/etc/" ],
				"data_stdin_command": "mysqldump",            
				"tags": [ "db", "mysql" ]
			}
		]
	}`)

	_, error := LoadFromReader(reader)
	assert.True(t, error != nil)
	assert.True(t, strings.Contains(error.Error(), "Backup can't use data from filesystem and stdin."))
}

func TestLoadConfigDataWithouthStdinFilenameShoulFail(t *testing.T) {
	reader := strings.NewReader(`{
		"repositories": [
			{
				"name": "test1",
				"url": "/home/test/repos/test1",
				"password": "1"
			}
		],
		"backups": [
			{
				"name": "mysql",
				"repositories": [ "test1" ],
				"data_stdin_command": "mysqldump",
				"tags": [ "db", "mysql" ]
			}
		]
	}`)

	_, error := LoadFromReader(reader)
	assert.True(t, error != nil)
	assert.True(t, strings.Contains(error.Error(), "Backup from stdin needs a stdin filename."))
}

func TestLoadConfigNothingToBackup(t *testing.T) {
	reader := strings.NewReader(`{
		"repositories": [
			{
				"name": "test1",
				"url": "/home/test/repos/test1",
				"password": "1"
			}
		],
		"backups": [
			{
				"name": "mysql",
				"repositories": [ "test1" ],
				"tags": [ "db", "mysql" ]
			}
		]
	}`)

	_, error := LoadFromReader(reader)
	assert.True(t, error != nil)
	assert.True(t, strings.Contains(error.Error(), "Backup needs something to backup."))
}

func TestLoadConfigInvalidRepostory(t *testing.T) {
	reader := strings.NewReader(`{
		"repositories": [
			{
				"name": "test1",
				"url": "/home/test/repos/test1",
				"password": "1"
			}
		],
		"backups": [
			{
				"name": "mysql",
				"repositories": [ "INVALID REPOSITORY" ],
				"data_stdin_command": "mysqldump"
			}
		]
	}`)

	_, error := LoadFromReader(reader)
	assert.True(t, error != nil)
	assert.True(t, strings.Contains(error.Error(), "Backup repository INVALID REPOSITORY not defined."))
}

func TestLoadConfigInvalidRepositoryName(t *testing.T) {
	reader := strings.NewReader(`{
		"repositories": [
			{
				"name": "test1/ ",
				"url": "/home/test/repos/test1",
				"password": "1"
			}
		],
		"backups": [
			{
				"name": "mysql",
				"repositories": [ "test1" ],
				"data_stdin_command": "mysqldump"
			}
		]
	}`)

	_, error := LoadFromReader(reader)
	assert.True(t, error != nil)
	assert.True(t, strings.Contains(error.Error(), "Repository name contains invalid character."))
}

func TestLoadConfigInvalidBackupName(t *testing.T) {
	reader := strings.NewReader(`{
		"repositories": [
			{
				"name": "test1",
				"url": "/home/test/repos/test1",
				"password": "1"
			}
		],
		"backups": [
			{
				"name": "mysql/ 123",
				"repositories": [ "test1" ],
				"data_stdin_command": "mysqldump"
			}
		]
	}`)

	_, error := LoadFromReader(reader)
	assert.True(t, error != nil)
	assert.True(t, strings.Contains(error.Error(), "Backup name contains invalid character."))
}

func TestGetRepositoryByName(t *testing.T) {
	reader := strings.NewReader(`{
		"repositories": [
			{
				"name": "test1",
				"url": "/home/test/repos/test1",
				"password": "1"
			},
			{
				"name": "abc",
				"url": "/home/test/repos/abc",
				"password": "2"
			}
		],
		"backups": []
	}`)

	config, error := LoadFromReader(reader)

	assert.Zero(t, error)

	abc := config.GetRepositoryByName("abc")
	assert.NotZero(t, abc)
	assert.Equal(t, "abc", abc.Name)

	repo1 := config.GetRepositoryByName("test1")
	assert.NotZero(t, repo1)
	assert.Equal(t, "test1", repo1.Name)

	assert.Zero(t, config.GetRepositoryByName("undefined"))
}

func TestGetBackupByName(t *testing.T) {
	reader := strings.NewReader(`{
		"repositories": [
			{
				"name": "test1",
				"url": "/home/test/repos/test1",
				"password": "1"
			},
			{
				"name": "abc",
				"url": "/home/test/repos/abc",
				"password": "2"
			}
		],
		"backups": [
			{
				"name": "mysql",
				"repositories": [ "test1" ],
				"data_stdin_command": "mysqldump",
				"stdin_filename": "mysqldump.sql"
			},
			{
				"name": "some_data",
				"repositories": [ "test1" ],
				"data_stdin_command": "mysqldump",
				"stdin_filename": "some_data.sql"
			}
		]
	}`)

	config, error := LoadFromReader(reader)

	assert.Zero(t, error)

	mysql := config.GetBackupByName("mysql")
	assert.NotZero(t, mysql)
	assert.Equal(t, "mysql", mysql.Name)

	someData := config.GetBackupByName("some_data")
	assert.NotZero(t, someData)
	assert.Equal(t, "some_data", someData.Name)

	assert.Zero(t, config.GetBackupByName("undefined"))
}

func TestLoadConfigWithResticExecutable(t *testing.T) {
	reader := strings.NewReader(`{
		"restic_executable": "restic_custom",
		"repositories": [
			{
				"name": "test1",
				"url": "/home/test/repos/test1",
				"password": "1"
			}
		],
		"backups": [
			{
				"name": "mysql",
				"repositories": [ "test1" ],
				"data": [ "/etc/" ],
				"tags": [ "db", "mysql" ]
			}
		]
	}`)

	c, error := LoadFromReader(reader)
	assert.True(t, error == nil)
	assert.Equal(t, "restic_custom", c.ResticExecutable)
}

func TestLoadConfigWithCustomEnvironment(t *testing.T) {
	reader := strings.NewReader(`{
		"repositories": [
			{
				"name": "test1",
				"url": "/home/test/repos/test1",
				"password": "1",
				"environment": {
					"AWS_ACCESS_KEY_ID": "<YOUR-MINIO-ACCESS-KEY-ID>",
					"AWS_SECRET_ACCESS_KEY": "<YOUR-MINIO-SECRET-ACCESS-KEY>"
				}
			}
		],
		"backups": [
			{
				"name": "mysql",
				"repositories": [ "test1" ],
				"data": [ "/etc/" ],
				"tags": [ "db", "mysql" ],
				"environment": {
					"AWS_ACCESS_KEY_ID": "<OVERRRIDE-REPO-KEY>",
					"AWS_SECRET_ACCESS_KEY": "<OVERRRIDE-REPO-SECRET>"
				}
			}
		]
	}`)

	c, error := LoadFromReader(reader)
	assert.True(t, error == nil)
	assert.Equal(t, "<YOUR-MINIO-ACCESS-KEY-ID>", c.Repositories[0].Environment["AWS_ACCESS_KEY_ID"])
	assert.Equal(t, "<YOUR-MINIO-SECRET-ACCESS-KEY>", c.Repositories[0].Environment["AWS_SECRET_ACCESS_KEY"])
	assert.Equal(t, "<OVERRRIDE-REPO-KEY>", c.Backups[0].Environment["AWS_ACCESS_KEY_ID"])
	assert.Equal(t, "<OVERRRIDE-REPO-SECRET>", c.Backups[0].Environment["AWS_SECRET_ACCESS_KEY"])
}

func TestLoadConfigWithPolicy(t *testing.T) {
	reader := strings.NewReader(`{
		"repositories": [
			{
				"name": "test1",
				"url": "/home/test/repos/test1",
				"password": "1",
				"policy": {
					"keep_last": 5,
					"keep_hourly": 6,
					"keep_daily": 7,
					"keep_weekly": 8,
					"keep_monthly": 9,
					"keep_yearly": 10,
					"keep_within": "2y5m7d",
					"keep_tags": [ "data", "test" ]
				}
			}
		],
		"backups": []
	}`)

	c, error := LoadFromReader(reader)
	assert.True(t, error == nil)
	assert.Equal(t, uint(5), c.Repositories[0].Policy.KeepLast)
	assert.Equal(t, uint(6), c.Repositories[0].Policy.KeepHourly)
	assert.Equal(t, uint(7), c.Repositories[0].Policy.KeepDaily)
	assert.Equal(t, uint(8), c.Repositories[0].Policy.KeepWeekly)
	assert.Equal(t, uint(9), c.Repositories[0].Policy.KeepMonthly)
	assert.Equal(t, uint(10), c.Repositories[0].Policy.KeepYearly)
	assert.Equal(t, "2y5m7d", c.Repositories[0].Policy.KeepWithin)
	assert.Equal(t, 2, len(c.Repositories[0].Policy.KeepTags))
	assert.Equal(t, "data", c.Repositories[0].Policy.KeepTags[0])
	assert.Equal(t, "test", c.Repositories[0].Policy.KeepTags[1])
}

func TestLoadConfigWithAge(t *testing.T) {
	reader := strings.NewReader(`{
		"repositories": [
			{
				"name": "test1",
				"url": "/home/test/repos/test1",
				"password": "1"
			}
		],
		"backups": [
			{
				"name": "mysql",
				"repositories": [ "test1" ],
				"data": [ "/etc/" ],
				"tags": [ "db", "mysql" ],
				"age": {
					"warn": "1h30m",
					"error": "16h30m15s"					
				}
			}
		]
	}`)

	c, error := LoadFromReader(reader)
	assert.True(t, error == nil)

	warn, _ := time.ParseDuration("1h30m")
	err, _ := time.ParseDuration("16h30m15s")

	assert.Equal(t, 1, len(c.Backups))
	assert.Equal(t, warn, c.Backups[0].Age.Warn.Duration)
	assert.Equal(t, err, c.Backups[0].Age.Error.Duration)
}

func TestLoadConfigWithAgeHandler(t *testing.T) {
	reader := strings.NewReader(`{
		"repositories": [
			{
				"name": "test1",
				"url": "/home/test/repos/test1",
				"password": "1"
			}
		],
		"backups": [
			{
				"name": "mysql",
				"repositories": [ "test1" ],
				"data": [ "/etc/" ],
				"tags": [ "db", "mysql" ],
				"age": {
					"warn": "1h30m",
					"error": "16h30m15s"
				},
				"handler": {
					"age_warn": "/some/warning",
					"age_error": "/some/error"
				}
			}
		]
	}`)

	c, error := LoadFromReader(reader)
	assert.True(t, error == nil)

	warn, _ := time.ParseDuration("1h30m")
	err, _ := time.ParseDuration("16h30m15s")

	assert.Equal(t, 1, len(c.Backups))
	assert.Equal(t, warn, c.Backups[0].Age.Warn.Duration)
	assert.Equal(t, err, c.Backups[0].Age.Error.Duration)

	assert.Equal(t, "/some/warning", c.Backups[0].Handler.AgeWarn)
	assert.Equal(t, "/some/error", c.Backups[0].Handler.AgeError)
}

func TestLoadConfigWithAgeWarnLimitAboveErrorLimitShouldFail(t *testing.T) {
	reader := strings.NewReader(`{
		"repositories": [
			{
				"name": "test1",
				"url": "/home/test/repos/test1",
				"password": "1"
			}
		],
		"backups": [
			{
				"name": "mysql",
				"repositories": [ "test1" ],
				"data": [ "/etc/" ],
				"tags": [ "db", "mysql" ],
				"age": {
					"warn": "1h30m",
					"error": "1h"					
				}
			}
		]
	}`)

	_, error := LoadFromReader(reader)
	assert.True(t, error != nil)
}

func TestLoadConfigWithDefaults(t *testing.T) {
	reader := strings.NewReader(`{
		"defaults": {
			"repositories": {
				"handler": {
					"check_success": "notify_send -t 1000 check done",
					"check_failure": "notify_send check failed",
					"forget_success": "notify_send -t 1000 forget done",
					"forget_failure": "notify_send forget failed"
				},
				"policy": {
					"keep_last": 5,
					"keep_hourly": 6,
					"keep_daily": 7,
					"keep_weekly": 8,
					"keep_monthly": 9,
					"keep_yearly": 10,
					"keep_within": "2y5m7d",
					"keep_tags": [ "data", "test" ]
				}
			},
			"backups": {
				"age": {
					"warn": "1h30m",
					"error": "16h30m15s"
				},
				"handler": { 
					"before": "notify_send -t 5000 START \"backing up {{.BackupName}}\"",
					"success": "notify_send -t 5000 SUCCESS \"{{.BackupName}} has been backed up\"",
					"failure": "notify_send -t 0 FAILED \"{{.BackupName}} has NOT been backed up\"",
					"after": "notify_send -t 0 FAILED \"{{.BackupName}} cleanup\"",
					"age_warn": "notify_send -t 15000 WARNING \"{{.BackupName}} backup to old\"",
					"age_error": "notify_send -t 0 FAILED \"{{.BackupName}} has not been backed up in time\""
				}
			}
		},
		"repositories": [
			{
				"name": "test1",
				"url": "/home/test/repos/test1",
				"password": "1"
			},
			{
				"name": "test2",
				"url": "/home/test/repos/test2",
				"password": "2"
			}
		],
		"backups": [
			{
				"name": "mysql",
				"repositories": [ "test1" ],
				"data": [ "/etc/" ],
				"tags": [ "db", "mysql" ]
			}
		]
	}`)

	c, error := LoadFromReader(reader)
	assert.True(t, error == nil)

	assert.Equal(t, c.Defaults.Repositories.Handler.CheckSuccess, "notify_send -t 1000 check done")
	assert.Equal(t, 2, len(c.Repositories))
	assert.Equal(t, c.Repositories[0].Handler.CheckSuccess, "notify_send -t 1000 check done")
	assert.Equal(t, c.Repositories[1].Handler.CheckSuccess, "notify_send -t 1000 check done")
	assert.Equal(t, c.Repositories[0].Handler.CheckFailure, "notify_send check failed")
	assert.Equal(t, c.Repositories[0].Handler.ForgetSuccess, "notify_send -t 1000 forget done")
	assert.Equal(t, c.Repositories[0].Handler.ForgetFailure, "notify_send forget failed")
	assert.Equal(t, c.Repositories[0].Policy.KeepLast, uint(5))
	assert.Equal(t, c.Repositories[0].Policy.KeepHourly, uint(6))
	assert.Equal(t, c.Repositories[0].Policy.KeepDaily, uint(7))
	assert.Equal(t, c.Repositories[0].Policy.KeepWeekly, uint(8))
	assert.Equal(t, c.Repositories[0].Policy.KeepMonthly, uint(9))
	assert.Equal(t, c.Repositories[0].Policy.KeepYearly, uint(10))
	assert.Equal(t, c.Repositories[0].Policy.KeepWithin, "2y5m7d")
	assert.Equal(t, c.Repositories[0].Policy.KeepTags, []string{"data", "test"})

	assert.Equal(t, c.Defaults.Backups.Handler.Before, "notify_send -t 5000 START \"backing up {{.BackupName}}\"")
	assert.Equal(t, 1, len(c.Backups))
	assert.Equal(t, c.Backups[0].Handler.Before, "notify_send -t 5000 START \"backing up {{.BackupName}}\"")
	assert.Equal(t, c.Backups[0].Handler.Success, "notify_send -t 5000 SUCCESS \"{{.BackupName}} has been backed up\"")
	assert.Equal(t, c.Backups[0].Handler.Failure, "notify_send -t 0 FAILED \"{{.BackupName}} has NOT been backed up\"")
	assert.Equal(t, c.Backups[0].Handler.After, "notify_send -t 0 FAILED \"{{.BackupName}} cleanup\"")
	assert.Equal(t, c.Backups[0].Handler.AgeWarn, "notify_send -t 15000 WARNING \"{{.BackupName}} backup to old\"")
	assert.Equal(t, c.Backups[0].Handler.AgeError, "notify_send -t 0 FAILED \"{{.BackupName}} has not been backed up in time\"")

	warn, _ := time.ParseDuration("1h30m")
	err, _ := time.ParseDuration("16h30m15s")
	assert.Equal(t, c.Backups[0].Age.Warn.Duration, warn)
	assert.Equal(t, c.Backups[0].Age.Error.Duration, err)
}

func TestLoadConfigWithSomeDefaults(t *testing.T) {
	reader := strings.NewReader(`{
		"defaults": {
			"repositories": {
				"handler": {
					"check_success": "notify_send -t 1000 check done",
					"forget_success": "notify_send -t 1000 forget done"
				},
				"policy": {
					"keep_last": 5,
					"keep_hourly": 6,
					"keep_daily": 7,
					"keep_weekly": 8,
					"keep_monthly": 9,
					"keep_yearly": 10,
					"keep_within": "2y5m7d",
					"keep_tags": [ "data", "test" ]
				},
				"limit_download": 1024,
				"limit_upload": 4096
			},
			"backups": {
				"age": {
					"error": "1h30m"
				},
				"handler": { 
					"success": "notify_send -t 5000 SUCCESS \"{{.BackupName}} has been backed up\""
				}
			}
		},
		"repositories": [
			{
				"name": "test1",
				"url": "/home/test/repos/test1",
				"password": "1"				
			},
			{
				"name": "test2",
				"url": "/home/test/repos/test2",
				"password": "2",
				"limit_download": 2048,
				"limit_upload": 8192
			}
		],
		"backups": [
			{
				"name": "mysql",
				"repositories": [ "test1" ],
				"data": [ "/etc/" ],
				"tags": [ "db", "mysql" ],
				"handler": {
					"age_warn": "/warn/age"
				}
			}
		]
	}`)

	c, error := LoadFromReader(reader)
	assert.True(t, error == nil)

	assert.Equal(t, c.Defaults.Repositories.Handler.CheckSuccess, "notify_send -t 1000 check done")
	assert.Equal(t, 2, len(c.Repositories))
	assert.Equal(t, c.Repositories[0].Handler.CheckSuccess, "notify_send -t 1000 check done")
	assert.Equal(t, c.Repositories[1].Handler.CheckSuccess, "notify_send -t 1000 check done")
	assert.Equal(t, c.Repositories[0].Handler.CheckFailure, "")
	assert.Equal(t, c.Repositories[0].Handler.ForgetSuccess, "notify_send -t 1000 forget done")
	assert.Equal(t, c.Repositories[0].Handler.ForgetFailure, "")
	assert.Equal(t, c.Repositories[0].Policy.KeepLast, uint(5))
	assert.Equal(t, c.Repositories[0].Policy.KeepHourly, uint(6))
	assert.Equal(t, c.Repositories[0].Policy.KeepDaily, uint(7))
	assert.Equal(t, c.Repositories[0].Policy.KeepWeekly, uint(8))
	assert.Equal(t, c.Repositories[0].Policy.KeepMonthly, uint(9))
	assert.Equal(t, c.Repositories[0].Policy.KeepYearly, uint(10))
	assert.Equal(t, c.Repositories[0].Policy.KeepWithin, "2y5m7d")
	assert.Equal(t, c.Repositories[0].Policy.KeepTags, []string{"data", "test"})
	assert.Equal(t, c.Repositories[0].LimitDownload, 1024)
	assert.Equal(t, c.Repositories[0].LimitUpload, 4096)
	assert.Equal(t, c.Repositories[1].LimitDownload, 2048)
	assert.Equal(t, c.Repositories[1].LimitUpload, 8192)

	assert.Equal(t, c.Defaults.Backups.Handler.Success, "notify_send -t 5000 SUCCESS \"{{.BackupName}} has been backed up\"")
	assert.Equal(t, 1, len(c.Backups))
	assert.Equal(t, c.Backups[0].Handler.Before, "")
	assert.Equal(t, c.Backups[0].Handler.Success, "notify_send -t 5000 SUCCESS \"{{.BackupName}} has been backed up\"")
	assert.Equal(t, c.Backups[0].Handler.Failure, "")
	assert.Equal(t, c.Backups[0].Handler.After, "")
	assert.Equal(t, c.Backups[0].Handler.AgeWarn, "/warn/age")
	assert.Equal(t, c.Backups[0].Handler.AgeError, "")

	err, _ := time.ParseDuration("1h30m")
	assert.Equal(t, c.Backups[0].Age.Warn.Duration, time.Duration(0))
	assert.Equal(t, c.Backups[0].Age.Error.Duration, err)
}
