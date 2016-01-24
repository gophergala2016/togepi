![logo](https://raw.githubusercontent.com/gophergala2016/togepi/master/logo.png)

**Togepi** is a user-friendly tool for sharing files over the network. Sender's/receiver's location doesn't matter for Togepi, neither it cares about your firewall settings, it just works and makes file sharing as easy as never before.

## Why Togepi?

- **Forget about file location and host address.** For each shared file the application generates a unique hash string which is used to pull the file from the remote machine.
- **Firewall is not a problem.** Do you have a firewall blocking the incoming connections? Togepi will deal with it. It automatically determines whether it's possible of not for a client to pull a file directly from your computer. And in case if it's not - the connection will go through the Togepi server.
- **Deploy your own server.** The same way most of the companies host theit own Git repository hosting service or a Docker registry, the Togepi server can be deployed to accept file transations somewhere in a network or between a limited group of users.
- **The original file is located on your machine.** You don't upload files when sharing, instead, you tell the world that you'have shared them, so all the others can pull the files from your computer. None of your data is being hosted on the server but hash IDs.

## A little demo

Take a look at how the application shares files between 3 machines.

![demo](https://raw.githubusercontent.com/gophergala2016/togepi/master/demo.gif)

![diagram](https://raw.githubusercontent.com/gophergala2016/togepi/master/diagram.gif)

## Usage

In order to share files, the daemon must be started first:
```bash
$ togepi -start &
```
By default it will connect to the server running in my cloud, so no need to set up anything.

#### Sharing files

To share a file simply provide it's path (can be relative or full) as a single argument
```bash
$ togepi path/to/file
e9ad9cf77403719f4e06351355c1781a1ebe57
```

#### Pulling files

It can't be easier than to provide the "Share Hash"
```bash
$ togepi e9ad9cf77403719f4e06351355c1781a1ebe57
file file saved
```

#### List shared files

Executing Togepi with the -a flag will output a list of shared hashes along with the corresponding file paths.
```bash
$ togepi -a
2b892dfea9819fb5a1a937bc34d2934189cf4355249d0186 /run/media/alex.ant/HDD/Music/01-chickenfoot-avenida_revolution.mp3
7ee33d7d62f1e564b080366aa1bc1e4c6ca5b01c838adcf3 /home/alex.ant/demo/02.Its Electric.mp3
893eeed69da96fdbe9dc28261c09116a19b1c4868a9aa24c /home/alex.ant/LICENSE
```

#### Remove shared files

In case you no longer want a file to be shared with the world, you can remove it from the shared list with the -rm command.
```bash
$ togepi -rm e9ad9cf77403719f4e06351355c1781a1ebe57
```

#### Start your own server

If you want to run your own server, kick it off the following way:
```bash
$ togepi -server
```
And then connect to it the daemon
```bash
$ togepi -start -http-host 127.0.0.1:8011 -tcp-host 127.0.0.1:8012 -redis-host 127.0.0.1:6379
```
The only required service for the server is Redis DB.

#### And BTW!
5c4ab095a32ff352d301e08ae966b287fcbfe8cf371d998411e2b3a29e2c7ada4366367a5bc7fd1b3ea856c7af14a838

If you wanna know what's in there, don't be shy and try Togepi right now! :)
