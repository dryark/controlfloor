# controlfloor
System for controlling devices remotely

# Basic Install Instructions
* First download Control Floor
* Alter main.go to place the hostname into the listen command of your host
* Build Control Floor itself. Then run it to get an interface. The current alpha version just uses a hardcoded password of 'ok' and 'ok' to login.
* Download and build ios_remote_provider.
* Download and build iosif. Place the iosif binary in ios_remote_provider/bin folder
* Build WebDriverAgent through stf_ios_support, then copy the stf_ios_folder/bin/wda folder to ios_remote_provider/bin/wda .
* Run `main -register` ; hit enter to use the default alpha registration password
* Run main to start the provider

Example config.json file for ios_remote_provider:
```
{
    bin_paths: {
        ios-deploy: "bin/ios-deploy",
        mobiledevice: "bin/mobiledevice"
    },
    controlfloor: {
        host: "[hostname of your server]:8080",
        username: "[some name for your provider node]",
    }
    port: 8027
}
```

Diagram of architecture of Control Floor attached.
![ControlFloor](https://user-images.githubusercontent.com/905365/106125382-f30cb780-6110-11eb-9db1-d74b289205fd.png)
