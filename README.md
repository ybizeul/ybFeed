### Introduction

In a nutshell, ybFeed is a personnal micro feed where you can post snippets of 
text or images.

The primary use case is to share information between computers where you don't
have the possibility to copy/paste, like a restricted VDI environment.

Open the feed on your local computer and the remote one, then everything you add
will be displayed on the other browser as well.

### Concepts

When going to the home page, you are invited to create a feed with a unique
name.

Once on a feed, you can paste data on it, text or images, they will be added
to the feed by reverse order.

You can then decide to share the feed two different ways :

- Copy a secret link to the feed, than you can paste on a different computer,
you will be automatically authenticated
- If copy/paste is not an option, you can set a temporary 4 digit PIN. You then
go to the other computer and it will ask for the PIN when you open the page.

That's pretty much all there is to it, you can paste and delete items from any
browser.

### Screenshot

![Screenshot](screenshot.png)

### Caveats

This is just a side project I'm working on, so there is probably lots of issues

Here are some I already identified :

- ybFeed relies on a cookie to authenticate the session, if the cookie is lost
there is no easy way to retrieve the feed
- Most modern browser won't honor loog cookie lifetime, so a JWT based session
would be much better, stay tuned
- Mobile usage is pretty limited as you cannot paste data currently
- Security could probably be improved, tokens and PINs are stored in clear on
the filesystem
- Logging is pretty much inexistant
- You cannot currently customize the default port 8080
- No rate control or capacity limits

### Building

#### Using Makefile

```
make
```

#### Manually

Once you cloned the repository, issue the following commands :
```
# Install node dependencies
cd ui
npm install

# Build UI
npm run build

# Build Go binary
cd ../
go build -o ybFeed *.go

# Run ybFeed
./ybFeed

# Point your browser to port 8080
```

### Environment variables
`YBF_DATADIR` points to an alternative direcotry to store data, default is
`./data/` in current directory.

### Docker

#### Building container

```
docker build . -t ybfeed
```

#### Running with Compose

```
docker compose up -d
```
