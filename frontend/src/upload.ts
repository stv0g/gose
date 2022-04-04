import { ProgressHandler } from "./progress-handler";
import { buf2hex, hex2buf, arraybufferEqual } from "./utils";
import { apiRequest} from "./api";
import * as md5 from "js-md5";

export class UploadParams {
    server: string
    expiration: string
    notify_mail: string
    notify_browser: boolean
    short_url: boolean
}

class Callbacks {
    [key: string]: any
}

class Part {
    number: number;
    offset: number;
    length: number;
    etag: ArrayBuffer;

    constructor(num: number, etag: ArrayBuffer, len?: number, off?: number) {
        this.number = num;
        this.offset = off;
        this.length = len;
        this.etag = etag;
    }

    toJSON(): any {
        return {
            ...this,
            etag: buf2hex(this.etag)
        };
    }

    static fromJSON(json: any): Part {
        return new Part(json.number, hex2buf(json.etag), json.length);
    }
}

export class Upload {
    file: File | null = null;
    url: string;
    progress: ProgressHandler | null = null;
    etag: string;
    stage: string;
    inProgress: boolean = false;

    protected parts: Part[] = [];
    protected callbacks: Callbacks;
    protected params: UploadParams;
    protected xhr: XMLHttpRequest;

    // TODO: take this from the configuration
    readonly partSize = 6e6;

    constructor(file: File, cbs: Callbacks, params: UploadParams) {
        this.file = file;
        this.callbacks = cbs;
        this.params = params;

        const partsCount = Math.ceil(this.file.size / this.partSize);

        this.progress = new ProgressHandler({
            start: () => this.callbacks.start(this),
            end: () => this.callbacks.end(this),
            progress: () => this.callbacks.progress(this),
        }, this.file.size, partsCount);
    }

    async start() {
        try {
            this.inProgress = true;

            if (this.file.size == 0) {
                throw "Cannot upload empty file";
            }
    
            [this.parts, this.etag] = await this.hash();
    
            return await this.upload();
        } catch(e) {
            throw e;
        } finally {
            this.inProgress = false;
        }
    }

    async hash(): Promise<[ Part[], string ]> {
        this.stage = "hashing";

        this.progress.start();

        let parts: Part[] = [];
        let partNumber = 1;
        for (let offset = 0; offset < this.file.size; offset += this.partSize) {
            let length = this.partSize;
            if (offset + length > this.file.size) {
                length = this.file.size - offset; // handle last part
            }

            let part = this.file.slice(offset, offset + length);
            let partBuffer = await part.arrayBuffer();
            if (!this.file) {
                throw "Aborted";
            }

            this.progress.partStart(new ProgressEvent("", {
                loaded: 0,
                total: length
            }));

            let md = md5.create();
            // let chunkSize = 1<<20;
            let chunkSize = this.partSize;
            for (let chunkOffset = 0; chunkOffset < partBuffer.byteLength; chunkOffset += chunkSize) {
                let chunkLength = chunkSize;
                if (chunkOffset + chunkSize > partBuffer.byteLength) {
                    chunkLength = partBuffer.byteLength - chunkOffset;
                }

                let chunkBuffer = partBuffer.slice(chunkOffset, chunkOffset+chunkLength);

                md.update(chunkBuffer);

                this.progress.partProgress(new ProgressEvent("", {
                    loaded: chunkOffset+chunkLength,
                    total: length
                }));
            }

            this.progress.partEnd(new ProgressEvent("", {
                loaded: length,
                total: length
            }));

            let etag = md.arrayBuffer();

            parts.push(new Part(partNumber++, etag, length, offset));
        }

        this.progress.end();

        let etagBlob = new Blob(parts.map(p => p.etag));
        let etagBuf = await etagBlob.arrayBuffer();
        let etag = md5.arrayBuffer(etagBuf);
        let etagStr = `${buf2hex(etag)}-${parts.length}`;

        return [parts, etagStr];
    }

    async upload() {
        this.stage = "uploading";

        let respInitiate = await apiRequest("initiate", {
            server: this.params.server,
            filename: this.file.name,
            etag: this.etag,
            short_url: this.params.short_url,
            type: this.file.type
        });

        if (respInitiate.upload_id === undefined) {
            return respInitiate.url;
        }

        this.url = respInitiate.url;

        let existingParts: {[x: number]: Part} = {}
        for (let part of respInitiate.parts) {
            existingParts[part.number] = Part.fromJSON(part);
        }

        this.progress.start();

        for (let i = 0; i < this.parts.length; i++) {
            let part = this.parts[i];
            if (part.number in existingParts) {
                let existingPart = existingParts[part.number];

                if (arraybufferEqual(existingPart.etag, part.etag)) {
                    this.progress.partSkip(existingPart.length);
                    continue;
                }
            }

            let chunk = this.file.slice(part.offset, part.offset + part.length);

            let partResp = await apiRequest("part", {
                server: this.params.server,
                etag: respInitiate.etag,
                upload_id: respInitiate.upload_id,
                checksum: buf2hex(part.etag),
                length: part.length,
                number: part.number
            })

            let etag = await this.uploadPart(partResp.url, chunk);
            if (!this.file) {
                throw "Aborted";
            }

            if (etag !== "\"" + buf2hex(part.etag) + "\"") {
                throw "Checksum mismatch";
            }
        }

        this.progress.end();

        let respComplete = await apiRequest("complete", {
            server: this.params.server,
            etag: respInitiate.etag,
            upload_id: respInitiate.upload_id,
            parts: this.parts.map(p => p.toJSON()),
            expiration: this.params.expiration,
            notify_mail: this.params.notify_mail,
        });

        if (respComplete.etag !== this.etag) {
            throw "Final checksum mismatch";
        }

        return respInitiate.url || respComplete.url;
    }

    async uploadPart(url: string, part: Blob) {
        let prom = new Promise<XMLHttpRequest>((resolve, reject) => {
            this.xhr = new XMLHttpRequest();
            this.xhr.open("PUT", url);
            this.xhr.onload = () => {
                if (this.xhr.status >= 200 && this.xhr.status < 300) {
                    resolve(this.xhr);
                } else {
                    reject({
                        status: this.xhr.status,
                        statusText: this.xhr.statusText
                    });
                }
            };

            this.xhr.onerror = () => {
                reject({
                    status: this.xhr.status,
                    statusText: this.xhr.statusText
                });
            };

            this.xhr.onabort = () => {
                reject("Aborted");
            }

            this.xhr.upload.onprogress = (ev) => this.progress.partProgress(ev);
            this.xhr.upload.onloadstart = (ev) => this.progress.partStart(ev);
            this.xhr.upload.onloadend = (ev) => this.progress.partEnd(ev);

            this.xhr.send(part);
        });

        let resp = await prom;
        if (resp.status !== 200) {
            // TODO: Decode AWS S3 error here.
            throw "Failed to upload part";
        }

        return resp.getResponseHeader("etag");
    }

    abort() {
        if (this.xhr) {
            this.xhr.abort();
        }

        // Used to signal hash()
        this.file = null;
    }
}
