[![Coverage Tests](https://github.com/jmsilvadev/zc/actions/workflows/automated_tests.yml/badge.svg)](https://github.com/jmsilvadev/zc/actions/workflows/automated_tests.yml)
[![E2E Tests](https://github.com/jmsilvadev/zc/actions/workflows/e2e_tests.yml/badge.svg)](https://github.com/jmsilvadev/zc/actions/workflows/e2e_tests.yml)
[![Quality](https://github.com/jmsilvadev/zc/actions/workflows/quality.yml/badge.svg)](https://github.com/jmsilvadev/zc/actions/workflows/quality.yml)

# Secure File Storage and Retrieval System

This project involves developing a secure file storage and retrieval system using a client-server architecture enhanced with Merkle tree verification. The client can upload files to a server, delete the local copies, and later download any specific file with a guarantee of its integrity.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Usage](#usage)
- [Commands](#commands)

## Prerequisites

Before you begin, ensure you have the following installed on your system:

- Docker: [Install Docker](https://docs.docker.com/get-docker/)
- Docker Compose: [Install Docker Compose](https://docs.docker.com/compose/install/)
- Make: [Install Make](https://www.gnu.org/software/make/)

## Installation

1. Clone the repository to your local machine:

   ```sh
   git clone https://github.com/jmsilvadev/zc.git
   cd zc
   ```

2. Build and start the Docker containers:

   ```sh
   make up-build
   ```

   This command will:

   - Build the Docker images specified in the `docker-compose.yml` file.
   - Start the containers using Docker Compose.

3. Build the client application:

   ```sh
   make build-client
   ```

   This command will:

   - Compile the client application and prepare it for use.

## Usage

After following the installation steps, your client and server should be up and running. You can now use the system to upload, store, and verify files.


```
Usage of bin/zc-cli:
  -config-dir string
    	Directory to store rootHash and downloaded files (default "/home/jmsilvadev/.zc")
  -delete
    	If the client can delete the local files after the upload (default true)
  -dir string
    	Directory containing files for upload
  -files string
    	Comma-separated list of files for upload
  -host string
    	Server host (default "http://localhost:5000")
  -index int
    	Index of the file to download (default -1)
  -operation string
    	Operation to perform: upload, update or download. Attention: perform an upload will always remove the existent data (default "upload")

```

### Examples

To upload files:

```
bin/zc-cli -operation upload -files ./file2.txt,./file1.txt,./file3.txt,./file4.txt
```

To update or add more files without delete the existents:

```
bin/zc-cli -operation update -files ./file5.txt
```

To dowanload a i-th file:

```
bin/zc-cli -operation download -index 2
```

## Running Tests

To ensure everything is working correctly, you can run the provided tests. Use the following command:

```sh
make tests
```

This command will:

- Execute the test suite to verify the functionality of both the client and server components.
- Report any issues or failures detected during the testing process.


## Commands

```
build-client                   Build server component
build-image                    Build docker image in daemon mode
build-server                   Build server component
clean                          Clean all builts
clean-tests                    Clean tests
down                           Stop docker container
logs                           Watch docker log files
tests-client                   Run unit tests ion the client
tests-coverage                 Run all tests with coverage in html
tests-pkg-cover                Run package tests with coverage
tests-pkg                      Run package tests
tests                          Run all tests
tests-server                   Run unit tests in the server
up-build                       Start docker container and rebuild the image
up                             Start docker container

```

