# Go Players Data

`go-players-data` is a serverless application written in Go, designed to run as a Yandex Cloud Function. 

It fetches player data from an external API, filters offline players based on configurable criteria, groups them by store number, and sends email notifications using SMTP. The function supports both timer-based triggers (e.g., daily runs) and HTTP triggers.

## Features
- Fetches player data from a configurable API endpoint.
- Filters players by offline duration, group, and company.
- Groups players by store number for clustered reporting.
- Sends email notifications in parallel using customizable templates.
- Logs execution details for monitoring and debugging.
- Deployable to Yandex Cloud with a timer trigger for scheduled runs.

## Project Structure
```
go-players-data/
├── cmd/              # Local entry point for testing
│   └── main.go
├── internal/         # Internal packages
│   ├── cluster/      # Groups players by store number
│   ├── config/       # Loads configuration from env vars or .env
│   ├── fetcher/      # Fetches data from an external API
│   ├── filter/       # Filters players based on criteria
│   ├── logger/       # Logging utility using zerolog
│   ├── mailer/       # Sends email notifications via SMTP
│   ├── model/        # Defines player data structures
│   ├── player/       # Parses raw JSON into player structs
│   └── templateloader/ # Loads and renders email templates
├── templates/        # Email template files
│   └── byStore.tmpl
├── handler.go        # Yandex Cloud Function entry point
├── go.mod            # Go module definition
├── go.sum            # Go dependencies checksums
└── Makefile          # Build and deployment automation
```

## Prerequisites
- [Go](https://golang.org/dl/) 1.21 or later
- [Yandex Cloud CLI (`yc`)](https://cloud.yandex.com/docs/cli/quickstart) installed and configured
- A Yandex Cloud account with a service account having `serverless.functions.invoker` and the other necessary roles
- SMTP server credentials for sending emails
- Access to an API providing player data (URL and API key)

## Installation
1. Clone the repository:
   ```bash
   git clone https://github.com/rptl8sr/go-players-data.git
   cd go-players-data
   ```

2. Install dependencies:
    ```bash
    go mod tidy
    ```

## Configuration

Copy `dumb.env.prod` to `.env.prod` and fill it with you values

```dotenv
# Application settings
APP_VERSION=0.0.1 # Optional
APP_MODE=prod          # "dev" or "prod"
APP_LOG_LEVEL=info     # Log level: debug, info, warn, error
APP_MAX_GOROUTINES=10  # Max concurrent goroutines for email sending

# Mailer
MAIL_FROM=email@domain.com # Email sender
MAIL_HOST=smtp.domain.com  # Email host
MAIL_PASSWORD=email_password # Email sender password
MAIL_PORT=12345 # Email port
MAIL_TO=receiver01@domain.com,receiver02@domain.com # Comma separated email recepients
MAIL_SUBJECT=Any email subject # Email subject
MAIL_TEMPLATE_NAME=byStore # Template for email
MAIL_STORES=1111:store01@domain.com,22222:store02@domain.com # Optional. Mapping storeNumbers with its email

# Data source settings
DATA_URL=https://api.example.com/players # Data source
DATA_API_KEY=your-api-key # Data source API key
DATA_COMPANIES=shortName:fullCompanyName,sn:fsn # Comma separated companies names maping. See the parser.parseTags and the filter.stringInSlice
DATA_IGNORED_GROUPS=group1,group2 # Comma separated ignored groups for filtering. See the model.Player and the filter.Filter 
DATA_ALLOWED_COMPANIES=company1,company2 # Comma separated allowed companies for filtering. See the model.Player and the filter.Filter
DATA_MAX_OFFLINE=24    # Max offline time in hours
DATA_STORE_TEST_NUMBER=0000 # Ignoring testing store number
DATA_STORE_NUMBER_PREFIX=STORE_ # Store tag prefix. See the parser.parseTags
DATA_COMPANY_NAME_PREFIX=LLC_ # Company name prefix. See the parser.parseTags

# Yandex Cloud
YC_SA_ID=abcdef1234 # Your Yandex Cloud service account ID
YC_CRON='0 0 ? * * *' # Cron to trigger bu timer
YC_FUNC_NAME=go-yc-func # Yandex Cloud Function name
```
### Notes

Yandex Cloud Function `CRON` contains **six** fields.
See the [documentation](https://yandex.cloud/en-ru/docs/functions/concepts/trigger/timer).


## Local Running
Run the function locally
```bash
  go run main.go
```

## Deployment to Yandex Cloud

The `Makefile` provides targets to deploy the function:

1. Set environment variables (optional): If not using `.env`, export variables:
```bash
    export SA_ID="your-service-account-id"
    export FN="your-function-name"
    export CRON="0 10 * * ?"  #Daily at 10:00 UTC
```

2. Deploy the function:
```bash
  make fn-deploy
```
Or if you aren’t set environment variables
```bash
  make fn-deploy FN="your-function-name" SA_ID="your-service-account-id" CRON="0 10 * * ?"
```
This:
- Creates the function if it doesn't exist.
- Zip the source code.
- Deploys a new version with environment variables.
- Creates and sets up a timer trigger.
- Cleans up the zip file.
  
3. Verify deployment: List functions:
```bash
  make fn-version-list
```

4. Check logs:
```bash
  yc serverless function logs your-function-name
```

## Usage

- Timer Trigger: The function runs daily at 7:20 UTC (configurable via CRON), fetching data and sending emails.
- HTTP Trigger: Invoke manually via the function's HTTP endpoint (returned by yc serverless function get):

```bash
    curl https://functions.yandexcloud.net/<your-function-id>
```

## Makefile Targets
- fn-create: Creates the function if it doesn't exist.
- fn-zip: Creates a zip archive of the source code.
- fn-create-version: Deploys a new function version.
- fn-timer: Sets up a timer trigger.
- fn-clear: Removes the zip file.
- fn-deploy: Full deployment workflow.
- fn-list: Lists all functions.
- fn-version-list: Lists function versions.