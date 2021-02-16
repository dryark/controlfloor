# controlfloor
System for controlling devices remotely

# Basic Install Instructions
* Setup Control Floor  
  * `git clone https://github.com/nanoscopic/controlfloor.git`  
  * `cd controlfloor`  
  * `make`
  * `./main` ( starts Control Floor )
* Verify Control Floor is running
  * Open `http://localhost:8080` in your browser
  * Login with default hardcoded username and password 'ok' and 'ok'
* Setup ios_remote_provider
  * `git clone https://github.com/nanoscopic/ios_remote_provider.git`  
  * `cd ios_remote_provider`  
  * Login to Xcode with your paid developer account if you have not already  
  * `./util/signers.pl` ( to display your Developer Team OU )  
  * Update `config.json` to include your Developer Team OU  
  * `make wda`  
    * Downloads WebDriverAgent to repos/WebDriverAgent  
    * Reconfigures Xcode project for WDA according to your `config.json`  
    * Builds WDA using `xcodebuild built-for-testing`, placing the build into `repos/WebDriverAgent/build`. The prebuilt `xctestrun` file can then be seen in `bin/wda`  
    * Clones ios_video_app to `repos/vidapp`
  * `make`
    * Builds `main`, the main executable of the provider  
    * Builds `./bin/iosif`, a CLI tool for interacting with ios devices
* Install ios_video_app to your iOS device(s)
  * Open the Xcode project at `ios_remote_provider/repos/vidapp`  
  * Configure your Developer Team on the project  
  * Build and install it to your iOS device
* Run `main -register` ; hit enter to use the default alpha registration password  
* Start ios_video_app "Recording" on your iOS device(s)
  * Add Screen Recording to Control Center on your iOS device if you never did so before  
  * Open Control Center on your iOS device  
  * Select the red circle icon that is Screen Recording  
  * Select `vidtest2`  
  * Click `Start Recording`

Diagram of architecture of Control Floor attached.
![ControlFloor](https://user-images.githubusercontent.com/905365/106125382-f30cb780-6110-11eb-9db1-d74b289205fd.png)
