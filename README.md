# ghostorag_client
The client is an amateur CLI REPL app. It literally uses a while loop to read user input and processes commands accordingly.

## How Encryption/Decryption works
Encryption and decryption of files in FileGhost are handled using AES-256-GCM. Because this is a zero-trust architecture, all cryptographic operations happen entirely on your local machine before the data ever touches the network. Here is a high-level overview of the implementation.

### Encryption
- The user provides a password. The client generates a cryptographically secure 16-byte random Salt. The password and salt are fed into the Argon2id memory-hard algorithm to generate a 32-byte (256-bit) AES key.
- The client generates a random 12-byte Nonce (Number Used Once) to ensure every file encryption is entirely unique. GCM takes this nonce and attaches a 4-byte counter to it (starting at 1).
- For every 16-byte block of the file, GCM increments that counter by 1. The AES cipher encrypts this unique Nonce+Counter combo to create random "noise," which is then securely XOR'd against the plaintext file to create the ciphertext. 
- While the file is being encrypted, GCM continuously feeds the resulting ciphertext into a mathematical function (Galois field multiplication). It acts like a rolling checksum. When the file is finished, this rolling calculation is finalized into a 16-byte Message Authentication Code (MAC) and attached to the very end.

The encrypted file looks something like this:
```bash
[16 byte salt][12 byte nonce][body (ciphertext + 12 byte MAC)]
```

### Decryption
- The user provides their password. The client slices the 16-byte salt and the 12-byte nonce off the front of the encrypted file.
- Argon2id uses the provided password and the extracted salt to regenerate the exact same 32-byte AES key.
- GCM is initialized using the derived key and the extracted nonce. As it decrypts the ciphertext back into plaintext, it simultaneously calculates its own rolling MAC.
- Once finished, GCM compares its mathematically calculated MAC against the 16-byte MAC attached to the end of the file. If they do not match perfectly, decryption is reported as a failure. This guarantees that either the password was incorrect, or the encrypted file was tampered with in transit.

### Hardware Limitations
Go's internal structure for the Galois Counter Method operations (`gcm.Open` and `gcm.Seal`) require a file to be physically present in the RAM to contruct the authentication tag.\
So, say you want to encrypt/decrypt a file, Go will load that entire file in RAM, and then process it. The output will also remain on RAM before being written on the disk. Therefore, to process a file, you need at least twice the amount filesize available on your RAM.\
This is a major bottleneck on the client end as your computer probably cannot demand 50 GB RAM to encrypt or decrypt a 25 GB file in a blink.

## How Tokens are Stored
When the client app boots, it intializes two empty strings for username and token. The user is expected to log in before doing anything. Logging in generates a token in the server which is sent back as the HTTP response from the server. Thus, the username and token strings are populated with the logged in user's username, and current token.\
This token is sent with every command (with a couple of exceptions like pinging the server or checking global storage) either as an HTTP header or in the body itself. Before the server does anything, it verifies that the user's token is valid. If it isn't, the server returns an error. If it is, the server generates a new token and sends it back to the client along with the response. The client reads the new token and updates its internal current_token variable.

## How to change the Server URL
By default, the client app tries to communicate with `http://localhost:2050`.\
The server URL can be changed by using the REPL command `/server`.\
You can also hardcode a different default server URL by changing the `cli.go` file by changing the following line:
```bash
var serverURL = "http://localhost:2050/"
```


## Recommended Way to Close the App
It is recommended to use the CLI's `/exit` command to close the client app. This command triggers automatic logout and hence ensures that the user's entry in the database doesn't contain a stray token.\
While I did trap SIGINT (`ctrl+c` or anything similar to close the app) to trigger a logout, it works inconsistently.