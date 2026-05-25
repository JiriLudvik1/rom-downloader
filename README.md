# ROM Downloader

A Go-based automated utility that listens for ROM upload events via Google Cloud Pub/Sub, downloads them from Google Cloud Storage (GCS), processes archives, and organizes them into console-specific directories.

## Features

- **Automated Downloads**: Listens to GCS upload events through Pub/Sub.
- **Archive Support**: Automatically extracts `.zip`, `.tar`, `.gz`, `.tgz`, and `.rar` files.
- **Smart Organization**: Routes ROMs to specific folders based on filename tags (e.g., `Game_NES.zip` goes to the NES folder).
- **Firestore Integration**: Tracks successful downloads and processing status in Google Cloud Firestore.
- **Resilient**: Graceful shutdown handling and error logging.

## Project Structure

- `client/`: The main Go application intended to run on a local machine or server (e.g., Raspberry Pi).
- `func-cleaner/`: Cloud Function for cleaning up resources.
- `func-dlq-handler/`: Cloud Function for handling Dead Letter Queue messages.
- `dummy-puller/`: Utility for testing Pub/Sub messaging.

## Setup

### Prerequisites

- A Google Cloud Project with Pub/Sub, GCS, and Firestore enabled.
- A Service Account with appropriate permissions and its JSON key file.
- Go 1.21+ installed.

### Configuration

1. Create a `client/config.json` based on `client/config_example.json`:
   ```json
   {
     "credentialsFileName": "service-account.json",
     "projectId": "your-project-id",
     "subscriptionName": "your-subscription-id",
     "topicName": "your-topic-id",
     "tempFolder": "./temp",
     "destinationFolderRoot": "/path/to/roms",
     "romTypeDestinations": {
       "NES": "nes",
       "SNES": "snes",
       "GBA": "gba"
     }
   }
   ```
2. Place your Google Cloud service account key in the `client/` directory (matching `credentialsFileName`).

### Running the Client

```bash
cd client
go run main.go
```

## How it Works (Tagging System)

The downloader uses a specific naming convention to identify the target console:
`{GameName}_{CONSOLE_TAG}.{extension}`

Example: `Super Mario Bros_NES.zip`
- The client detects the `_NES` tag.
- It downloads the zip.
- It extracts the content.
- it moves the extracted files to `{destinationFolderRoot}/nes/`.

## Deployment

The `client_scripts/build-raspberry.ps1` script is provided to help cross-compile the client for Raspberry Pi (ARM).

## License

[MIT](LICENSE) (or specify your license)
