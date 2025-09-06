# dlqt
 
A really cool Go CLI tool for interacting with Azure Service Bus queues
 
## Usage
 
### `dlqt`
 
- developer tool for re-submitting dead-letter messages
- uses auth service for secure retriggering
- run `dlqt -h` for usage information
 
### `dlqtools`
 
- admin tool for managing queues and dead-letter queues
- uses `az login` for authentication
- has direct access to Service Bus for admin operations
- run `dlqtools -h` for usage information

### `auth`

- HTTP API service for authenticated DLQ message retriggering
- runs in Azure Container Apps with managed identity
- authenticates users via Azure AD tokens
- provides fine-grained access control for message retriggering
 
## Architecture
 
The system consists of:
1. `dlqt` - Developer CLI tool for secure message retriggering
2. `dlqtools` - Admin CLI tool with direct Service Bus access
3. `auth` - Containerized API service for secure message retriggering
4. Azure Service Bus with RBAC for the auth service
 
**Developer Workflow:**
- Developers use `dlqt retrigger` which calls the `auth` API with their Azure AD token
- The auth service validates the token and performs the retrigger operation using its managed identity
- Developers cannot modify message contents, only retrigger
 
**Admin Workflow:**
- Admins use `dlqtools` with direct Service Bus access for full queue management
- Includes seed, purge, and other admin operations
 
## Build
 
### `dlqt`
 
local dev
```bash
go install ./cmd/dlqt && source <(dlqt completion zsh)
which dlqt
dlqt -h
```
 
shipping
```bash
CGO_ENABLED=0 go build -ldflags="-s -w" ./cmd/dlqt
```
 
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

### `auth`

```bash
cd auth
docker build -t dlqt/auth .
```

## Deployment

1. Deploy infrastructure:
```bash
cd infra
terraform init
terraform plan
terraform apply
```

2. Build and push the auth service container:
```bash
cd auth
docker build -t dlqt/auth .
docker tag dlqt/auth <your-registry>/dlqt/auth:latest
docker push <your-registry>/dlqt/auth:latest
```

3. Update the container app image in Terraform and redeploy

## To Do
 
### `dlqt`
 
- auth API integration completed ✅
- add more developer-friendly features
 
### `dlqtools`
 
- purge dead-letter queue
