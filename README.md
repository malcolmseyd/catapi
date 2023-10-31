# Cat API

This is a web API for serving cat images! More generally, it serves a random image from a folder, optionally drawing meme text on it. It supports JPEG, PNG, and GIF images.

As of writing, I have an instance hosted at https://catapi.seyd.ca, go check it out!

This project is shamelessly inspired by https://cataas.com/

## API

Get a random cat
```
/cat
```

Get a random cat with a message overlayed
```
/cat?msg=[MESSAGE]
```

Get a specific cat (mostly for debugging)
```
/cat?id=[IMAGE_ID]
```

## Running

You will need two resources to run this program:

- `impact.ttf`: The famous meme font, in TrueType format
- `img/`: A folder full of cat pictures! Run `cmd/scraper` to download a collection from cataas. Set --outdir to `img/` and, if you get rate limited, reduce the number of parallel workers using the --maxworkers flag.

Additionally, the following environment variables are respected:

- `IMPACT_FILENAME`: Location of the impact font (or any other font if you'd like). Defaults to "impact.ttf"
- `LISTEN_HOST`: Hostname for the HTTP server to listen on. Defaults to ""
- `LISTEN_PORT`: Port for the HTTP server to listen on. Defaults to 8080
- `GIN_MODE`: Mode for the Gin web server. Options are "debug", "release", or "test". Defaults to "debug".

I have also included my deployment files which include a Dockerfile and a Fly.io config. Feel free to use them if you wish.