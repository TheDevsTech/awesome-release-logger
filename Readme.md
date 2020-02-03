# :clipboard:Awesome release logger:clipboard:
A handy tool that help to write release/change log from git commit message

## Development

- **Requirement**
    - Go >= 1.13
 
- Clone the repo
    ```
    git clone https://github.com/TheDevsTech/awesome-release-logger.git
    cd awesome-release-logger
    ```
- **Build**
    `make` or `go build -o bin/ar-logger main.go`

## Cross Platform Build
- Cross compilation is hard, and docker is help us in that way! Install docker and pull
    docker image `docker pull karalabe/xgo-latest` and install a go package `go get github.com/karalabe/xgo`
- For build most of the platforms binary use 
    ```
    xgo github.com/TheDevsTech/awesome-release-logger
    ```
- Or specific platform
    ```
    xgo --targets=linux/amd64 github.com/TheDevsTech/awesome-release-logger
    ```
- After build is finished you should have all platforms binary in your
current directory.
- More build details find [here](https://github.com/karalabe/xgo)

## Installation
- after build or download the binary just move it to `/usr/local/bin` i.e: `mv bin/ar-logger /usr/local/bin/arl`
- restart your shell
- verify installation `which arl` output should be `/usr/local/bin/arl`

## Usage
- type `arl -h` in terminal for help. You should get below output
    ```bash
    Usage of arl:
      -b	get logs from the beginning
      -d string
            project directory path (default ".")
      -n	write new release log file
      -o string
            output file path (default ".")
    ```
- Generate release log
    - go to your project directory and type `arl` there
    - Or you can run `arl` command from anywhere and point your project directory like this `arl -d <project_directory_path`
    - By default `release-log.md` file will generate in current directory. If you want it to other location then
    `arl -o <output_directory_path`
    - By default this logger prepend logs if file exists. If you need separate file just pass `-n` flag 
    i.e: `arl -n`
    - By default this logger get commit logs between latest tag & HEAD(if tags aren't exists then its get from the beginning)
    - If tags are exists and you want to generate a fresh log from the beginning then  ad `-b` flag
    i.e `arl -b`

# License
[GPL-3.0](https://github.com/TheDevsTech/awesome-release-logger/blob/master/LICENSE)
