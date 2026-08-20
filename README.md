# ghostorag_
**Zero-Knowledge, Network-based Cloud Storage**\
ghostorag_ is a self-hosted cloud storage system designed with some serious data privacy in mind. It allows you to host your own remote file locker over the internet or a local network, where the server knows absolutely nothing about the files it holds.

## A Tale of Two Cities
This project has two main systems at play. You can visit the branches in this repository.
#### The Server
Acts purely as a "blind" storage locker and network router. It handles database management, disk I/O, and network streaming, but it possesses zero cryptographic intelligence. It uses an SQLite database to store users and files. The app is quite lightweight and can be comfortably run on low-power devices such as a Raspberry Pi.
#### The Client
This is a CLI app that allows a user to communicate with the server. It handles the entire cryptographic engine locally and serves as the user's interface for uploading or downloading encrypted files.

## The Zero-Knowledge Architecture
ghostorag_ operates on a zero-trust model. The server never sees your password or your encryption keys, and it has no access to your raw files.
- All files are encrypted locally using AES-256-GCM before upload.
- If a malicious actor compromises the server and steals the SQLite database, everything is essentially useless without the encryption keys.

## Rolling Token Authentication
To protect against Man-in-the-Middle (MitM) and Replay Attacks, I chose not to use static API keys in favour of a Rolling Token mechanism.
- Upon logging in, the server issues a session token, which is stored in the database.
- Every time the user makes a request (like an upload or download), the server checks if the token from the request and the one stored in the database are the same. If yes, a new token is generated, the database token is updated, and the server response includes the updated token, which the client app on the backend extracts. If the tokens don't match, a `401: Unauthorised` error is raised.
- If a network sniffer intercepts your token and attempts to replay it, it would be met with a `401: Unauthorised` response as the token would've already been invalidated (assuming your request successfully updated the token).
- If an attacker breaches the database and finds the correct token, they can perform actions as the user. But then the user would probably notice that their requests aren't going through, and they'll be prompted to log in again, which resets their token.
- Logging out resets the token value in the database to NULL.

## Setup Instructions

### Software Requirements
- Go 1.2x; remember to update go.mod file accordingly.
- SQLite 3.x (server only)

#### Client
```bash
git clone --depth 1 -b client --single-branch https://github.com/SupercalifragilisticexpialidociousPoppy/fileghost
```

#### Server
```bash
git clone --depth 1 -b server --single-branch https://github.com/SupercalifragilisticexpialidociousPoppy/fileghost
```

### Minimum Hardware Requirements
- Computer turns on.

### Maximum Hardware Requirements
- [RAM double the size of your SSD](https://github.com/SupercalifragilisticexpialidociousPoppy/fileghost/tree/client#hardware-limitations) (client only)
