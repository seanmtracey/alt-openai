# alt-llava
A super-simple utility for generating alt-text for images using the LVM "llava" via Ollama.

## Building + Running

First, clone this repo to your local system:

`git clone https://github.com/seanmtracey/alt-llava`

### Building

To build, you will need Go >= v1.22.2. `cd` into your cloned repo, and then run the following:

```bash
go mod tidy # Install dependencies
go build . 
```

This will create a binary `alt-llava`.

### Running

To run with minimal configuration, you will need to set at least 1 environment variable: `IMAGE_URL` with a URL that links to an image. This image will be downloaded and stored at `./images` with a UUID and a file extension derived from the image's MIME type.

`IMAGE_URL=<URL_TO_IMAGE> ./alt-llava`

`alt-llava` will then generate the alt-text for the image and log it to the CLI.

## Flags

If you wish, you can write the output of `alt-llava` to a text file by passing the `--out` flag with the filepath you wish to write your results to.

```bash
./alt-llava --out="./output/results.txt"
```

## Environment Variables.

1. `OLLAMA_HOST` - The hostname where the Ollama server is running.
2. `OLLAMA_PORT` - The port the Ollama server is listening on.
3. `SILENT_OUTPUT` - Only output the alt-text to `stdout`. Default: `false`
4. `OLLAMA_MODEL` - The model that Ollama should run to generate the alt-text. Default: `llava`
5. `OLLAMA_PROTOCOL` - `http` or `https`. Default: `http`
6. `IMAGE_URL` - The image that should be downloaded, and have alt-text generated for.


