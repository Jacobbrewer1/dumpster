# Dumpster

Dumpster is a GO CLI tool that makes creating MySQL dumps easy. The application can upload the dump to a Google Cloud
Storage bucket (S3 support is planned). There is a docker image available and attached to this repository under the
packages section. 

We are planning to add more features to this tool, so stay tuned.

## Installation

You can install the tool using the following command:

```bash
go get -u github.com/Jacobbrewer1/dumpster
```

## Usage

### Locally

The tool is very simple to use. You can run the following command to see the available options:

```bash
dumpster commands
```

### Docker

There is a docker image available for this tool. I personally use the docker image to run the tool as a Kubernetes 
cronjob in my projects.

## Commands

The following commands are available:

- `version` - This command will display the version of the tool.
- `dump` - This command will create a dump of the specified database and upload it to the specified bucket.
- `purge` - This command will delete all the files in the specified bucket.

## Configuration

The tool requires a small setup if certain features are to be used. you can run the following command to get help on
configuring the tool:

```bash
dumpster <command> --help
```
