# Hadolint-Api
Dockerfile linter api server in kernel of [hadolint](https://github.com/lukasmartinelli/hadolint).

## API

### Request

METHOD: POST

URL: host/api/dockerfile

BODY:
```json
{
  "dockerfile":""
}
```
### Response
```json
{
  "line_number":["linter"]
}
```

## Example
```shell
curl -X POST -H "Content-Type: application/json"  \
  -d '{"dockerfile": "FROM ubuntu\nEXPOSE 808000\nRUN cd /usr/app"}' \
  "http://hadolint.daoapp.io/api/dockerfile"

{
  "0": [
    "DL4000 Specify a maintainer of the Dockerfile"
  ],
  "1": [
    "DL3006 Always tag the version of an image explicitly."
  ],
  "2": [
    "DL3011 Valid UNIX ports range from 0 to 65535"
  ],
  "3": [
    "SC2164 Use cd ... || exit in case cd fails.",
    "DL3003 Use WORKDIR to switch to a directory"
  ]
}
```

## LICENSE

In GPL V3
