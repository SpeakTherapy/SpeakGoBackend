# PeakSpeak Golang Backend

This is the backend service for **PeakSpeak**, a speech therapy platform. The backend is built using Go and provides APIs for video recording, encryption, user authentication, and more. It integrates with AWS KMS, DigitalOcean Spaces, MongoDB, and other services.

## Table of Contents

- [Features](#features)
- [Technologies](#technologies)
- [Requirements](#requirements)
- [Installation](#installation)
- [Configuration](#configuration)
- [Running the Application](#running-the-application)
- [API Documentation](#api-documentation)
- [Encryption Workflow](#encryption-workflow)
- [Contributing](#contributing)
- [License](#license)

## Features

- **User Management**: Provides APIs for user authentication, role management (therapist, patient), and profile updates.
- **Video Recording and Uploading**: Upload encrypted video recordings to DigitalOcean Spaces.
- **Encryption**: Uses AWS KMS for key management and AES-GCM encryption for video files.
- **Presigned URLs**: Generate presigned URLs for securely uploading and downloading videos.
- **HIPAA Compliance**: Ensures encryption at rest and in transit, access controls, and secure key management for sensitive data.

## Technologies

- **Golang**: The primary language for backend development.
- **MongoDB**: NoSQL database for storing user and exercise data.
- **AWS KMS**: For key management and encryption.
- **DigitalOcean Spaces (S3-compatible)**: Object storage for storing video files.
- **Gin**: Lightweight web framework for API development.

## Requirements

- Go 1.18+
- MongoDB
- AWS KMS
- DigitalOcean Spaces account

## Installation

1. **Clone the Repository**:

   ```bash
   git clone https://github.com/your-username/peakspeak-backend.git
   cd peakspeak-backend
   ```

2. **Install Dependencies**:
   ```bash
   go mod download
   ```
   
3. **Set Up MongoDB**
   Install and configure MongoDB. The default configuration connects to ```mongodb://localhost:27017```. You can update the connection string in the .env file.

## Configuration

1. **Environment Variables**
   Create a ```.env``` file at the root of the project and provide the following environment variables:

   ```bash
   MONGO_URI=mongodb://localhost:27017/peakspeak
   SPACES_KEY=your_digitalocean_spaces_key
   SPACES_SECRET=your_digitalocean_spaces_secret
   AWS_ACCESS=your_aws_access_key
   AWS_SECRET=your_aws_secret_key
   KMS_KEY_ID=your_kms_key_id
   SPACES_ENDPOINT=https://nyc3.digitaloceanspaces.com
   ```
2. **AWS KMS**:

   Ensure your AWS KMS key is set up and has the necessary permissions to encrypt and decrypt keys.
   
3. **DigitalOcean Spaces**:
   
  Make sure your DigitalOcean Spaces bucket is set up and that the appropriate permissions are applied.

## Running the Application

### Locally:

To run the application locally:

```bash
go run main.go
```
The server will start on [http://localhost:8080](http://localhost:8080)

## API Documentation

### Base URL
- **Local**: [http://localhost:8080](http://localhost:8080)

### Authentication
- **POST** `/auth/login`: Login with username and password.
- **POST** `/auth/register`: Register a new user.

### Video Upload
- **POST** `/getuploadurl/:patient_exercise_id`: Get a presigned URL for uploading an encrypted video.
- **GET** `/getdownloadurl/:patient_exercise_id`: Get a presigned URL for downloading a video.

### User Management
- **GET** `/users/:id`: Get user details.
- **PUT** `/users/:id`: Update user details.

---

## Encryption Workflow

### Video Upload:
1. The client records a video.
2. The video is encrypted on the client-side using an AES-GCM symmetric key.
3. The AES key is encrypted using AWS KMS.
4. The client uploads the encrypted video to DigitalOcean Spaces using a presigned URL.

### Video Download:
1. The client requests a presigned URL to download the video.
2. The AES key is unwrapped using AWS KMS.
3. The client downloads and decrypts the video locally.

This ensures that the video data is encrypted both at rest and in transit, adhering to HIPAA compliance standards.

---

## Contributing

Contributions are welcome! Please open an issue or submit a pull request for any bug fixes, new features, or improvements.
