class Callbacks {
    [key: string]: any
}

export class ProgressHandler {
    callbacks: Callbacks;
    totalSize: number;
    totalParts: number;
    part: number;
    eta: number;
    speed: number;
    started: number;
    elapsed: number;
    total: number;
    transferred: number;
    totalElapsed: number;
    totalTransferred: number;
    partStarted: number;

    constructor(cbs: Callbacks, totalSize: number, totalParts: number) {
        this.callbacks = cbs;
        this.totalSize = totalSize;
        this.totalParts = totalParts;
    }

    loadStart() {
        this.started = Date.now();

        this.part = 0;
        this.speed = 0;
        this.eta = 0;

        this.elapsed = 0;
        this.transferred = 0;

        this.totalElapsed = 0;
        this.totalTransferred = 0;

        this.callbacks.start(this);
        this.callbacks.progress(this);
    }

    loadEnd() {
        this.callbacks.end(this);

        this.totalElapsed = Date.now() - this.started;
    }

    partLoadStart(ev: ProgressEvent) {
        this.partStarted = Date.now();

        this.part++;

        this.elapsed = 0;
        this.transferred = 0;
    }

    partLoadEnd(ev: ProgressEvent) {
        this.totalTransferred += this.transferred;
        this.totalElapsed += this.elapsed;
    }

    partProgress(ev: ProgressEvent) {
        this.total = ev.total;
        this.transferred = ev.loaded;

        this.elapsed = Date.now() - this.partStarted;

        this.update();

        this.callbacks.progress(this);
    }

    update() {
        let transferred = this.totalTransferred + this.transferred;
        let elapsed = this.totalElapsed + this.elapsed;

        this.speed = 8e3 * transferred / elapsed;
        this.eta = (this.totalSize - transferred) / this.speed;
    }
}
