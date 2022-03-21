import { md5sum } from "./checksum.js";
import { ProgressHandler } from "./progress-handler.js";
import { buf2hex } from "./utils.js";

const apiBase = "/api/v1";

async function md5sumHex(blob) {
    let ab = await blob.arrayBuffer();

    let hash = await md5sum(ab);

    return buf2hex(hash);
}

async function apiRequest(req, body) {
    let resp = await fetch(`${apiBase}/${req}`, {
        method: "POST",
        mode: "cors",
        headers: {
            "Content-Type": "application/json"
        },
        body: JSON.stringify(body)
    });

    if (resp.status !== 200) {
        throw resp;
    }

    return resp.json();
}

export class Upload {

    constructor(cb) {
        this.callbacks = cb;
    }

    async upload(file) {
        let respInitiate = await apiRequest("initiate", {
            filename: file.name,
            content_length: file.size,
            content_type: file.type,
            checksum: buf2hex(file.checksum)
        });

        this.progress = new ProgressHandler({
            start: (progress) => this.callbacks.start(this),
            end: (progress) => this.callbacks.end(this),
            progress: (progress) => this.callbacks.progress(this),
        }, file.size, respInitiate.parts.length);

        this.url = respInitiate.url;

        this.progress.loadStart();

        let parts = [];
        let etags = [];
        for (let i = 0; i < respInitiate.parts.length; i++) {
            let start = i * respInitiate.part_size;
            let end = (i + 1) * respInitiate.part_size;
            if (i === respInitiate.parts.length - 1) {
                end = file.size;
            }

            let chunk = file.slice(start, end);
            let url = respInitiate.parts[i];

            let chunkBuffer = await chunk.arrayBuffer();
            let etag = await this.uploadPart(url, chunk);
            let etagExpected = await md5sum(chunkBuffer);
            if (etag !== "\"" + buf2hex(etagExpected) + "\"") {
                throw {
                    status: 400,
                    statusText: "Checksum mismatch"
                };
            }

            etags.push(etagExpected);
            parts.push({
                etag: etag,
                part_number: i + 1
            });
        }

        this.progress.loadEnd();

        let respComplete = await apiRequest("complete", {
            key: respInitiate.key,
            upload_id: respInitiate.upload_id,
            parts: parts
        });

        let etagBlob = new Blob(etags);
        let objEtag = await md5sumHex(etagBlob) + `-${parts.length}`;

        if (respComplete.etag !== objEtag) {
            throw {
                status: 400,
                statusText: "Final checksum mismatch"
            };
        }

        return respInitiate.url;
    }

    async uploadPart(url, part) {
        let prom = new Promise((resolve, reject) => {
            let xhr = new XMLHttpRequest();

            xhr.open("PUT", url);

            xhr.onload = function() {
                if (this.status >= 200 && this.status < 300) {
                    resolve(this);
                } else {
                    reject({
                        status: this.status,
                        statusText: xhr.statusText
                    });
                }
            };

            xhr.onerror = function() {
                reject({
                    status: this.status,
                    statusText: xhr.statusText
                });
            };

            xhr.upload.onprogress = (ev) => this.progress.partProgress(ev);
            xhr.upload.onloadstart = (ev) => this.progress.partLoadStart(ev);
            xhr.upload.onloadend = (ev) => this.progress.partLoadEnd(ev);

            xhr.send(part);
        });

        let resp = await prom;
        if (resp.status !== 200) {
            throw resp;
        }

        let etag = resp.getResponseHeader("etag");

        return etag;
    }
}
