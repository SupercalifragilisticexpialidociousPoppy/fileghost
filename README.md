# fileghost_server
This branch contains the server code, which defines how the database communicates with the client app.

## Hosting
By default, the server starts running on `http:localhost:2050`, any device on the same network as the server can use this to communicate just fine. However, for internet hosting, you can use a service like CloudFlare for maximum security and reliability. I, however, am broke and will hence resort to a free Pinggy + Discord combo.

##### There are two prerequisites before you compile:

### 1. ed25519 Authentication Key
- We will be using Pinggy for hosting our server, but an account isn't really required, we'll instead use an SSH key pair for authentication.
- The key pair can be made by using the following command:
#### Linux/Mac
```bash
ssh-keygen -t ed25519 -N '' -f "\$HOME/.ssh/pinggy_key"
```
#### Windows Powershell
```powershell
ssh-keygen -t ed25519 -N `"`" -f "\$HOME\.ssh\pinggy_key"
```

Powershell sometimes struggles with taking `"` in the command, we can use the interactive command for this:
```bash
ssh-keygen -t ed25519 -f "\$HOME/.ssh/pinggy_key"
```

When prompted, press **Enter** twice to leave the passphrase empty:
1. `Enter passphrase (empty for no passphrase):` ➔ Press **Enter**
2. `Enter same passphrase again:` ➔ Press **Enter**

- You can use a Pinggy Account to run the reverse SSH tunnelling for more robust hosting, but you'll need to update `pkg/tunnel/tunnel.go` accordingly.
- You can choose to name the key-pair anything else. `$HOME/.ssh/<anything>`, but `$HOME/.ssh/pinggy_key` is hardcoded in `pkg/tunnel/tunnel.go`; remember to update it.

### 2. Discord Webhook
- Follow this video and obtain the webhook URL: https://www.youtube.com/watch?v=CHmjprU4Zck
- Create a .env file in `cmd/.env`, you can change the location of the .env file but you'll need to change the location specified in `cmd/main.go` accordingly. Relative filepaths are rooted wherever your binary file is.
- Write the following in .env; ensure that there are no spaces.
```bash
discordWebhook=[Webhook URL which you copied.]
```

### How it works:
- Pinggy takes localhost:2050 and connects its own servers. It generates a randomized link which looks like: `https://^[a-zA-Z0-9-]+$.free.pinggy.net`. We can use this new link now to communicate with our server.
- This link can be shared in a public Discord channel (or something similar) so that the users can see it.

This is what `pkg\tunnel\tunnel.go` is doing. Once the server has started locally, `cmd/main.go` calls `tunnel.StartPinggyTunnel()` in `pkg\tunnel\tunnel.go`. The function extracts the Discord Webhook URL from the .env file, extracts the SSH Key from the local machine, runs the pinggy command:
```bash
ssh -T -p 443 -i `keyPath` -R0:localhost:2050 -o StrictHostKeyChecking=no -o ServerAliveInterval=30 +text@free.pinggy.io
```
If you run this command manually, the generated links will be printed in the terminal, but since another process if running the command as a background process, the output would remain invisible; hence we use Stdout pipes to read the output of the command, build a string from the output and then relay it to our Discord Webhook.

#### If you don't want to use this method:
Comment out these lines in `cmd/main.go`:
```bash
	fmt.Println("[    SERVER     ] Attempting pinggy tunnnelling...")
	tunnel.StartPinggyTunnel(port)
	defer tunnel.StopPinggyTunnel()
```
Now, the server will run only on `http://localhost:2050`. You can additionally set up an alternative tunnelling service like CloudFlare on your own.

## Storage
All of the uploaded files are stored in a directory named `Storage`. This directory is created wherever the binary resides.\
You can change the directory address in `pkg/db/db.go`. Find this line: 
```bash
	os.MkdirAll("storage", 0755)
```

### Soft Storage Cap
To not fill up a server with random files, I decided to add a soft upper limit to my storage of 16 GB.\
Occupied file storage is determined by adding the sizes of all files present in the `file` table of the database.\
Whenever somebody sends an upload request, the server checks if filesize <= total - occupied space before commencing upload.

You can alter this upper limit by changing the line in `pdk/db/dg.go`.
```bash
const MaxStorageBytes int64 = 16 * 1024 * 1024 * 1024
```

It is also advised to physically partition your server drive.

## Uploads and Downloads
Uploads and downloads use similar logic.
### How an Upload Works
Because FileGhost is designed to run on memory-constrained hardware, it cannot load uploaded files into RAM. Instead, it relies on a two-phase stream reading mechanism to pipe network data directly to the disk.

##### Phase 1: The Header Intercept (The Trigger)
When a client initiates an upload, the data arrives at the server as a continuous stream of TCP packets. The Go HTTP server intercepts the very first packets and reads only the HTTP Headers (which contain the `X-Session-Token` and `Content-Length`). The moment it detects the end of the headers, Go intentionally pauses the network read and triggers the `db.ProcessUpload()` function. At this stage, the actual file data is still waiting in the operating system's network buffer.

##### Phase 2: The Direct Disk Pipeline
Once ProcessUpload() verifies the security token and confirms the server has enough storage capacity, it opens a secure file on the hard drive. It then executes a single io.Copy command. This acts as a valve, instructing the server to resume reading the rest of the incoming TCP packets. Instead of loading the packets into the server's main memory, Go scoops the data in tiny 32KB chunks and streams it directly to the physical disk until the transfer is complete.

- This keeps the process of reading and writing take ~ 32 KB RAM on the client and the server side, despite the actual filesize being much larger.
- The token is only shared in the HTTP headers of the intial TCP packets. It doesn't roll for each packet. This is mainly because I don't know how to do it and even if I did, it would drop performance significantly. (Also, tokens not updating in the database and the upload/download stopping in the middle of an already slow 5 GB file transfer would be quite annoying.)

### How a Download Works
Just swap client and server in the explanation above while keeping token rolling as it is. That's pretty much it.

### Uploads and Downloads cannot be initiated from a browser.
Typing `http://localhost:2050/download/<fileId>` would return a `401: Unauthorized`. This is because both upload and download processes put the user's token as a custom header of the HTTP request. Browsers don't support custom headers, they use cookies to store your token. This is why the user requires a client app.

## How Users are stored
Each user has a username, a password, a token and an auto-incrementing integer which features as the primary key.\
- The username only checks if it's unique. It's not necessary to use an email.
- The password is hashed using standard Go library. There is no way to recover a password.
- This token is updated with every command using the rolling token architecture.

## How Files are stored
Each file has a string ID (primary key), a foreign key which determines the owner ID, the file path of the file stored in the server and its size.
- It also contains a public or private tag which I intend to implement in the future, so as to support downloads without authentication on public files.
- The name of an uploaded file isn't actually stored in the server (zero-knowledge); the only way to identify a file is via it's randomly generated string ID.
- While this should be apparent, the password used for encrypting the file isn't stored in the server either. It isn't even sent over the network or stored in the client app. There's no way to decrypt a file if its password is lost.
- The server doesn't really identify if an uploaded file is actually encrypted or not. You can technically upload anything... but that defeats the philosophy of this project.