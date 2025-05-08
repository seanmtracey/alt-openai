# alt-openai
A super-simple utility for generating alt-text for images using OpenAI

## Building + Running

First, clone this repo to your local system:

`git clone https://github.com/seanmtracey/alt-openai`

### Building

To build, you will need Go >= v1.22.2. `cd` into your cloned repo, and then run the following:

```bash
go mod tidy # Install dependencies
go build . 
```

This will create a binary `alt-openai`.

### Running

To run with minimal configuration, you will need to set at least 3 environment variables: 

1. `IMAGE_URL` - a URL that links to an image. This image will be downloaded and stored at `./images` with a UUID and a file extension derived from the image's MIME type.

2. `OPEN_AI_KEY` - an OpenAI API Key

3. `OPEN_AI_ORIGIN` - the domain that the OpenAI API is accessed at. By default, this should be `api.openai.com`.

`IMAGE_URL=<URL_TO_IMAGE> OPEN_AI_KEY=<API_KEY> OPEN_AI_ORIGIN="api.openai.com" ./alt-openai`

`alt-openai` will then generate the alt-text for the image and log it to the CLI.

## Flags

If you wish, you can write the output of `alt-openai` to a text file by passing the `--out` flag with the filepath you wish to write your results to.

```bash
./alt-openai --out="./output/results.txt"
```

## Environment Variables.

1. `OPEN_AI_ORIGIN` - the domain that the the OpenAI API is accessed at,
2. `OPEN_AI_KEY` - Your OpenAI API Key
3. `OPEN_AI_MODEL` - The model you want to use to generate the alt-text. `gpt-4-turbo` works fairly well.
4. `SILENT_OUTPUT` - Only output the alt-text to `stdout`. Default: `false`
5. `IMAGE_URL` - The image that should be downloaded, and have alt-text generated for.
6. `PROMPT_TEXT` - The prompt for the LVM that you'd like to use to describe the intended output for the model.


