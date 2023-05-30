// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

class Callbacks {
    [key: string]: any
}

export class ProgressHandler {
    callbacks: Callbacks;

    averageSpeed: number;
    currentSpeed: number;

    part: number;
    eta: number;
    started: number;
    elapsed: number;
    transferred: number;

    overallElapsed: number;

    totalSize: number;
    totalParts: number;
    totalElapsed: number;
    totalTransferred: number;
    totalSkipped: number;

    partStarted: number;
    lastProgress: number;

    constructor(cbs: Callbacks, totalSize: number, totalParts: number) {
        this.callbacks = cbs;
        this.totalSize = totalSize;
        this.totalParts = totalParts;

        this.overallElapsed = 0;

        this.totalElapsed = 0;
        this.totalTransferred = 0;
        this.totalSkipped = 0;
    }

    start() {
        this.averageSpeed = 0;
        this.currentSpeed = 0;

        this.part = 0;
        this.eta = 0;
        this.started = Date.now();
        this.elapsed = 0;
        this.transferred = 0;

        this.totalElapsed = 0;
        this.totalTransferred = 0;
        this.totalSkipped = 0;

        this.callbacks.start(this);
        this.callbacks.progress(this);
    }

    end() {
        this.callbacks.end(this);

        this.totalElapsed = Date.now() - this.started;
        this.overallElapsed += this.totalElapsed;
    }

    partStart(ev: ProgressEvent) {
        this.partStarted = Date.now();

        this.part++;

        this.elapsed = 0;
        this.transferred = 0;
    }

    partEnd(ev: ProgressEvent) {
        this.totalTransferred += this.transferred;
        this.totalElapsed += this.elapsed;
    }

    partProgress(ev: ProgressEvent) {    
        let incrElapsed = Date.now() - this.partStarted - this.elapsed;
        let incrTransferred = ev.loaded - this.transferred;

        this.elapsed += incrElapsed;
        this.transferred += incrTransferred;

        let transferred = this.totalTransferred + this.transferred;
        let elapsed = this.totalElapsed + this.elapsed;

        if (incrElapsed > 0) {
            this.currentSpeed = 8e3 * incrTransferred / incrElapsed; // b/s
        }

        if (elapsed > 0) {
            this.averageSpeed = 8e3 * transferred / elapsed; // b/s
        }
        
        this.eta = 8e3 * (this.totalSize - this.totalSkipped - transferred) / this.averageSpeed;
        
        this.callbacks.progress(this);
    }

    partSkip(size: number) {
        this.part++;
        this.totalSkipped += size;
    }
}
