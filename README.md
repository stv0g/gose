<p align="center" >
    <img style="width: 30%; margin: 4em 0" src="frontend/img/gose-logo.svg" alt="GoSƐ logo" />
    <h1 align="center">GoSƐ - A terascale file-uploader</h1>
</p>

<!-- [![Codacy coverage](https://img.shields.io/codacy/coverage/27eec133fcfd4459885d78f52d03daa8?style=flat-square)](https://app.codacy.com/gh/stv0g/gose/) -->
<!-- [![GitHub Workflow Status](https://img.shields.io/github/workflow/status/stv0g/wice/build?style=flat-square)](https://github.com/stv0g/wice/actions) -->
[![goreportcard](https://goreportcard.com/badge/github.com/stv0g/gose?style=flat-square)](https://goreportcard.com/report/github.com/stv0g/gose/)
[![Codacy grade](https://img.shields.io/codacy/grade/27eec133fcfd4459885d78f52d03daa8?style=flat-square)](https://app.codacy.com/gh/stv0g/gose/)
[![libraries.io](https://img.shields.io/librariesio/github/stv0g/gose?style=flat-square)](https://libraries.io/github/stv0g/gose)
[![License](https://img.shields.io/github/license/stv0g/gose?style=flat-square)](https://github.com/stv0g/gose/blob/master/LICENSE)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/stv0g/gose?style=flat-square)
[![Go Reference](https://pkg.go.dev/badge/github.com/stv0g/gose.svg)](https://pkg.go.dev/github.com/stv0g/gose)

## [Demo](https://gose.0l.de)

## Features

-   Upload progress-bar and statistics
-   Direct upload to Amazon S3 via presigned-URLs
-   Direct download from Amazon S3
-   Link shortening
-   Drag & Drop
-   Multi-part / chunked upload
-   File integrity checks via using MD5 checksum & ETags

## Roadmap

-   Resumable uploads

-   Configurable retention time

-   Server-side encryption

-   End-to-end encryption
    -   Using [streaming requests](https://web.dev/fetch-upload-streaming/)

-   Support for multiple buckets

-   Torrent web-seeding
    -   [BEP-17](http://bittorrent.org/beps/bep_0017.html) and/or [BEP-19](http://bittorrent.org/beps/bep_0019.html)

## Author

GoSƐ has been written by [Steffen Vogel](mailto:post@steffenvogel.de).

## License

GoSƐ is licensed under the [Apache 2.0 license](./LICENSE).
