class Callbacks {
    [key: string]: any
}

export class ProgressStream extends TransformStream {
    constructor(cbs: Callbacks, total: number) {
        super({
            start() {
                this.started = Date.now();
                this.running = 0;
                this.total = total;
                this.transferred = 0;
                this.speed = 0;
                this.eta = Infinity;

                cbs.start(this);
                cbs.progress(this);
            }, // required.
            async transform(chunk, controller) {
                controller.enqueue(await chunk);

                this.transferred += chunk.length;
                this.elapsed = Date.now() - this.started;
                this.speed = this.transferred / (this.elapsed / 1e3); // elapsed is expressed in mili-seconds
                this.eta = (this.total - this.transferred) / this.speed;

                cbs.progress(this);
            },
            flush() {
                this.elapsed = Date.now() - this.started;

                cbs.finish(this);
            }
        });
    }
}
