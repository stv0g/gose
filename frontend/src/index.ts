import "bootstrap";
import { Tooltip } from "bootstrap";

import "../css/index.scss";

import '@fortawesome/fontawesome-free/js/fontawesome';
import '@fortawesome/fontawesome-free/js/solid';

import prettyBytes  from "pretty-bytes";
import * as prettyMilliseconds from "pretty-ms";

import { ProgressBar } from "./progress-bar";
import { Upload, UploadParams } from "./upload";
import { apiRequest } from "./api";
import { Config, Server } from "./config";
import { Chart } from "./chart";
import { Dropzone } from "./dropzone";

var progressBar: ProgressBar;
var config: Config;
var chart: Chart;
var upload: Upload | null;
let points: Array<number[]> = []

function reset() {
    if (upload && upload.inProgress) {
        upload.abort();
    }
    else {
        resetView();
    }
}

function resetView() {
    let divStats = document.getElementById("statistics");
    divStats.classList.add("d-none");
        
    let btnReset = document.getElementById("reset");
    btnReset.classList.add("d-none");
        
    let divResult = document.getElementById("result");
    divResult.classList.add("d-none");
}

function alert(cls: string, msg: string, url?: string, icon?: string) {
    let elm = document.getElementById("result");

    elm.classList.remove("alert-danger", "alert-success", "alert-warning", "d-none");
    elm.classList.add("alert-" + cls);

    elm.innerHTML = "";

    if (icon) {
        if (icon === "spinner") {
            elm.innerHTML += `<div class="me-1 spinner-border text-warning" role="status">
                <span class="visually-hidden">Loading...</span>
            </div>`;
        }
        else {
            elm.innerHTML += `<i class="me-1 fa-solid fa-${icon}"></i>`;
        }
    }

    elm.innerHTML += `<span>${msg}</span>`;
    
    if (url) {
        elm.innerHTML += `<a class="alert-link ms-auto" id="copy" data-bs-toggle="tooltip" data-bs-placement="top" title="Copy to clipboard"><i class="fa-solid fa-copy"></i></a>`;
        elm.innerHTML += `<a class="alert-link" id="upload-url" href="${url}">${url}</a>`;

        // Setup copy to clipboard
        let btnCopy = document.getElementById("copy");
        let spanUrl = document.getElementById("upload-url");
        let tooltip = new Tooltip(btnCopy);

        btnCopy.addEventListener("click", async (ev: Event) => {
            await navigator.clipboard.writeText(spanUrl.innerText);

            tooltip.dispose();
            btnCopy.title = "Copied! 🥳";
            tooltip = new Tooltip(btnCopy);
            tooltip.show();

            window.setTimeout(() => {
                tooltip.dispose();
                btnCopy.title = "Copy to clipboard";
                tooltip = new Tooltip(btnCopy);
            }, 1000)
        });
    }
}

function uploadStarted(upload: Upload) {
    let msg: string = upload.stage == "hashing"
        ? "Hashing in progress"
        : "Uploading in progress";

    alert("warning", msg, upload.url, "spinner");

    let p = upload.progress;

    let divStats = document.getElementById("statistics");
    divStats.classList.remove("d-none");

    let btnReset = document.getElementById("reset");
    btnReset.classList.remove("d-none");

    progressBar.setMinMax(0, upload.progress.totalSize);
    progressBar.set(0);

    let statsTotalBytes = document.getElementById("stats-total-bytes");
    let statsTotalParts = document.getElementById("stats-total-parts");

    statsTotalBytes.textContent = prettyBytes(p.totalSize);
    statsTotalParts.textContent = p.totalParts.toString();

    points = [];
}

function uploadEnded(upload: Upload) {
    let p = upload.progress;

    let statsBytes = document.getElementById(`stats-${upload.stage}-bytes`);
    let statsTime = document.getElementById(`stats-${upload.stage}-time`);
    let statsTimeETA = document.getElementById(`stats-${upload.stage}-eta`);

    statsBytes.textContent = prettyBytes(p.totalTransferred);
    statsTime.textContent = prettyMilliseconds(p.totalElapsed, { compact: true });
    statsTimeETA.textContent = '0 s';

    progressBar.set(p.totalSize);
}

function uploadProgressed(upload: Upload) {
    let p = upload.progress;

    let statsBytes = document.getElementById(`stats-${upload.stage}-bytes`);
    let statsTime = document.getElementById(`stats-${upload.stage}-time`);
    let statsTimeETA = document.getElementById(`stats-${upload.stage}-eta`);
    let statsSpeed = document.getElementById(`stats-${upload.stage}-speed`);
    let statsParts = document.getElementById(`stats-${upload.stage}-parts`);
    let statsTotalTime = document.getElementById(`stats-total-time`);

    statsBytes.textContent = prettyBytes(p.transferred + p.totalTransferred);
    statsTime.textContent = prettyMilliseconds(p.elapsed + p.totalElapsed);
    statsTotalTime.textContent = prettyMilliseconds(p.elapsed + p.totalElapsed + p.overallElapsed);
    
    if (Number.isFinite(p.eta)) {
        statsTimeETA.textContent = prettyMilliseconds(p.eta);
    }

    statsSpeed.textContent = prettyBytes(p.averageSpeed, { bits: true }) + "/s";
    statsParts.textContent = p.part.toString();

    progressBar.set(p.transferred + p.totalTransferred + p.totalSkipped);

    points.push([points.length, p.currentSpeed]);
    chart.render(points);
}

async function startUpload(files: FileList) {
    let params = getUploadParams();

    try {
        if (files.length === 0) {
            throw "There are now files to upload";
        }
        else if (files.length > 1) {
            throw "Can only upload a single file";
        }

        upload = new Upload(files[0], {
            start: uploadStarted,
            end: uploadEnded,
            progress: uploadProgressed,
        }, params);
        
        let url = await upload.start();

        alert("success", "Upload completed", url, "circle-check");

        if (params.notify_browser) {
            let dur = prettyMilliseconds(upload.progress.totalElapsed, { compact: true });
            let size = prettyBytes(upload.progress.totalSize);
    
            new Notification("Upload completed", {
                body: `Upload of ${size} for ${upload.file.name} has been completed in ${dur}.`,
                icon: "/img/gose-logo.png",
                renotify: true,
                tag: upload.etag
            });
        }
    }
    catch (e) {
        if (e === "Aborted") {
            resetView();
        } else {
            alert("danger", `Upload failed: ${e}`, null, "triangle-exclamation");

            if (params.notify_browser) {    
                new Notification("Upload failed", {
                    body: `Upload failed: ${e}`,
                    icon: "/img/gose-logo.png",
                });
            }
        }
    }
}

function canDrop(ev: DragEvent) {
    if (upload && upload.inProgress) {
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

function handleDrop(ev: DragEvent) {
    ev.preventDefault();
    this.hideDropZone();

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
        var opt = document.createElement("option");
        opt.value = cls.id;
        opt.innerHTML = cls.title;
        selExpirationClasses.appendChild(opt);
    }

    if (server.expiration.length > 1) {
        divExpiration.classList.remove("d-none");
    } else {
        divExpiration.classList.add("d-none");
    }
}

function getUploadParams(): UploadParams {
    let selServers = document.getElementById("servers") as HTMLSelectElement;
    let selExpiration = document.getElementById("expiration") as HTMLSelectElement;
    let cbShortURL = document.getElementById("shorten-link") as HTMLInputElement;
    let cbNotifyBrowser = document.getElementById("notify-browser") as HTMLInputElement;
    let cbNotifyMail = document.getElementById("notify-mail") as HTMLInputElement;
    let inpNotifyMail = document.getElementById("notify-mail-address") as HTMLInputElement;

    let params = new UploadParams();
    params.short_url = cbShortURL.checked;
    params.server = selServers.value;
    params.notify_browser = cbNotifyBrowser.checked;

    if (selExpiration.value !== "") {
        params.expiration = selExpiration.value;
    }

    if (cbNotifyMail.checked) {
        params.notify_mail = inpNotifyMail.value;
    }

    return params;
}

function onConfig(config: Config) {
    let selServers = document.getElementById("servers") as HTMLSelectElement;
    let divServers = document.getElementById("config-servers");

    for (let svr of config.servers) {
        var opt = document.createElement("option");
        opt.value = svr.id;
        opt.innerHTML = svr.title;
        selServers.appendChild(opt);
    }

    if (config.servers.length > 1) {
        divServers.classList.remove("d-none");
    }

    selServers.addEventListener("change", (ev) => {
        for (let svr of config.servers)  {
            let opt = ev.target as HTMLOptionElement;
            if (svr.id === opt.value) {
                updateExpiration(svr);
            }
        }
    });
    updateExpiration(config.servers[0]);

    // Update settings pane
    if (config.features.short_url) {
        let divShorten = document.getElementById("config-shorten");
        divShorten.classList.remove("d-none");
    }

    if (config.features.notify_mail) {
        let divNotifyMail = document.getElementById("config-notify-mail");
        divNotifyMail.classList.remove("d-none");
    }

    if (config.features.notify_browser && "Notification" in window) {
        let divNotifyBrowser = document.getElementById("config-notify-browser");
        divNotifyBrowser.classList.remove("d-none");
    }

    // Update footer
    let spanVersion = document.getElementById("version");
    spanVersion.innerHTML = `<a class="mx-1" href="https://github.com/stv0g/gose/commit/${config.build.commit}">v${config.build.version}</a>`;
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

async function load() {
    const btnReset = document.getElementById("reset");
    btnReset.addEventListener("click", reset);
    
    const divDropzone = document.getElementById("dropzone") as HTMLDivElement;
    new Dropzone(divDropzone, canDrop, handleDrop);

    const divChart = document.getElementById("chart") as HTMLDivElement;
    chart = new Chart(divChart);

    const progressElm = document.getElementById("progress") as HTMLProgressElement;
    progressBar = new ProgressBar(progressElm);

    const inputElm = document.getElementById("file") as HTMLInputElement;
    inputElm.addEventListener("change", fileChanged);

    // Toggle notification mail
    const swNotifyMail = document.getElementById("notify-mail");
    const divNotifyMailAddress = document.getElementById("config-notify-mail-address");
    swNotifyMail.addEventListener("change", (ev) => {
        let cb = ev.target as HTMLInputElement;
        if (cb.checked) {
            divNotifyMailAddress.classList.remove("d-none");
        } else {
            divNotifyMailAddress.classList.add("d-none");
        }
    });

    // Enable desktop notifications
    if ("Notification" in window) {
        let swNotifyBrowser = document.getElementById("notify-browser") as HTMLInputElement;
        
        if (Notification.permission === "granted") {
            swNotifyBrowser.checked = true;
        } else if (Notification.permission === "denied") {
            swNotifyBrowser.disabled = true;
        }
        swNotifyBrowser.addEventListener("change", setupNotification);
    }

    config = await apiRequest("config", {}, "GET");
    onConfig(config);
}

window.addEventListener("load", load);
