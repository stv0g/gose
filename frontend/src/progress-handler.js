export class ProgressHandler {
    constructor(cb, totalSize, totalParts) {
        this.callbacks = cb;
        this.totalSize = totalSize;
        this.totalParts = totalParts;
    }

    loadStart(ev) {
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

    loadEnd(ev) {
        this.callbacks.end(this);

        this.totalElapsed = Date.now() - this.started;
    }

    partLoadStart(ev) {
        this.partStarted = Date.now();

        this.part++;

        this.elapsed = 0;
        this.transferred = 0;
    }

    partLoadEnd(ev) {
        this.totalTransferred += this.transferred;
        this.totalElapsed += this.elapsed;
    }

    partProgress(ev) {
        this.total = ev.total;
        this.transferred = ev.loaded;

        this.elapsed = Date.now() - this.partStarted;

        this.update()

        this.callbacks.progress(this);
    }

    update() {
        let transferred = this.totalTransferred + this.transferred;
        let elapsed = this.totalElapsed + this.elapsed;

        this.speed = 8e3 * transferred / elapsed;
        this.eta = (this.totalSize - transferred) / this.speed;
    }
}