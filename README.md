# dlqt
 
A really cool Go CLI tool for interacting with Azure Service Bus queues
 
## Usage
 
### `dlqt`
 
- developer tool for re-submitting dead-letter messages
 
### `dlqtools`
 
- admin tool for managing queues and dead-letter queues
- requires "Azure Service Bus Data Sender" PIM role
- uses `az login` for authentication
- run `dlqtools -h` for usage information
 
## Build
 
### `dlqt`
 
### `dlqtools`
 
local dev
```bash
go install ./cmd/dlqtools && source <(dlqtools completion zsh)
which dlqtools
dlqtools -h
```
 
shipping
```bash
# env var & flags reduces binary size
# execution time 0.005s, size 16MB
CGO_ENABLED=0 go build -ldflags="-s -w" ./cmd/dlqtools
 
# compresses binary but increases execution time
# good for container images, not so much for CLI binaries
# execution time 0.225s, size 3.3MB
upx --best --lzma ./dlqtools
```
 
## To Do
 
### `dlqt`
 
- auth API for developers
 
### `dlqtools`
 
- purge dead-letter queue
- support sending a YAML/JSON file of messages