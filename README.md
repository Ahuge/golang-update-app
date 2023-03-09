# golang-update-app
This is an example golang app that will utilize `github.com/mouuff/go-rocket-update` to run automatic updates of the service.

This adds on to the `go-rocket-update` library in the following ways:
- Implements an example AWS S3 provider
- Implements logic for PKI signature verification of the downloaded binary


## How does it work
- You will build your application, bundle it up however you'd like (zip)
- `go run keygen/main.go -generate`
- Replace the `pemData` in `verify.go` with the contents of your `client.pem`
- `go run keygen/main.go -sign <zipPath>`
- Update the `VERSION` file with the version number of your built app and the signature that came from `keygen`
- Upload both the zip bundle and the `VERSION` file to an S3 folder represented by your `AmazonS3` provider config (main.go)

When the app starts, it will look at your latest folder, pull the `VERSION` blob to check for the latest version.
If that version differs than what we are, it will pull that blob down and attempt to verify its server signature against the public key embeded in the application.

![An example demonstration showing it securely update](.github/gif.gif)