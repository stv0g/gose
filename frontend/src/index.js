import 'bootstrap'

import '../css/index.scss'

import { ProgressBar } from './progress-bar.js';
import { Upload } from './upload.js';
import prettyBytes from 'pretty-bytes';
import prettyMilliseconds from 'pretty-ms';

import { sha256sum } from './checksum.js';

var inputElm = null;
var progressElm = null;
var resultElm = null;
var statsTransferred = null;
var statsElapsed = null;
var statsEta = null;
var statsSpeed = null;
var statsParts = null;
var progressBar = null;
var dropZone = null;
var uploadInProgress = false;

async function startUpload(files) {
    try {
        uploadInProgress = true;

        if (files.length == 0)
            return;
        else if (files.length > 1)
            throw {
                status: 400,
                statusText: 'Can only upload a single file'
            }

        let file = files[0];

        let ab = await file.arrayBuffer();

        file.checksum = await sha256sum(ab);

        let upload = new Upload({
            start: uploadStarted,
            end: uploadEnded,
            progress: uploadProgressed,
        })

        let url = await upload.upload(file);

        console.log('Upload succeeded', url);

        result.classList.remove('alert-danger', 'alert-warning');
        result.classList.add('alert-success');
        result.innerHTML = `Upload complete: <a href="${url}">${url}</a>`;

    } catch (e) {
        console.log('Upload failed', e);

        result.classList.remove('alert-success', 'alert-warning');
        result.classList.add('alert-danger');
        result.innerHTML = `${e.status} - ${e.statusText}`;
    } finally {
        uploadInProgress = false;
    }
}

function uploadStarted(upload) {
    let p = upload.progress;

    result.classList.remove('alert-danger', 'alert-success');
    result.classList.add('alert-warning');
    result.innerHTML = `Upload in progress: <a href="${upload.url}">${upload.url}</a>`;

    progressBar.setMinMax(0, p.totalSize);
    progressBar.set(0);
}

function uploadEnded(upload) {
    let p = upload.progress;

    statsTransferred.innerHTML = prettyBytes(p.totalTransferred);

    progressBar.set(p.totalSize);
}

function uploadProgressed(upload) {
    let p = upload.progress;

    progressBar.set(p.transferred + upload.progress.totalTransferred);

    statsTransferred.innerHTML = prettyBytes(p.transferred + p.totalTransferred) + ' / ' + prettyBytes(p.totalSize);
    statsElapsed.innerHTML = prettyMilliseconds(p.elapsed + p.totalElapsed, { compact: true });
    statsEta.innerHTML = prettyMilliseconds(p.eta, { compact: true });
    statsSpeed.innerHTML = prettyBytes(p.speed, { bits: true }) + '/s';
    statsParts.innerHTML = `${p.part} / ${p.totalParts}`;
}

function showDropZone(ev) {
    dropZone.style.display = 'block';
}

function hideDropZone() {
    dropZone.style.display = 'none';
}

function allowDrag(ev) {
    if (!canDrop(ev))
        return;

    ev.preventDefault();
    ev.dataTransfer.dropEffect = 'copy';
}

function handleDrop(ev) {
    if (!canDrop(ev))
        return;

    ev.preventDefault();
    hideDropZone();

    inputElm.files = ev.dataTransfer.files;
    inputElm.dispatchEvent(new Event('change'));
}

function canDrop(ev) {
    if (uploadInProgress)
        return false;

    if (ev.dataTransfer.items.length != 1)
        return false;

    if (!ev.dataTransfer.types.includes('Files'))
        return false;

    return true;
}

async function fileChanged(ev) {
    ev.preventDefault();

    await startUpload(ev.target.files);
}

export async function load() {
    inputElm = document.getElementById('file');
    progressElm = document.getElementById('progress');
    resultElm = document.getElementById('result');
    statsTransferred = document.getElementById('stats-transferred');
    statsElapsed = document.getElementById('stats-elapsed');
    statsEta = document.getElementById('stats-eta');
    statsSpeed = document.getElementById('stats-speed');
    statsParts = document.getElementById('stats-parts');
    dropZone = document.getElementById('dropzone');

    progressBar = new ProgressBar(progressElm);

    inputElm.addEventListener('change', fileChanged);

    window.addEventListener('dragenter', showDropZone);
    dropZone.addEventListener('dragenter', allowDrag);
    dropZone.addEventListener('dragover', allowDrag);
    dropZone.addEventListener('drop', handleDrop);
    dropZone.addEventListener('dragleave', hideDropZone);
}

window.addEventListener('load', load)