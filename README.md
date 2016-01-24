![logo](https://raw.githubusercontent.com/gophergala2016/togepi/master/logo.png)

**Togepi** is a user-friendly tool for sharing files over the network. Sender's/receiver's location doesn't matter for Togepi, neither it cares about your firewall settings, it just works and makes file sharing as easy as never before.

## Why Togepi?

- **Forget about file location and host address.** For each shared file the application will generate a unique hash string which will be used to pull the file from the remote machine.
- **Firewall is not a problem.** Do you have a firewall blocking all the incoming connections? Togepi will deal with it. It automatically determines whether it's possible of not for a client to pull a file directly from your computer. And in case if it's not - the connection will go through the Togepi server.
- **Deploy your own server.** The same way most of the companies host theit own Git repository hosting service or a Docker registry, the Togepi server can be deployed to accept file transations somewhere in a local network.
