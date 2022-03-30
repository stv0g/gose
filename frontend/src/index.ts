import "bootstrap";

import "../css/index.scss";

import prettyBytes  from "pretty-bytes";
import * as prettyMilliseconds from "pretty-ms";

import { ProgressBar } from "./progress-bar";
import { Upload, UploadParams } from "./upload";
import { apiRequest } from "./api";
import { sha256sum } from "./checksum";
import { Config, Server } from "./config";
import { ChecksummedFile } from "./file";

var statsTransferred: HTMLElement;
var statsElapsed: HTMLElement;
var statsEta: HTMLElement;
var statsSpeed: HTMLElement;
var statsParts: HTMLElement;
var progressBar: ProgressBar;
var dropZone: HTMLElement;
var uploadInProgress: boolean = false;
var config: object;

function uploadStarted(upload: Upload) {
    let p = upload.progress;

    let resultElm = document.getElementById("result");

    resultElm.classList.remove("alert-danger", "alert-success", "d-none");
    resultElm.classList.add("alert-warning");
    resultElm.innerHTML = `Upload in progress: <a href="${upload.url}">${upload.url}</a>`;

    progressBar.setMinMax(0, p.totalSize);
    progressBar.set(0);
}

function uploadEnded(upload: Upload) {
    let p = upload.progress;

    statsTransferred.textContent = prettyBytes(p.totalTransferred);

    progressBar.set(p.totalSize);
}

function uploadProgressed(upload: Upload) {
    let p = upload.progress;

    progressBar.set(p.transferred + upload.progress.totalTransferred);

    statsTransferred.textContent = prettyBytes(p.transferred + p.totalTransferred) + " / " + prettyBytes(p.totalSize);
    statsElapsed.textContent = prettyMilliseconds(p.elapsed + p.totalElapsed, { compact: true });
    statsEta.textContent = prettyMilliseconds(p.eta, { compact: true });
    statsSpeed.textContent = prettyBytes(p.speed, { bits: true }) + "/s";
    statsParts.textContent = `${p.part} / ${p.totalParts}`;
}

async function startUpload(files: FileList) {
    let resultElm = document.getElementById("result");
    let params = getUploadParams();

    try {
        uploadInProgress = true;

        if (files.length === 0) {
                return;
        }
        else if (files.length > 1) {
            throw {
                status: 400,
                statusText: "Can only upload a single file"
            };
        }

        let file = files[0] as ChecksummedFile;
        let ab = await file.arrayBuffer();

        file.checksum = await sha256sum(new Uint8Array(ab));

        let upload = new Upload({
            start: uploadStarted,
            end: uploadEnded,
            progress: uploadProgressed,
        }, params);

        let url = await upload.upload(file);

        console.log("Upload succeeded", url);

        resultElm.classList.remove("alert-danger", "alert-warning");
        resultElm.classList.add("alert-success");
        resultElm.innerHTML = `Upload complete: <a href="${url}">${url}</a>`;

        if (params.notify_browser) {
            let dur = prettyMilliseconds(upload.progress.totalElapsed, { compact: true });
            let size = prettyBytes(upload.progress.totalSize);
    
            new Notification("Upload completed", {
                body: `Upload of ${size} for ${upload.file.name} has been completed in ${dur}.`,
                icon: 'gose-logo.svg',
                renotify: true,
                tag: upload.uploadID
            });
        }
    } catch (e) {
        console.log("Upload failed", e);

        resultElm.classList.remove("alert-success", "alert-warning", "d-none");
        resultElm.classList.add("alert-danger");
        resultElm.textContent = `${e.status} - ${e.statusText}`;

        if (params.notify_browser) {    
            new Notification("Upload failed", {
                body: `Upload failed: ${e.status} - ${e.statusText}`,
                icon: 'gose-logo.svg',
            });
        }
    } finally {
        uploadInProgress = false;
    }
}

function showDropZone(ev: DragEvent) {
    dropZone.style.display = "block";
}

function hideDropZone() {
    dropZone.style.display = "none";
}

function canDrop(ev: DragEvent) {
    if (uploadInProgress) {
        return false;
    }

    if (ev.dataTransfer.items.length !== 1) {
        return false;
    }

    if (!ev.dataTransfer.types.includes("Files")) {
        return false;
    }

    return true;
}

function allowDrag(ev: DragEvent) {
    if (!canDrop(ev)) {
        return;
    }

    ev.preventDefault();
    ev.dataTransfer.dropEffect = "copy";
}

function handleDrop(ev: DragEvent) {
    if (!canDrop(ev)) {
        return;
    }

    ev.preventDefault();
    hideDropZone();

    let inputElm = document.getElementById("file") as HTMLInputElement;
    inputElm.files = ev.dataTransfer.files;
    inputElm.dispatchEvent(new Event("change"));
}

async function fileChanged(ev: Event) {
    ev.preventDefault();

    let tgt = ev.target as HTMLInputElement;
    if (tgt === null || tgt.files === null)
        return

    await startUpload(tgt.files);
}

function updateExpiration(server: Server) {
    let divExpiration = document.getElementById("config-expiration");
    let selExpirationClasses = document.getElementById("expiration") as HTMLSelectElement;

    selExpirationClasses.innerHTML = "";

    for (let cls of server.expiration) {
        var opt = document.createElement('option');
        opt.value = cls.id;
        opt.innerHTML = cls.title;
        selExpirationClasses.appendChild(opt);
    }

    if (server.expiration.length > 1) {
        divExpiration.classList.remove('d-none');
    } else {
        divExpiration.classList.add('d-none');
    }
}

function getUploadParams(): UploadParams {
    let selServers = document.getElementById("servers") as HTMLSelectElement;
    let selExpiration = document.getElementById("expiration") as HTMLSelectElement;
    let cbShortenLink = document.getElementById("shorten-link") as HTMLInputElement;
    let cbNotifyBrowser = document.getElementById("notify-browser") as HTMLInputElement;
    let cbNotifyMail = document.getElementById("notify-mail") as HTMLInputElement;
    let inpNotifyMail = document.getElementById("notify-mail-address") as HTMLInputElement;

    let params = new UploadParams();
    params.shorten_link = cbShortenLink.checked;
    params.server = selServers.value;
    params.notify_browser = cbNotifyBrowser.checked;

    if (selExpiration.value != "") {
        params.expiration = selExpiration.value;
    }

    if (cbNotifyMail.checked) {
        params.notify_mail = inpNotifyMail.value;
    }

    return params;
}

function onConfig(config: Config) {
    let selServers = document.getElementById("servers") as HTMLSelectElement;
    let divServers = document.getElementById('config-servers');

    for (let svr of config.servers) {
        var opt = document.createElement('option');
        opt.value = svr.id;
        opt.innerHTML = svr.title;
        selServers.appendChild(opt);
    }

    if (config.servers.length > 1) {
        divServers.classList.remove('d-none');
    }

    selServers.addEventListener('change', (ev) => {
        for (let svr of config.servers)  {
            let opt = ev.target as HTMLOptionElement;
            if (svr.id == opt.value) {
                updateExpiration(svr);
            }
        }
    });
    updateExpiration(config.servers[0]);

    if (config.features.shorten_link) {
        let divShorten = document.getElementById("config-shorten");
        divShorten.classList.remove("d-none");
    }

    if (config.features.notify_mail) {
        let divNotifyMail = document.getElementById("config-notify-mail");
        divNotifyMail.classList.remove("d-none");
    }

    if (config.features.notify_browser && 'Notification' in window) {
        let divNotifyBrowser = document.getElementById("config-notify-browser");
        divNotifyBrowser.classList.remove("d-none");
    }
}

async function setupNotification(ev: Event) {
    let cb = ev.target as HTMLInputElement;
    if (cb.checked && Notification.permission !== "granted") {
        cb.checked = false;
        let res = await Notification.requestPermission();
        if (res === "granted") {
            cb.checked = true;
        } else if (res === "denied") {
            cb.disabled = true;
        }
    }
}

export async function load() {
    statsTransferred = document.getElementById("stats-transferred");
    statsElapsed = document.getElementById("stats-elapsed");
    statsEta = document.getElementById("stats-eta");
    statsSpeed = document.getElementById("stats-speed");
    statsParts = document.getElementById("stats-parts");
    dropZone = document.getElementById("dropzone");

    let progressElm = document.getElementById("progress") as HTMLProgressElement;
    progressBar = new ProgressBar(progressElm);

    let inputElm = document.getElementById("file") as HTMLInputElement;
    inputElm.addEventListener("change", fileChanged);

    window.addEventListener("dragenter", showDropZone);
    dropZone.addEventListener("dragenter", allowDrag);
    dropZone.addEventListener("dragover", allowDrag);
    dropZone.addEventListener("drop", handleDrop);
    dropZone.addEventListener("dragleave", hideDropZone);

    // Toggle notification mail
    let swNotifyMail = document.getElementById("notify-mail");
    let divNotifyMailAddress = document.getElementById("config-notify-mail-address");
    swNotifyMail.addEventListener("change", (ev) => {
        let cb = ev.target as HTMLInputElement;
        if (cb.checked) {
            divNotifyMailAddress.classList.remove("d-none");
        } else {
            divNotifyMailAddress.classList.add("d-none");
        }
    })

    // Enable desktop notifications
    if ('Notification' in window) {
        let swNotifyBrowser = document.getElementById("notify-browser") as HTMLInputElement;
        
        if (Notification.permission == "granted") {
            swNotifyBrowser.checked = true;
        } else if (Notification.permission == "denied") {
            swNotifyBrowser.disabled = true;
        }
        swNotifyBrowser.addEventListener("change", setupNotification);
    }

    config = await apiRequest("config", {}, "GET");

    onConfig(config as Config);
}

window.addEventListener("load", load);
