import { md5sum } from "./checksum";
import { ProgressHandler } from "./progress-handler";
import { buf2hex } from "./utils";
import { apiRequest} from "./api";
import { ChecksummedFile } from "./file";

export class UploadParams {
    server: string
    expiration: string
    notify_mail: string
    notify_browser: boolean
    shorten_link: boolean
}

async function md5sumHex(blob: Blob) {
    let ab = await blob.arrayBuffer();

    let hash = await md5sum(ab);

    return buf2hex(hash);
}

class Callbacks {
    [key: string]: any
}

export class Upload {
    url: string;
    uploadID: string | null;
    callbacks: Callbacks;
    params: UploadParams;
    progress: ProgressHandler | null = null;
    file: ChecksummedFile | null = null;

    constructor(cbs: Callbacks, params: UploadParams) {
        this.callbacks = cbs;
        this.params = params;
    }

    async upload(file: ChecksummedFile) {
        this.file = file;

        let respInitiate = await apiRequest("initiate", {
            server: this.params.server,
            filename: file.name,
            shorten_link: this.params.shorten_link,
            expiration: this.params.expiration,
            content_length: file.size,
            content_type: file.type,
            checksum: buf2hex(file.checksum)
        });

        this.url = respInitiate.url;
        this.uploadID = respInitiate.upload_id;

        this.progress = new ProgressHandler({
            start: () => this.callbacks.start(this),
            end: () => this.callbacks.end(this),
            progress: () => this.callbacks.progress(this),
        }, file.size, respInitiate.parts.length);

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
            server: this.params.server,
            key: respInitiate.key,
            upload_id: respInitiate.upload_id,
            parts: parts,
            notify_mail: this.params.notify_mail
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

    async uploadPart(url: string, part: Blob) {
        let prom = new Promise<XMLHttpRequest>((resolve, reject) => {
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
