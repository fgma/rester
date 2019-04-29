Introduction
------------

rester is a wrapper around `restic <https://restic.net/>`. Backups are setup using a configuration file. Executing a backup may be initiated with a single command manually or automatically e.g. using cron or systemd.

- single configuration file
- backup files or command output from stdin
- run manually or scheduled
- support for all restic backends
- integrate custom scripts using handlers


System requirements
-------------------

Rester is implemented in golang and should run on all platforms restic supports. It only needs a recent version of restic.

Right now it is only tested running on linux.

Install
-------

If you've got a go development environment setup already just clone the repository and run

.. code-block:: shell

   go build

inside the cloned repository. Otherwise check https://golang.org on how to setup your development environment. For debian based system you may also build a .deb running

.. code-block:: shell

   dpkg-buildpackage -rfakeroot

inside the cloned repository.

How to use
----------

Before running a backup a configuration file has to be created. The easiest way is to adjust the example configuration given below or run

.. code-block:: shell

    rester example-config

to print a minimal example configuration to stdout. For details about the configuration have a look at the configuration_ section.

After creating the configuration file you can start using restic. To check your configuration for problems run

.. code-block:: shell

    rester repos

or

.. code-block:: shell

    rester backups

This will parse your configuration and show your configured repositories and backups. Before you run your first backup make sure your repository is prepared. For local backups make sure the repository folder exists. For S3 ensure the bucket and user exist. You don't need to manually initialize the restic repository. You can use rester's init command to do so:

.. code-block:: shell

    rester init my-configured-repo1 my-configured-repo2 ...

To initialize all configured repositories just run:

.. code-block:: shell

    rester init

Most commands that work on either repositories or backups will work that way. If you specify no repository or backup rester will just consider all configured repositories or backups. If your backup repository is already set up you can skip the initialization and start to run backups:

.. code-block:: shell

    rester backup

To check your repositories for problems run:

.. code-block:: shell

    rester check

If everything is ok the command will exit without any output or error status. If you run

.. code-block:: shell

    rester snapshots

you should see your new backup(s). To get rid of old backups you can specify a policy which backups to keep when running. For details on how to specify the policy have a look at repositories_. To actually forget old backups run:

.. code-block:: shell

    rester forget

In addition to restic's forget command this will also run restic's prune command to actually free unused disk space. When running you backups regularly you might want to check the age of the last backup. Rester can do that for you according to the limits given in the backup configuration. You can specify a warning limit and an error limit for the age of the last backup. Run

.. code-block:: shell

    rester check-age

to check your backups ages. If you need to restore data you can use regular restic commands to do so or just mount a repository:

.. code-block:: shell

    mkdir mount-backup
    rester mount my-configured-backup mount-backup

If you want to run unsupported restic commands just run

.. code-block:: shell

    rester shell my-configured-backup

which will run a new shell prepared with restic's environment variables like repository, username, password etc. to run custom commands. After setting up and testing your backup configuration you may want to run your backup automatically from cron or systemd. To monitor your backups you can use different handlers that are executed on different events e.g. a failed backup or a backup age warning. Using these handlers you can integrate custom scripts to send you an email, send a desktop notification or integrate your backup status into a network monitoring system.

An overview of all available commands:

.. code-block:: shell

    $ rester
    A wrapper around restic for configuring and running backups

    Usage:
    ./rester [command]

    Available Commands:
    age            Show age of each backup
    backup         Run backups
    backups        Show configured backups
    check          Check configured repositories
    check-age      Check age of the given backups
    example-config Print an example configuration as a template
    forget         Forget backups in repositories according to policy
    help           Help about any command
    init           Initialize configured repositories using restic
    mount          Mount repostitory
    repos          List configured repositories
    shell          Start interative shell prepared with restic environment variables
    snapshots      List snapshots
    version        Print the version number

    Flags:
    -c, --config string   config file (default is $HOME/.config/rester/config.json)
    -h, --help            help for ./rester

    Use "./rester [command] --help" for more information about a command.
    $

.. _configuration:

Configuration
-------------

Rester is configured through a single configuration file. By default this file is located inside the users home directory under ``~/.config/rester/config.json`` ($XDG_CONFIG_HOME is respected if available). A different file may also be specified on the commandline using the ``--config`` option. This may be useful to run systemwide backups reading the config file from /etc/. In general most rester options map directly to the respective restic options.

On windows you can't create folders starting with a ``.`` using explorer. As a workaround you can create the config folder running

.. code-block:: shell

    md %USERPROFILE%\.config\rester

in the command prompt.

.. _repositories:

Repositories
============

To actually backup data at least one repository has to be configured. Rester supports all repository formats restic supports.

name
    A unique name to refer to this repository.

url
    The URL of the repository as passed to restic. For details on the format have a look at into restic's manual.

password
    The password of the repository.

environment
    Custom environment variables used when accessing the repository. This is used e.g. when accessing S3 storage to specify access keys. The environment variables are also available when rester calls handlers in the context of the repository. Therefore it is possible to add custom parameters for handler scripts.

policy
    The policy for keeping backups when running ``forget`` on the repository.

        keep_last
            Keep the last n backups.
        keep_hourly
            Keep n hourly backups.
        keep_daily
            Keep n daily backups.
        keep_weekly
            Keep n weekly backups.
        keep_monthly
            Keep n monthly backups.
        keep_yearly
            Keep n yearly backups.
        keep_within
            Keep backups within the given timespan. Given as string e.g. "7d12h".
        keep_tags
            Keep backups with the given tags.

check
    The parameters used when checking the repository:

        read_data_percentage
            An integer value between 0 and 100. Specifies the percentage of randomly choosen data in the repository that is checked for modifications on each run of check. If 100% is not an integer multiple of the given percentage the given percentage will be adjusted accordingly. E.g. a percentage of 50% will check half of the repository on each check while a percentage of 43% will only check 33% of the repository on each check.

handler
    Handlers are called at specific events during execution. They may be used to run custom scripts e.g. to notify the user about a successful check of the repository. 

        forget_success
            Run when ``forget`` command completed successful.
        forget_failure
            Run when ``forget`` command failed.
        check_success
            Run when ``check`` command completed successful.
        check_failure
            Run when ``check`` command failed.

    If the commands start with a ``~`` sign it is expanded to the user's home directory. Additionally some special variables inside the commands are replaced with the appropriate values to automatically customize commands:

        - {{.BackupName}}
        - {{.BackupRepository}}
        - {{.RepositoryName}}
        - {{.RepositoryURL}}

limit_download
    Limit the download rate to n KiB/s.

limit_upload
    Limit the upload rate to n KiB/s.

For more details have a look at the example_ configuration.

Backups
=======

name
    A unique name to refer to this backup.

repository
    The name of the repository to backup to as specified in the repositories section of the configuration.

data
    An array of files and directories to include in the backup. On windows you have to escape ``\`` characters inside a path using ``\\`` e.g. ``c:\\data\\pictures``.

data_stdin_command
    Backup the output of the given command instead of files. Mutually exclusive with ``data``. 

stdin_filename
    The filename of the stdin data inside the backup. Mandatory when using ``data_stdin_command``. 

exclude
    An array of files and directories to exclude from the backup.

one_file_system
    Boolean value that specifies if backups include mounted subfolders.

tags
    Tags for the backup.

environment
    Custom environment variables used when accessing the backup similar to the same variable in ``backups``.

custom_flags
    String array of custom flags that are not directly supported e.g. ``--ignore-inode``. All flags are directly passed to restic. Unsupported flags might break restic backups.

handler
    before
        Run before ``backup`` command.
	after
        Run after ``backup`` command independend of the result.
	success
        Run on success of ``backup`` command.
	failure
        Run on failure of ``backup`` command.
	age_warn
        Run if ``age-check`` command detects a backup age above the warn limit.
	age_error
        Run if ``age-check`` command detects a backup age above the error limit.

    For more details on handler usage have a look at the repository handler documentation.

age
    The age limits for a specific backup to be considered ok. Right now only units up to hours are supported for technical reasons:

    warn
        The warning limit as a string e.g. "12h30m".

    error
        The error limit as a string e.g. "48h".

For more details have a look at the example_ configuration.

Defaults
========

In more complex situations it is possible to specify default settings for all backups and repositories. A typical example might be handlers for notifications about the backup status. Currently only a subset of settings may be used in the defaults section. For repositories:

- handler
- policy
- limit_download
- limit_upload

For backups:

- handler
- age

For more details have a look at the example_ configuration.

Example configuration
=====================
.. _example:
.. code-block:: json

    {
        "defaults": {
            "repositories": {
                "handler": { 
                    "forget_success": "notify.sh SUCCESS \"{{.BackupName}} forget successful\"",
                    "forget_failure": "notify.sh FAILED \"{{.BackupName}} forget FAILED\"",
                    "check_success": "notify.sh SUCCESS \"{{.BackupName}} has been checked\"",
                    "check_failure": "notify.sh FAILED \"{{.BackupName}} check FAILED\""
                }
            },
            "backups": {
                "age": {
                    "warn": "1h30m",
                    "error": "3h"
                },
                "handler": { 
                    "before": "notify.sh START \"backing up {{.BackupName}}\"",
                    "success": "notify.sh SUCCESS \"{{.BackupName}} has been backed up\"",
                    "failure": "notify.sh FAILED \"{{.BackupName}} has NOT been backed up\"",
                    "age_warn": "notify.sh WARNING \"{{.BackupName}} backup to old\"",
                    "age_error": "notify.sh FAILED \"{{.BackupName}} has NOT been backed up in time\""
                }
            }
        },
        "repositories": [
            {
                "name": "minio-backup",
                "url": "s3:http://backups.example.com:9000/minio-backup",
                "password": "codqzkf30bcl1hz9",
                "environment": {
                    "AWS_ACCESS_KEY_ID": "odf4572yc147wd53",
                    "AWS_SECRET_ACCESS_KEY": "dt936p7clkp06ii4"
                },
                "policy": {
                    "keep_last": 5,
                    "keep_daily": 7,
                    "keep_weekly": 5,
                    "keep_monthly": 12,
                    "keep_yearly": 3
                },
                "check": {
                    "read_data_percentage": 5
                },
                "limit_download": 1024,
				"limit_upload": 4096
            }
        ],
        "backups": [
            {
                "name": "/home/user",
                "repository": "minio-backup",
                "data": [
                    "/home/user/"
                ],
                "exclude": [ 
                    ".cache/",
                    ".Trash/"
                ],
                "one_file_system": true,            
                "tags": [ "home", "data" ]
            },
            {
                "name": "crontab",
                "repository": "minio-backup",
                "data_stdin_command": "crontab -l",
                "stdin_filename": "crontab.txt",
                "one_file_system": true,            
                "tags": [ "cron" ]
            }
        ]
    }

