<!--
SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
SPDX-License-Identifier: Apache-2.0
-->

<p align="center">
    <img style="width: 30%; margin: 4em 0" src="frontend/img/gose-logo.svg" alt="GoSƐ logo" />
    <h1 align="center">GoSƐ - A terascale file-uploader</h1>
</p>

<!-- [![Codacy coverage](https://img.shields.io/codacy/coverage/27eec133fcfd4459885d78f52d03daa8?style=flat-square)](https://app.codacy.com/gh/stv0g/gose/) -->
<!-- [![GitHub Workflow Status](https://img.shields.io/github/workflow/status/stv0g/wice/build?style=flat-square)](https://github.com/stv0g/wice/actions) -->
[![goreportcard](https://goreportcard.com/badge/github.com/stv0g/gose?style=flat-square)](https://goreportcard.com/report/github.com/stv0g/gose/)
[![Codacy grade](https://img.shields.io/codacy/grade/27eec133fcfd4459885d78f52d03daa8?style=flat-square)](https://app.codacy.com/gh/stv0g/gose/)
[![License](https://img.shields.io/github/license/stv0g/gose?style=flat-square)](https://github.com/stv0g/gose/blob/master/LICENSE)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/stv0g/gose?style=flat-square)
[![Go Reference](https://pkg.go.dev/badge/github.com/stv0g/gose.svg)](https://pkg.go.dev/github.com/stv0g/gose)

GoSƐ is a modern and scalable file-uploader focusing on scalability and simplicity. It is a little hobby project I’ve been working on over the last weekends.

The only requirement for GoSƐ is a S3 storage backend which allows to it to scale horizontally without the need for additional databases or caches. Uploaded files a divided into equally sized chunks which are hashed with a MD5 digest in the browser for upload. This allows GoSƐ to skip chunks which already exist. Seamless resumption of interrupted uploads and storage savings are the consequence.

And either way both upload and downloads are always directed directly at the S3 server so GoSƐ only sees a few small HTTP requests instead of the bulk of the data. Behind the scenes, GoSƐ uses many of the more advanced S3 features like [Multi-part Uploads](https://docs.aws.amazon.com/AmazonS3/latest/userguide/mpuoverview.html) and [Pre-signed Requests](https://docs.aws.amazon.com/AmazonS3/latest/userguide/using-presigned-url.html) to make this happen.

Users have a few options to select between multiple pre-configured S3 buckets/servers or enable browser & mail notifications about completed uploads. A customisable retention / expiration time for each upload is also selectable by the user and implemented by [S3 life-cycle policies](https://docs.aws.amazon.com/AmazonS3/latest/userguide/object-lifecycle-mgmt.html). Optionally, users can also opt-in to use an external service to shorten the URL of the uploaded file.

Currently a single concurrent upload of a single file is supported. Users can observe the progress via a table of details statistics, a progress-bar and a chart showing the current transfer speed.

GoSƐ aims at keeping its deployment simple and by bundling both front- & backend components in a single binary or Docker image. GoSƐ has been tested with AWS S3, Ceph’s RadosGW and Minio. Pre-built binaries and Docker images of GoSƐ are available for all major operating systems and architectures at the [release page](https://github.com/stv0g/gose/releases).

GoSƐ is open-source software licensed under the Apache 2.0 license.

Check our my [blog article](https://noteblok.net/2022/04/03/gos%c9%9b-a-terascale-file-uploader/) for more background info.

## Features

-   De-duplication of uploaded files based on their content-hash
    -   Uploads of existing files will complete in no-time without re-upload
-   S3 Multi-part uploads
    -   Resumption of interrupted uploads
-   Drag & Drop of files
-   Browser notifications about failed & completed uploads
-   User-provided object expiration/retention time
-   Copy URL of uploaded file to clip-board
-   Detailed transfer statistics and progress-bar / chart
-   Installation via single binary or container
    -   JS/HTML/CSS Frontend is bundled into binary
-   Scalable to multiple replicas
    -   All state is kept in the S3 storage backend
    -   No other database or cache is required
-   Direct up & download to Amazon S3 via presigned URLs
    -   Gose deployment does not see an significant traffic
-   UTF-8 filenames
-   Multiple user-selectable buckets / servers
-   Optional link shortening via an external service
-   Optional notification about new uploads via [shoutrrr](https://containrrr.dev/shoutrrr/v0.5/)
    -   Mail notifications to user-provided recipient
-   Cross-platform support:
    -   Operating systems: Windows, macOS, Linux, BSD
    -   Architectures: arm64, amd64, armv7, i386

## Roadmap

Checkout the [Github issue tracker](https://github.com/stv0g/gose/issues?q=is%3Aissue+is%3Aopen+label%3Aenhancement).

## [Demo](https://gose.0l.de) (click for Live-Demo)

<p align="center">
    <a href="https://gose.0l.de">
        <img style="max-width: 400px" src="https://user-images.githubusercontent.com/285829/161408820-f0f304ff-2066-462a-98d6-d59a79889d83.gif" alt="Gose demo screencast" />
    </a>
</p>

## Installation

### Pre-compiled binaries from GitHub releases

Take the download link for your OS/Arch from the [Releases Page](https://github.com/stv0g/gose/releases/) and run:

```bash
sudo wget https://github.com/stv0g/gose/releases/download/v0.0.2/gose_0.0.2_linux_amd64 -O /usr/local/bin/gose
chmod +x /usr/local/bin/gose
```

### Kubernetes / Kustomize

1.  Copy default configuration file: `cp config.yaml kustomize/config.yaml`
2.  Adjust config: `nano kustomize/config.yaml`
3.  Apply configuration: `kubectl apply -k kustomize`

### Docker

Via environment variables in `.env` file:

```bash
docker run --env-file=.env --publish=8080:8080 ghcr.io/stv0g/gose
```

or via a configuration file:

```bash
docker run -v$(pwd)/config.yaml:/config.yaml --publish=8080:8080 ghcr.io/stv0g/gose -config /config.yaml
```

#### Docker Compose

We ship a `docker-compose.yml` file to get you started.
Please adjust the environment variables in it and then run:

```bash
docker-compose up -d
```

## Configuration

Gose can be configured via a configuration file and/or environment variables

### File

For reference have a look at the [example configuration file](config.yaml).

### Environment variables

All settings from the configuration file can also be set via environment variables:

| Variable                               | Default Value                                                             | Description                           |
| :--                                    | :--                                                                       | :--                                   |
| `GOSE_LISTEN`                          | `":8080"`                                                                 | Listen address and port of Gose       |
| `GOSE_BASE_URL`                        | `"http://localhost:8080"`                                                 | Base URL at which Gose is accessible  |
| `GOSE_STATIC`                          | `"./dist"`                                                                | Directory of frontend assets (pre-compiled binaries of GoSƐ come with assets embedded into binary.) |
| `GOSE_BUCKET`                          | `gose-uploads`                                                            | Name of S3 bucket                     |
| `GOSE_ENDPOINT`                        | (without `http(s)://` prefix, but with port number)                       | Hostname:Port of S3 server            |
| `GOSE_REGION`                          | `us-east-1`                                                               | Region of S3 server                   |
| `GOSE_PATH_STYLE`                      | `false`                                                                   | Prepend bucket name to path           |
| `GOSE_NO_SSL`                          | `false`                                                                   | Disable SSL encryption for S3         |
| `GOSE_ACCESS_KEY`                      |                                                                           | S3 Access Key                         |
| `GOSE_SECRET_KEY`                      |                                                                           | S3 Secret Key                         |
| `GOSE_SETUP_BUCKET`                    | `true`                                                                    | Create S3 bucket if do not exists     |
| `GOSE_SETUP_CORS`                      | `true` (if supported by S3 implementation)                                | Setup S3 bucket CORS rules            |
| `GOSE_SETUP_LIFECYCLE`                 | `true` (if supported by S3 implementation)                                | Setup S3 bucket lifecycle rules       |
| `GOSE_SETUP_ABORT_INCOMPLETE_UPLOADS`  | `31`                                                                      | Number of days after which incomplete uploads are cleaned-up (set to 0 to disable) |
| `GOSE_MAX_UPLOAD_SIZE`                 | `1TB`                                                                     | Maximum upload size                   |
| `GOSE_PART_SIZE`                       | `16MB`                                                                    | Part-size for multi-part uploads      |
| `AWS_ACCESS_KEY_ID`                    |                                                                           | alias for `GOSE_ACCESS_KEY`           |
| `AWS_SECRET_ACCESS_KEY`                |                                                                           | alias for `GOSE_SECRET_KEY`           |

Configuration of link shortener and notifiers must be done via a [configuration file](#file).

## Author

GoSƐ has been written by [Steffen Vogel](mailto:post@steffenvogel.de).

## License

GoSƐ is licensed under the [Apache 2.0 license](./LICENSE).
